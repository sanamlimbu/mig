BEGIN;

DROP TABLE IF EXISTS friendships;
DROP TABLE IF EXISTS group_users;
DROP TABLE IF EXISTS groups;
DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS users__workflow_state;
DROP TYPE IF EXISTS groups__workflow_state;
DROP TYPE IF EXISTS groups__type;
DROP TYPE IF EXISTS friendships__workflow_state;
DROP TYPE IF EXISTS group_users__workflow_state;

COMMIT;