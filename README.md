# Chá de Casa Nova — Sofia e João

Site-convite com confirmação de presença, contribuição via Pix e mensagens, mais um painel privado para os anfitriões. Implementado a partir do Figma (`w93Fely3wvFygnUAvkAZxc`).

- **Frontend:** React 18, TypeScript, Vite, Tailwind CSS, shadcn/ui (restilizado), React Router, TanStack Query, React Hook Form + Zod
- **Backend:** Go, Gin, GORM, SQLite com migrations versionadas
- **Infra:** Docker Compose, Nginx (SPA + proxy reverso), volume para o SQLite

Documentação: [`docs/design-analysis.md`](docs/design-analysis.md) (levantamento do Figma e decisões), [`docs/architecture.md`](docs/architecture.md), [`docs/api.md`](docs/api.md).

## Telas

| Rota | Tela do Figma |
|---|---|
| `/` | Frame 1: convite e escolha (presença, contribuição, mensagem) |
| `/presenca` → `/presenca/confirmado` | Frames 2, 4 e 3 |
| `/contribuir` | Frame 5 (Pix) |
| `/mensagem` → `/mensagem/enviada` | Frames 7 e 8 |
| `/convite/<token>` | Frame 6: convite pessoal com endereço e mapa (link enviado por e-mail) |
| `/privacidade` | Aviso de privacidade (LGPD) |
| `/admin/login`, `/admin` | Painel dos anfitriões (não existe no Figma; criado no mesmo estilo) |

O convidado marca uma ou mais opções na home e percorre as etapas na ordem presença → contribuição → mensagem.

## Rodando com Docker (recomendado)

Requisitos: Docker 24+ com Compose v2.

```bash
cp .env.example .env
# edite o .env: SESSION_SECRET (openssl rand -base64 48), ADMIN_EMAIL, ADMIN_PASSWORD,
# endereço do evento e, se quiser enviar e-mails, as variáveis SMTP_*
docker compose up -d --build
```

Acesse http://localhost:8080 e o painel em http://localhost:8080/admin.

No primeiro start, o backend aplica as migrations e cria o administrador definido em `ADMIN_EMAIL`/`ADMIN_PASSWORD` (só se ainda não houver nenhum). **Depois disso, apague `ADMIN_PASSWORD` do `.env`.**

> **Primeiro build:** o `go.sum` é gerado dentro do build (`go mod tidy`, com verificação no `sum.golang.org`). Para fixá-lo no repositório, rode uma vez `cd backend && go mod tidy` com Go 1.24+ e versione o arquivo.

### Produção

- Coloque um terminador TLS na frente (Caddy, Traefik, Cloudflare, load balancer) apontando para a porta do Nginx. Mantenha `COOKIE_SECURE=true` e descomente o `Strict-Transport-Security` em `frontend/nginx.conf`.
- Defina `PUBLIC_BASE_URL` com o domínio real: ele é usado nos links de convite.
- O backend não publica portas; só o Nginx fala com ele pela rede interna do Compose.

### Gerenciar administradores

```bash
# criar outro administrador
docker compose run --rm -e ADMIN_NAME="João" -e ADMIN_EMAIL=joao@exemplo.com -e ADMIN_PASSWORD='senha-longa-aqui' backend create-admin
# trocar senha (derruba as sessões abertas)
docker compose run --rm -e ADMIN_EMAIL=joao@exemplo.com -e ADMIN_PASSWORD='nova-senha-longa' backend reset-password
```

Senhas precisam de pelo menos 12 caracteres. Não existe cadastro público de administradores.

### Backup e persistência

O banco fica no volume `cha-de-casa-nova_app-data` e sobrevive a `docker compose down` e a rebuilds (`down -v` apaga). Para um backup consistente:

```bash
docker compose exec backend sh -c 'cp /data/app.db /tmp/backup.db' && docker compose cp backend:/tmp/backup.db ./backup-$(date +%F).db
```

## Desenvolvimento local

```bash
# backend (Go 1.24+ e gcc, por causa do driver SQLite com CGO)
cd backend
ADMIN_EMAIL=admin@example.com ADMIN_PASSWORD='troque-esta-senha-123' go run ./cmd/server

# frontend (Node 20+), em outro terminal; o Vite faz proxy de /api para :8080
cd frontend
npm install
npm run dev     # http://localhost:5173
```

## Assets visuais do Figma

A foto do casal e as ilustrações (flor e ramo) estão no Figma. O repositório traz versões provisórias das ilustrações e, sem a foto, o cabeçalho usa um fundo marrom. Para baixar os originais:

```bash
FIGMA_TOKEN=figd_seu_token node scripts/export-figma-assets.mjs
```

O token pessoal (somente leitura) é criado em Figma → Settings → Security. O script grava `hero.jpg`, `flower.svg` e `branch.svg` em `frontend/public/images/`.

## Pix

O QR Code ainda não existe; a tela mostra um espaço reservado. Quando houver:

1. Salve a imagem em `frontend/public/images/pix-qr.png`.
2. No `.env`: `VITE_PIX_QR_IMAGE=/images/pix-qr.png` e, opcionalmente, `VITE_PIX_KEY=sua-chave` (mostra a chave e um botão de copiar).
3. `docker compose up -d --build frontend`.

## E-mail do convite

Com `SMTP_HOST` e `SMTP_FROM` preenchidos, cada confirmação de presença dispara o e-mail do Frame 6 com o link pessoal `/convite/<token>` (onde fica o mapa). Portas 587 (STARTTLS) e 465 (TLS) são suportadas. Sem SMTP, nada é enviado e o painel mostra o link de cada convidado para você mandar manualmente.

## Testes

```bash
cd backend && go test -race ./...          # serviços, handlers, auth, autorização, validação, segurança, migrations
cd frontend && npm test                    # componentes, formulários, navegação, autenticação, painel (Vitest + MSW)
cd frontend && npx playwright install chromium && \
  E2E_BASE_URL=http://localhost:8080 E2E_ADMIN_EMAIL=... E2E_ADMIN_PASSWORD=... npm run test:e2e
```

Os testes E2E criam e apagam seus próprios dados; rode-os contra um ambiente de teste, não contra a lista real de convidados.

## Segurança (resumo)

Argon2id para senhas, sessão em cookie HttpOnly/Secure/SameSite=Strict, token guardado só como hash, CSRF por HMAC + checagem de Origin, rate limit no login e nos formulários (backend e Nginx), honeypot, validação no backend com JSON estrito e limite de 16 KB, consultas parametrizadas, CSP e cabeçalhos de segurança, CORS fechado por padrão, erros 5xx sem detalhes, logs sem corpo, query string, senhas ou tokens, mesma resposta para e-mail inexistente e senha errada, CSV protegido contra injeção de fórmulas, containers sem root, somente leitura e sem capabilities. Detalhes em `docs/architecture.md`.

## LGPD

Coletamos só nome, e-mail, as escolhas e a mensagem. A finalidade é informada ao lado de cada formulário e em `/privacidade`. **Revise o texto dessa página**: ele promete apagar os dados até 90 dias após a festa e indica falar com os anfitriões para exclusão. Ajuste se a prática for outra; o painel já permite excluir qualquer inscrição.
