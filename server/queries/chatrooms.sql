-- name: GetChatroom :one
SELECT * FROM chatrooms
WHERE id = $1 LIMIT 1;

-- name: GetChatroomByName :one
SELECT * FROM chatrooms
WHERE name = $1 LIMIT 1;

-- name: GetChatroomWithCreator :one
SELECT c.*,
  u.username creator_username,
  u.email creator_email,
  u.workflow_state creator_workflow_state
FROM chatrooms c
JOIN users u ON u.id = c.created_by 
WHERE c.id = $1 LIMIT 1;

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

-- name: GetChatroomsByCreatorIDAndWorkflowStates :many
SELECT * FROM chatrooms
WHERE created_by = @id AND
  workflow_state = ANY(@workflow_states::chatroom_workflow_state[])
LIMIT @page_size
OFFSET @page; 

-- name: GetChatroomMessages :many
SELECT m.*,
  c.name chatroom_name,
  c.type chatroom_type,
  c.workflow_state chatroom_workflow_state,
  c.created_by chatroom_creator_id,
  u.email sender_email,
  u.username sender_username,
  u.workflow_state sender_workflow_state
FROM messages m
JOIN chatrooms c ON m.chatroom_id = c.id
JOIN users u ON u.id = m.sender_id
WHERE c.id = $1
ORDER BY m.created_at DESC;

-- name: CreateChatroomMessage :one
INSERT INTO messages (
 id, sender_id, chatroom_id, content, workflow_state, message_type
) VALUES (
  $1, $2, $3, $4, 'created', 'chatroom'
)
RETURNING *;

-- name: CreateChatroom :one
INSERT INTO chatrooms (
 id, name, workflow_state, type, created_by
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING *;

-- name: UpdateChatroom :one
UPDATE chatrooms
  set name = $2,
  workflow_state = $3,
  type = $4
WHERE id = $1
RETURNING *;