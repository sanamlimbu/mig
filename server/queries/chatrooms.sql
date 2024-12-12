-- name: GetChatroom :one
SELECT * FROM chatrooms
WHERE id = $1 LIMIT 1;

-- name: GetChatroomWithCreator :one
SELECT c.*,
  u.username creator_username,
  u.email creator_email,
  u.workflow_state creator_workflow_state
FROM chatrooms c
JOIN users u ON u.id = c.created_by 
WHERE @id = $1 LIMIT 1;

-- name: CreateChatroom :one
INSERT INTO chatrooms (
  name, workflow_state, type, created_by
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: UpdateChatroom :one
UPDATE chatrooms
  set name = $2,
  workflow_state = $3,
  type = $4
WHERE id = $1
RETURNING *;

-- name: GetChatroomCreator :one
SELECT u.* FROM chatrooms c
JOIN users u ON c.created_by = u.id
WHERE c.id = $1 LIMIT 1;

-- name: GetChatroomsByWorkflowStates :many
SELECT * FROM chatrooms
WHERE workflow_state = ANY(@workflow_states::chatroom_workflow_state[])
LIMIT @page_size
OFFSET @page;

-- name: GetChatroomsBySearchTermAndWorkflowStates :many
SELECT * FROM chatrooms
WHERE name ILIKE @search_term AND
  workflow_state = ANY(@workflow_states::chatroom_workflow_state[])
LIMIT @page_size
OFFSET @page; 

-- name: GetChatroomsByCreatorID :many
SELECT * FROM chatrooms
WHERE created_by = @id
LIMIT @page_size
OFFSET @page; 

-- name: GetChatroomMessages :many
SELECT m.* from messages m
JOIN chatrooms c ON m.chatroom_id = c.id
WHERE c.id = $1;