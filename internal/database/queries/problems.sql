-- name: GetProblemByID :one
SELECT * FROM problems WHERE id = $1;

-- name: GetProblemBySlug :one
SELECT * FROM problems WHERE slug = $1;

-- name: ListProblems :many
SELECT * FROM problems ORDER BY category, title;

-- name: ListProblemsByCategory :many
SELECT * FROM problems WHERE category = $1 ORDER BY title;

-- name: GetProblemCategories :many
SELECT DISTINCT category FROM problems ORDER BY category;

-- name: CreateProblem :one
INSERT INTO problems (slug, title, category, time_limit_ms, memory_limit_mb, statement_md, input_desc, output_desc, constraints_desc, examples, checker_type, custom_checker_code, editorial_content, tags, testcase_visibility, mirror_from)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
RETURNING *;

-- name: UpdateProblem :exec
UPDATE problems SET title = $2, slug = $3, category = $4, time_limit_ms = $5, memory_limit_mb = $6,
    statement_md = $7, input_desc = $8, output_desc = $9, constraints_desc = $10,
    examples = $11, checker_type = $12, custom_checker_code = $13, editorial_content = $14,
    tags = $15, testcase_visibility = $16, mirror_from = $17
WHERE id = $1;

-- name: DeleteProblem :exec
DELETE FROM problems WHERE id = $1;
