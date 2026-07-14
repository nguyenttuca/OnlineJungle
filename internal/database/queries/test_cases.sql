-- name: ListTestCasesByProblem :many
SELECT * FROM test_cases WHERE problem_id = $1 ORDER BY order_index;

-- name: GetSampleTestCases :many
SELECT * FROM test_cases WHERE problem_id = $1 AND is_sample = true ORDER BY order_index;

-- name: CreateTestCase :one
INSERT INTO test_cases (problem_id, order_index, input, expected_output, is_sample, subtask_id, description)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: UpdateTestCase :exec
UPDATE test_cases SET input = $2, expected_output = $3, is_sample = $4, subtask_id = $5, order_index = $6, description = $7 WHERE id = $1;

-- name: DeleteTestCase :exec
DELETE FROM test_cases WHERE id = $1;

-- name: DeleteTestCasesByProblem :exec
DELETE FROM test_cases WHERE problem_id = $1;

-- name: ReorderTestCase :exec
UPDATE test_cases SET order_index = $2 WHERE id = $1;
