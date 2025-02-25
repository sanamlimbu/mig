-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_cron";

CREATE TYPE user_workflow_state AS ENUM (
    'active',
    'suspended',
    'unverified',
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

CREATE TABLE last_read_messages (
    sender_id               UUID NOT NULL REFERENCES users (id),    
    recipient_id            UUID NOT NULL REFERENCES users (id),
    message_id              UUID NOT NULL REFERENCES messages (id),
    updated_at              TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (sender_id, recipient_id)
);

-- Schedule a cron job to delete expired or revoked tokens at midnight.
SELECT cron.schedule(
    'delete_expired_tokens',
    '0 0 * * *',
    $$DELETE FROM refresh_tokens WHERE expires_at <= NOW() OR revoked = TRUE$$
);

-- +goose Down
DROP TABLE IF EXISTS last_read_messages;
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

-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM cron.job WHERE jobname = 'delete_expired_tokens') THEN
        PERFORM cron.unschedule('delete_expired_tokens');
    END IF;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

DROP EXTENSION IF EXISTS "pg_cron";
DROP EXTENSION IF EXISTS "uuid-ossp";