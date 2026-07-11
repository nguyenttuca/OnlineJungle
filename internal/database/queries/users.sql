-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1;

-- name: CreateUser :one
INSERT INTO users (username, password_hash, display_name, country, role)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListUsers :many
SELECT * FROM users ORDER BY created_at DESC LIMIT sqlc.arg(limit_val) OFFSET sqlc.arg(offset_val);

-- name: CountUsers :one
SELECT COUNT(*) FROM users;

-- name: UpdateUserProfile :exec
UPDATE users SET display_name = $2, country = $3 WHERE id = $1;

-- name: UpdateUserPassword :exec
UPDATE users SET password_hash = $2 WHERE id = $1;

-- name: GetUsersLeaderboard :many
SELECT u.id, u.username, u.display_name, u.role,
       COUNT(DISTINCT s.problem_id) AS solved_count
FROM users u
LEFT JOIN submissions s ON u.id = s.user_id AND s.verdict = 'AC'
GROUP BY u.id
ORDER BY solved_count DESC, u.id ASC
LIMIT sqlc.arg(limit_val) OFFSET sqlc.arg(offset_val);
