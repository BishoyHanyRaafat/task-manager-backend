-- Users table: stores only profile info
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,               -- UUID stored as TEXT
    first_name TEXT NOT NULL,
    last_name  TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Passwords table: stores the current password hash (1 row per user)
CREATE TABLE IF NOT EXISTS passwords (
    id TEXT PRIMARY KEY,               -- UUID stored as TEXT
    user_id TEXT NOT NULL UNIQUE,
    v TEXT NOT NULL,                   -- password hash
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- OAuth providers table: stores linked SSO provider accounts
CREATE TABLE IF NOT EXISTS auth_providers (
    id TEXT PRIMARY KEY,               -- UUID stored as TEXT
    user_id TEXT NOT NULL,
    provider TEXT NOT NULL CHECK (provider IN ('google', 'github')),
    provider_user_id TEXT NOT NULL,
    email TEXT,
    username TEXT,
    display_name TEXT,
    avatar_url TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(provider, provider_user_id),
    UNIQUE(user_id, provider),
    FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);
