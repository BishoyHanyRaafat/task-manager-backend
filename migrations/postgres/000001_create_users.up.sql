-- sqlfluff:dialect:postgres
-- Needed for gen_random_uuid()
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TYPE USER_TYPE AS ENUM ('standard', 'admin');

CREATE TABLE IF NOT EXISTS users
(
    id         UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    first_name TEXT        NOT NULL,
    last_name  TEXT        NOT NULL,
    email      TEXT        NOT NULL UNIQUE,
    user_type  USER_TYPE   NOT NULL DEFAULT 'standard',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TYPE auth_provider AS ENUM ('google', 'github');

-- Passwords table: stores the current password hash (1 row per user)
CREATE TABLE IF NOT EXISTS passwords
(
    id         UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL UNIQUE REFERENCES users (id) ON DELETE CASCADE,
    v          TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- OAuth providers table: stores linked SSO provider accounts
CREATE TABLE IF NOT EXISTS auth_providers
(
    id               UUID PRIMARY KEY,
    user_id          UUID          NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    provider         AUTH_PROVIDER NOT NULL,
    provider_user_id TEXT          NOT NULL,
    email            TEXT,
    username         TEXT,
    display_name     TEXT,
    avatar_url       TEXT,
    created_at       TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ   NOT NULL DEFAULT now(),
    UNIQUE (provider, provider_user_id, user_id)
);
