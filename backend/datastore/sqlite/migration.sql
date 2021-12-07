CREATE TABLE IF NOT EXISTS sessions (
    token TEXT PRIMARY KEY,
    data BLOB NOT NULL,
    expiry REAL NOT NULL
);
	
CREATE INDEX IF NOT EXISTS sessions_expiry_idx ON sessions(expiry);

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