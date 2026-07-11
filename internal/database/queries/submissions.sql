-- name: GetSubmissionByID :one
SELECT * FROM submissions WHERE id = $1;

-- name: ListSubmissionsByProblem :many
SELECT * FROM submissions WHERE problem_id = $1 ORDER BY submitted_at DESC LIMIT sqlc.arg(limit_val) OFFSET sqlc.arg(offset_val);

-- name: ListSubmissionsByUser :many
SELECT * FROM submissions WHERE user_id = $1 ORDER BY submitted_at DESC LIMIT sqlc.arg(limit_val) OFFSET sqlc.arg(offset_val);

-- name: ListSubmissionsByUserAndProblem :many
SELECT * FROM submissions WHERE user_id = $1 AND problem_id = $2 ORDER BY submitted_at DESC LIMIT sqlc.arg(limit_val) OFFSET sqlc.arg(offset_val);

-- name: GetPendingSubmissionByUser :one
SELECT * FROM submissions WHERE user_id = $1 AND problem_id = $2 AND status IN ('queued', 'dispatched', 'judging') LIMIT 1;

-- name: CreateSubmission :one
INSERT INTO submissions (user_id, problem_id, contest_id, language, source_code, code_size, run_all_tests, score)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: DequeueSubmission :one
WITH next AS (
    SELECT id FROM submissions
    WHERE status = 'queued'
    ORDER BY submitted_at ASC
    FOR UPDATE SKIP LOCKED
    LIMIT 1
)
UPDATE submissions SET status = 'dispatched', attempts = attempts + 1
FROM next WHERE submissions.id = next.id
RETURNING submissions.*;

-- name: UpdateSubmissionStatus :exec
UPDATE submissions SET status = $2 WHERE id = $1;

-- name: UpdateSubmissionJudging :exec
UPDATE submissions SET status = 'judging', judge_node_id = $2 WHERE id = $1;

-- name: UpdateSubmissionResult :exec
UPDATE submissions SET status = 'done', verdict = $2, time_ms = $3, memory_kb = $4, compile_output = $5, score = CASE WHEN CAST($2 AS varchar) = 'AC' THEN 100 ELSE 0 END WHERE id = $1;

-- name: UpdateSubmissionFailed :exec
UPDATE submissions SET status = 'failed', verdict = 'SYSTEM_ERROR', compile_output = $2 WHERE id = $1;

-- name: RequeueSubmission :exec
UPDATE submissions SET status = 'queued' WHERE id = $1;

-- name: CountSubmissionsByProblem :one
SELECT COUNT(*) FROM submissions WHERE problem_id = $1;

-- name: HasUserSolvedProblem :one
SELECT EXISTS(SELECT 1 FROM submissions WHERE user_id = $1 AND problem_id = $2 AND verdict = 'AC');

-- name: GetUserSolvedProblems :many
SELECT DISTINCT problem_id FROM submissions WHERE user_id = $1 AND verdict = 'AC';

-- name: ListAllSubmissions :many
SELECT * FROM submissions ORDER BY submitted_at DESC LIMIT sqlc.arg(limit_val) OFFSET sqlc.arg(offset_val);


-- name: GetContestSubmissions :many
SELECT * FROM submissions
WHERE contest_id = $1 AND user_id = $2
ORDER BY submitted_at DESC;
