-- name: CreateTestResult :exec
INSERT INTO submission_test_results (submission_id, test_index, verdict, time_ms, memory_kb, stdout, stderr)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: ListTestResultsBySubmission :many
SELECT * FROM submission_test_results WHERE submission_id = $1 ORDER BY test_index;

-- name: DeleteTestResultsBySubmission :exec
DELETE FROM submission_test_results WHERE submission_id = $1;
