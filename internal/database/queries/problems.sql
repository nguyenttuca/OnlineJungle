-- name: GetProblemByID :one
SELECT id, slug, title, category, time_limit_ms, memory_limit_mb, statement_md, input_desc, output_desc, constraints_desc, examples, checker_type, custom_checker_code, created_at, editorial_content, tags, testcase_visibility, mirror_from, is_hidden FROM problems WHERE id = $1;

-- name: GetProblemBySlug :one
SELECT id, slug, title, category, time_limit_ms, memory_limit_mb, statement_md, input_desc, output_desc, constraints_desc, examples, checker_type, custom_checker_code, created_at, editorial_content, tags, testcase_visibility, mirror_from, is_hidden FROM problems WHERE slug = $1;

-- name: ListProblems :many
SELECT id, slug, title, category, time_limit_ms, memory_limit_mb, statement_md, input_desc, output_desc, constraints_desc, examples, checker_type, custom_checker_code, created_at, editorial_content, tags, testcase_visibility, mirror_from, is_hidden FROM problems ORDER BY category, title;

-- name: ListProblemsByCategory :many
SELECT id, slug, title, category, time_limit_ms, memory_limit_mb, statement_md, input_desc, output_desc, constraints_desc, examples, checker_type, custom_checker_code, created_at, editorial_content, tags, testcase_visibility, mirror_from, is_hidden FROM problems WHERE category = $1 ORDER BY title;

-- name: GetProblemCategories :many
SELECT DISTINCT category FROM problems ORDER BY category;

-- name: SearchProblems :many
SELECT p.id, p.slug, p.title, p.category, p.time_limit_ms, p.memory_limit_mb, p.statement_md, p.input_desc, p.output_desc, p.constraints_desc, p.examples, p.checker_type, p.custom_checker_code, p.created_at, p.editorial_content, p.tags, p.testcase_visibility, p.mirror_from, p.is_hidden,
  COALESCE((
    SELECT s.verdict
    FROM submissions s
    WHERE s.problem_id = p.id
      AND s.user_id = sqlc.narg('user_id')
    ORDER BY (CASE WHEN s.verdict = 'AC' THEN 1 ELSE 2 END), s.submitted_at DESC
    LIMIT 1
  ), '') AS user_status
FROM problems p
WHERE
  (@search_query::text = '' OR p.title ILIKE '%' || @search_query || '%' OR p.slug ILIKE '%' || @search_query || '%' OR p.id::text = @search_query)
  AND (@category::text = '' OR p.category = @category)
  AND (sqlc.arg('include_hidden')::boolean = true OR p.is_hidden = false)
ORDER BY p.category, p.title;

-- name: CreateProblem :one
INSERT INTO problems (slug, title, category, time_limit_ms, memory_limit_mb, statement_md, input_desc, output_desc, constraints_desc, examples, checker_type, custom_checker_code, editorial_content, tags, testcase_visibility, mirror_from, is_hidden)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
RETURNING id, slug, title, category, time_limit_ms, memory_limit_mb, statement_md, input_desc, output_desc, constraints_desc, examples, checker_type, custom_checker_code, created_at, editorial_content, tags, testcase_visibility, mirror_from, is_hidden;

-- name: UpdateProblem :exec
UPDATE problems SET title = $2, slug = $3, category = $4, time_limit_ms = $5, memory_limit_mb = $6,
    statement_md = $7, input_desc = $8, output_desc = $9, constraints_desc = $10,
    examples = $11, checker_type = $12, custom_checker_code = $13, editorial_content = $14,
    tags = $15, testcase_visibility = $16, mirror_from = $17, is_hidden = $18
WHERE id = $1;

-- name: DeleteProblem :exec
DELETE FROM problems WHERE id = $1;
