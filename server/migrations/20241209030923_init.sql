-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE user_workflow_state AS ENUM (
    'active',
    'suspended',
    'deleted'
);

CREATE TABLE users (
    id                          UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    email                       VARCHAR(255) UNIQUE NOT NULL,
    username                    VARCHAR(255) UNIQUE NOT NULL,
    password                    TEXT NOT NULL,
    workflow_state              user_workflow_state NOT NULL,
    reset_password_url          TEXT,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at                  TIMESTAMPTZ
);

CREATE TABLE refresh_tokens (
    id                  UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),                     
    user_id             UUID NOT NULL REFERENCES users(id),
    token               TEXT NOT NULL UNIQUE,
    expires_at          TIMESTAMP NOT NULL,
    created_at          TIMESTAMP DEFAULT NOW(),
    revoked             BOOLEAN DEFAULT FALSE
);

CREATE TYPE chatroom_workflow_state AS ENUM ('active', 'deleted');

CREATE TYPE chatroom_type AS ENUM ('private', 'public');

CREATE TABLE chatrooms (
    id                  UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    name                VARCHAR(255) UNIQUE NOT NULL,
    workflow_state      chatroom_workflow_state NOT NULL,
    type                chatroom_type NOT NULL,
    created_by          UUID NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ
);

CREATE TYPE friendship_workflow_state AS ENUM (
    'pending',
    'active',
    'rejected',
    'cancelled',
    'deleted'
);

CREATE TABLE friendships (
    id                      UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    requester_id            UUID NOT NULL,
    user_id                 UUID NOT NULL,    
    workflow_state          friendship_workflow_state NOT NULL,
    workflow_completed_by   UUID NOT NULL,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at              TIMESTAMPTZ
);

CREATE TYPE message_workflow_state AS ENUM (
    'created',
    'updated',
    'deleted'
);

CREATE TYPE message_type AS ENUM (
    'private',
    'chatroom'
);

CREATE TABLE messages (
    id                      UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    sender_id               UUID NOT NULL,
    recipient_id            UUID,
    chatroom_id             UUID,   
    workflow_state          message_workflow_state NOT NULL,
    message_type            message_type NOT NULL,
    content                 TEXT NOT NULL,
    is_read                 BOOLEAN,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at              TIMESTAMPTZ
);

-- +goose Down
DROP TABLE IF EXISTS friendships;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS chatrooms;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS chatroom_workflow_state;
DROP TYPE IF EXISTS chatroom_type;
DROP TYPE IF EXISTS friendship_workflow_state;
DROP TYPE IF EXISTS message_workflow_state;
DROP TYPE IF EXISTS message_type;
DROP TYPE IF EXISTS user_workflow_state;