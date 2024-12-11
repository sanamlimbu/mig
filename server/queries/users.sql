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
SELECT * FROM messages m1
WHERE m1.sender_id = @first_user_id AND m1.recipient_id = @second_user_id 
UNION
SELECT * FROM messages m2
WHERE m2.sender_id = @second_user_id AND m2.recipient_id = @first_user_id
ORDER BY created_at DESC
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

-- name: GetFriends :many
SELECT u.*
FROM (
  SELECT 
    CASE
      WHEN f.user_id = @id THEN f.requester_id
      WHEN f.requester_id = @id THEN f.user_id
    END AS id
  FROM friendships f
  WHERE f.workflow_state = 'active' AND 
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