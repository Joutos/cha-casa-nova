-- Convidados: uma linha por e-mail. Presença, intenção de contribuir e
-- mensagens ficam vinculadas a esta pessoa.
CREATE TABLE guests (
    id                     TEXT PRIMARY KEY,
    name                   TEXT NOT NULL CHECK (length(name) BETWEEN 2 AND 100),
    email                  TEXT NOT NULL COLLATE NOCASE UNIQUE CHECK (length(email) <= 254),
    attending              INTEGER NOT NULL DEFAULT 0 CHECK (attending IN (0, 1)),
    wants_to_contribute    INTEGER NOT NULL DEFAULT 0 CHECK (wants_to_contribute IN (0, 1)),
    invite_token           TEXT UNIQUE,
    attending_confirmed_at DATETIME,
    created_at             DATETIME NOT NULL,
    updated_at             DATETIME NOT NULL
);
CREATE INDEX idx_guests_created_at ON guests (created_at);
CREATE INDEX idx_guests_name ON guests (name COLLATE NOCASE);

CREATE TABLE messages (
    id         TEXT PRIMARY KEY,
    guest_id   TEXT NOT NULL REFERENCES guests (id) ON DELETE CASCADE,
    body       TEXT NOT NULL CHECK (length(body) BETWEEN 1 AND 1000),
    created_at DATETIME NOT NULL
);
CREATE INDEX idx_messages_guest_id ON messages (guest_id, created_at);

CREATE TABLE admins (
    id            TEXT PRIMARY KEY,
    name          TEXT NOT NULL,
    email         TEXT NOT NULL COLLATE NOCASE UNIQUE,
    password_hash TEXT NOT NULL,
    created_at    DATETIME NOT NULL,
    updated_at    DATETIME NOT NULL
);

CREATE TABLE sessions (
    id           TEXT PRIMARY KEY,
    admin_id     TEXT NOT NULL REFERENCES admins (id) ON DELETE CASCADE,
    token_hash   TEXT NOT NULL UNIQUE,
    expires_at   DATETIME NOT NULL,
    created_at   DATETIME NOT NULL,
    last_seen_at DATETIME NOT NULL
);
CREATE INDEX idx_sessions_admin_id ON sessions (admin_id);
CREATE INDEX idx_sessions_expires_at ON sessions (expires_at);
