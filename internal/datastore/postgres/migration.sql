CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    "name" TEXT,
    email TEXT NOT NULL,
    "role" INT
);

CREATE TABLE IF NOT EXISTS todos (
    id TEXT PRIMARY KEY,
    detail TEXT NOT NULL,
    user_id TEXT NOT NULL,
    done BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS "sessions" (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    refresh_token_expired_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS sessions_refresh_token ON "sessions"(refresh_token);
