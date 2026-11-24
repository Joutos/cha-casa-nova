# API — `/api/v1`

Todas as respostas são JSON (exceto o CSV). Erros seguem um formato único:

```json
{ "error": { "code": "validation_error", "message": "Alguns campos precisam de atenção.", "fields": { "email": "Confira o e-mail: ele parece incompleto." } } }
```

| Código HTTP | `code` | Quando |
|---|---|---|
| 400 | `invalid_json`, `empty_patch` | Corpo ilegível, campo desconhecido, PATCH vazio |
| 401 | `unauthorized` | Sem sessão, sessão expirada ou login inválido |
| 403 | `forbidden` | CSRF ausente/inválido ou Origin de outro site |
| 404 | `not_found` | Recurso ou rota inexistente |
| 409 | `already_confirmed`, `message_limit`, `admin_exists` | Conflitos de regra de negócio |
| 413 | `payload_too_large` | Corpo acima de 16 KB |
| 422 | `validation_error` | Campos inválidos (ver `fields`) |
| 429 | `rate_limited` | Muitas tentativas (cabeçalho `Retry-After`) |
| 500 | `internal_error` | Erro inesperado (detalhe só no log) |

## Públicos

### `GET /health`
`200 {"status":"ok"}` ou `503 {"status":"unavailable"}` se o banco não responder.

### `POST /registrations` — confirmar presença
```json
{ "name": "Ana Lima", "email": "ana@exemplo.com", "contribute": true, "website": "" }
```
- `name`: 2–100 caracteres, sem `<`/`>`. `email`: endereço simples, até 254. `contribute`: marcou "Quero contribuir". `website`: honeypot, deve ir vazio.
- `201 {"message": "..."}`; `409 already_confirmed` se o e-mail já confirmou; `422` com `fields`.
- Rate limit: `PUBLIC_RATE_PER_MINUTE` por IP.
- Efeito colateral: se o SMTP estiver configurado, envia o convite com o link pessoal.

### `POST /messages` — deixar mensagem
```json
{ "name": "Tia Lúcia", "email": "lucia@exemplo.com", "message": "Felicidades!", "contribute": false, "website": "" }
```
- `message`: 1–1000 caracteres. Cria o convidado se o e-mail for novo.
- `201`; `409 message_limit` após `MAX_MESSAGES_PER_GUEST` mensagens.

### `GET /invites/:token` — dados do convite
Token de 43 caracteres enviado por e-mail. `200`:
```json
{ "guest_name": "Ana Lima", "date_label": "23 de Janeiro de 2027", "time": "11h00", "address_line1": "Av Saranhema, 1727", "address_line2": "Guarani, Belo Horizonte - MG", "maps_query": "Av Saranhema 1727 ...", "signature": "Sofia e João" }
```
`404` para token inválido ou de quem não confirmou presença.

## Autenticação

### `POST /auth/login`
`{"email": "...", "password": "..."}` → `200` com cookie de sessão:
```json
{ "admin": { "id": "…", "name": "Sofia", "email": "sofia@…" }, "csrf_token": "…", "expires_at": "2026-11-09T05:00:00Z" }
```
`401` com a mesma mensagem para e-mail inexistente ou senha errada. Rate limit: `LOGIN_RATE_PER_MINUTE` por IP (e 10/min no Nginx).

### `GET /auth/me`
Requer sessão. `200 {"admin": {...}, "csrf_token": "..."}`.

### `POST /auth/logout`
Requer sessão + `X-CSRF-Token`. `204`; apaga a sessão no servidor e o cookie.

## Administrativos

Todos exigem o cookie de sessão. `PATCH`/`DELETE` exigem também `X-CSRF-Token`.

### `GET /admin/stats`
`{"total": 12, "attending": 9, "contribute": 4, "with_message": 5}`

### `GET /admin/registrations`
| Parâmetro | Valores | Padrão |
|---|---|---|
| `q` | busca em nome, e-mail e texto das mensagens (até 100 caracteres; `%` e `_` são literais) | — |
| `type` | `attending`, `contribute`, `message` | todos |
| `sort` | `created_at`, `name`, `email` | `created_at` |
| `order` | `asc`, `desc` | `desc` |
| `page` | ≥ 1 | 1 |
| `page_size` | 1–100 | 20 |

```json
{ "items": [ { "id": "…", "name": "Ana Lima", "email": "ana@…", "attending": true, "wants_to_contribute": true, "message_count": 1, "created_at": "…", "updated_at": "…" } ], "total": 1, "page": 1, "page_size": 20 }
```
Valores fora da lista → `422`.

### `GET /admin/registrations/:id`
Convidado com `attending_confirmed_at`, `messages[]` (em ordem cronológica) e `invite_url` (se confirmou presença).

### `PATCH /admin/registrations/:id`
Qualquer combinação de `{"name": "…", "attending": true, "wants_to_contribute": false}`. Devolve o detalhe atualizado. Marcar presença gera o token do convite se ainda não existir.

### `DELETE /admin/registrations/:id`
`204`. Apaga também as mensagens (cascata).

### `GET /admin/registrations/export`
Mesmos filtros da listagem (sem paginação). CSV UTF-8 com BOM, separador `;` (abre direto no Excel em português), colunas `Nome; E-mail; Presença; Contribuição; Mensagens; Inscrito em`. Células que começam com `= + - @` recebem `'` para evitar injeção de fórmulas.
