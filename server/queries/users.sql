-- name: GetUser :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = $1 LIMIT 1;

-- name: GetUserPassword :one
SELECT password FROM users
WHERE id = $1 LIMIT 1;

-- name: GetPrivateConversation :many
SELECT m.*,
  s.username sender_username,
  s.email sender_email,
  s.workflow_state sender_workflow_state,
  r.username recipient_username,
  r.email recipient_email,
  r.workflow_state recipient_workflow_state
FROM messages m
JOIN users s ON s.id = m.sender_id
JOIN users r ON r.id = m.recipient_id
WHERE (m.sender_id = @first_user_id AND m.recipient_id = @second_user_id)
  OR (m.sender_id = @second_user_id AND m.recipient_id = @first_user_id)
ORDER BY m.created_at DESC
LIMIT @page_size
OFFSET @page;

-- name: GetFriend :one
SELECT u.*
FROM (
  SELECT 
    CASE
      WHEN f.user_id = @user_id THEN f.requester_id
      WHEN f.requester_id = @user_id THEN f.user_id
    END AS id
  FROM friendships f
  WHERE f.workflow_state = 'active' AND 
        ((f.requester_id = @user_id AND f.user_id = @friend_id) OR
         (f.requester_id = @friend_id AND f.user_id = @user_id))
  LIMIT 1
) AS friends
JOIN users u ON u.id = friends.id
LIMIT 1;

-- name: GetFriendsByFriendshipWorkflowStates :many
SELECT u.*
FROM (
  SELECT 
    CASE
      WHEN f.user_id = @id THEN f.requester_id
      WHEN f.requester_id = @id THEN f.user_id
    END AS id
  FROM friendships f
  WHERE f.workflow_state = ANY(@friendship_workflow_states::friendship_workflow_state[]) AND
        (f.requester_id = @id OR f.user_id = @id)
) AS friends
JOIN users u ON u.id = friends.id
LIMIT @page_size
OFFSET @page;

-- name: CreateUser :one
INSERT INTO users (
  email, username, password, workflow_state
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (
  user_id, token, expires_at
) VALUES (
  $1, $2, $3
)
RETURNING *;