BEGIN;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE EXTENSION IF NOT EXISTS "pg_trgm";

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TYPE users__workflow_state AS ENUM (
    'active',
    'suspended',
    'deleted'
);

CREATE TABLE users (
    id                  BIGINT PRIMARY KEY NOT NULL,
    uuid                uuid UNIQUE NOT NULL, -- Supabase auth user id
    username            VARCHAR(255) UNIQUE NOT NULL,
    workflow_state      users__workflow_state NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS users_username_trgm_idx ON users USING gin (username  gin_trgm_ops);

CREATE TYPE chatrooms__workflow_state AS ENUM (
    'active',
    'deleted'
);

CREATE TYPE chatrooms__type AS ENUM (
    'private',          -- users need approval to join          
    'public'            -- users can join without approval
);

CREATE TABLE chatrooms (
    id                  BIGINT PRIMARY KEY NOT NULL,
    name                VARCHAR(255) UNIQUE NOT NULL,
    workflow_state      chatrooms__workflow_state NOT NULL,
    type                chatrooms__type NOT NULL,
    created_by          BIGINT NOT NULL REFERENCES users (id),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS chatrooms_name_trgm_idx ON chatrooms USING gin (name  gin_trgm_ops);

CREATE TYPE friendships__workflow_state AS ENUM (
    'pending',
    'active',
    'rejected',
    'cancelled'
);

CREATE TABLE friendships(
    id                      BIGINT PRIMARY KEY NOT NULL,
    requester_id            BIGINT NOT NULL REFERENCES users (id),
    user_id                 BIGINT NOT NULL REFERENCES users (id),    
    workflow_state          friendships__workflow_state NOT NULL,
    workflow_completed_by   BIGINT NOT NULL REFERENCES users (id),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at              TIMESTAMPTZ
);

COMMIT;