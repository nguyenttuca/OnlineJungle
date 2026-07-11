-- name: CreateGroup :one
INSERT INTO groups (name, description)
VALUES ($1, $2) RETURNING *;

-- name: CreateBlog :one
INSERT INTO blogs (author_id, title, content_md)
VALUES ($1, $2, $3) RETURNING *;

-- name: CreateAnnouncement :one
INSERT INTO announcements (title, content_md, is_active)
VALUES ($1, $2, $3) RETURNING *;


