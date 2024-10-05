BEGIN;

DROP TABLE IF EXISTS friendships;
DROP TABLE IF EXISTS chatrooms;
DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS users__workflow_state;
DROP TYPE IF EXISTS chatrooms__workflow_state;
DROP TYPE IF EXISTS chatrooms__type;
DROP TYPE IF EXISTS friendships__workflow_state;

DROP INDEX IF EXISTS users_username_trgm_idx;
DROP INDEX IF EXISTS chatrooms_name_trgm_idx;

COMMIT;