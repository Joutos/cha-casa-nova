# Arquitetura

```
Navegador ──HTTPS──► Nginx (container frontend, :8080)
                      ├── /            → SPA React (arquivos estáticos)
                      └── /api/*       → proxy → backend Go (:8080, rede interna)
                                                   └── SQLite em volume Docker (/data/app.db)
```

Um único domínio serve a SPA e a API, então cookies e CORS ficam simples: o CORS é restritivo por padrão (lista vazia) e o cookie de sessão é `SameSite=Strict`. A API é independente e versionada em `/api/v1`, podendo ser consumida por outros clientes.

## Backend (Go)

```
cmd/server            ponto de entrada + comandos CLI (serve, migrate, create-admin, reset-password, healthcheck)
internal/config       variáveis de ambiente
internal/database     abertura do SQLite (WAL, foreign keys, busy timeout) + executor de migrations
migrations/           SQL versionado (NNNN_nome.up.sql / .down.sql), embutido no binário
internal/models       structs GORM
internal/repositories acesso a dados (apenas consultas parametrizadas; ordenação por lista fixa)
internal/services     regras de negócio, validação, Argon2id, sessões, CSRF, e-mail, CSV
internal/handlers     HTTP ↔ serviços
internal/middleware   request ID, logs, recuperação, erros, segurança, CORS, rate limit, auth, CSRF
internal/routes       montagem do roteador Gin
internal/apperr       erro de domínio → resposta JSON padronizada
```

Fluxo de uma requisição: `RequestID → Logger → Recovery → SecurityHeaders → CORS → BodyLimit(16 KB) → ErrorHandler → SameOrigin → [RateLimit] → [RequireAdmin → RequireCSRF] → handler`. Handlers registram erros com `c.Error(err)` e o `ErrorHandler` responde `{"error": {"code", "message", "fields"}}`. Erros 5xx viram uma mensagem genérica; o detalhe só vai para o log.

### Modelo de dados

```
guests ─┬─< messages        (ON DELETE CASCADE)
admins ─┴─< sessions        (ON DELETE CASCADE)
```

| Tabela | Campos principais | Observações |
|---|---|---|
| `guests` | `id` (UUID), `name`, `email` (UNIQUE, NOCASE), `attending`, `wants_to_contribute`, `invite_token` (UNIQUE), `attending_confirmed_at`, `created_at`, `updated_at` | Uma linha por pessoa. CHECKs de tamanho e de booleanos |
| `messages` | `id`, `guest_id` (FK), `body` (1–1000), `created_at` | Limite por convidado configurável |
| `admins` | `id`, `name`, `email` (UNIQUE), `password_hash` (Argon2id PHC), timestamps | Sem cadastro público |
| `sessions` | `id`, `admin_id` (FK), `token_hash` (SHA-256, UNIQUE), `expires_at`, `created_at`, `last_seen_at` | O token em si só existe no cookie |

SQLite usa uma única conexão de escrita (`SetMaxOpenConns(1)`) com `_txlock=immediate`: as transações de "buscar por e-mail e criar" ficam serializadas, sem corrida. Para o volume de um chá de casa nova isso sobra.

### Autenticação e sessão

1. `POST /auth/login` valida e-mail/senha. E-mail inexistente também roda o Argon2id (hash fictício), então tempo e mensagem são iguais aos de senha errada.
2. Gera token aleatório de 256 bits, grava só o SHA-256 e envia o token em cookie `HttpOnly`, `SameSite=Strict`, `Secure` em produção (com prefixo `__Host-`).
3. O token CSRF é `HMAC-SHA256(SESSION_SECRET, token)`: nada extra é guardado; `/auth/me` e o login o devolvem, o frontend o mantém só em memória e o envia em `X-CSRF-Token` em POST/PATCH/DELETE.
4. Sessões expiram (`SESSION_TTL`, padrão 8 h); expiradas são limpas no login e a cada 10 min. `reset-password` derruba todas as sessões do administrador.

### Proteção contra spam

Honeypot (`website`, invisível; se preenchido a API responde 201 sem gravar), rate limit por IP no backend (token bucket) e no Nginx, e-mail único, limite de mensagens por e-mail, tamanho máximo de corpo.

## Frontend (React)

```
src/pages            telas (públicas e admin/)
src/layouts          estrutura das páginas públicas e do painel
src/components       componentes de marca (HeroHeader, Section, FormField…) e ui/ (shadcn/ui restilizado)
src/contexts         RsvpFlowContext (etapas escolhidas, em sessionStorage) e AuthContext (sessão via TanStack Query)
src/services         cliente HTTP e chamadas da API
src/hooks            consultas/mutações do painel, query na URL, debounce, título
src/schemas          Zod (espelha a validação do backend)
src/routes           roteador e rota protegida
```

O painel é carregado sob demanda (code splitting), e a sessão só é consultada dentro de `/admin`. Filtros, busca, ordenação e página ficam na URL. Nenhuma credencial é armazenada no navegador: o cookie é HttpOnly e o token CSRF vive em memória.

## Decisões e limites conhecidos

- **Rate limit em memória:** correto para uma instância. Com várias réplicas, mover para Redis ou para o Nginx.
- **E-mail:** envio assíncrono com timeout de 30 s e sem fila; uma falha é registrada no log e o link continua disponível no painel.
- **Backup:** o banco é um arquivo; veja o README para copiar com segurança.
