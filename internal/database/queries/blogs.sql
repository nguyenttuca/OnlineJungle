-- name: GetBlogs :many
SELECT blogs.*, users.username AS author_username
FROM blogs
JOIN users ON blogs.author_id = users.id
WHERE is_published = true
ORDER BY blogs.created_at DESC;

-- name: GetBlogBySlug :one
SELECT blogs.*, users.username AS author_username
FROM blogs
JOIN users ON blogs.author_id = users.id
WHERE blogs.slug = $1 LIMIT 1;

-- name: GetAllBlogsAdmin :many
SELECT blogs.*, users.username AS author_username
FROM blogs
JOIN users ON blogs.author_id = users.id
ORDER BY blogs.created_at DESC;

-- name: GetBlogByIDAdmin :one
SELECT * FROM blogs WHERE id = $1 LIMIT 1;

-- name: CreateBlog :one
INSERT INTO blogs (title, slug, content_md, author_id, is_published)
VALUES ($1, $2, $3, $4, $5)
RETURNING id;

-- name: UpdateBlog :exec
UPDATE blogs
SET title = $2, slug = $3, content_md = $4, is_published = $5, updated_at = NOW()
WHERE id = $1;

-- name: DeleteBlog :exec
DELETE FROM blogs WHERE id = $1;
