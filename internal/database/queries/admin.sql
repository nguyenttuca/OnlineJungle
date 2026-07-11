-- name: CreateGroup :one
INSERT INTO groups (name, description)
VALUES ($1, $2) RETURNING *;


-- name: CreateAnnouncement :one
INSERT INTO announcements (title, content_md, is_active)
VALUES ($1, $2, $3) RETURNING *;


