-- name: GetJudgeNodeByID :one
SELECT * FROM judge_nodes WHERE id = $1;

-- name: ListJudgeNodes :many
SELECT * FROM judge_nodes ORDER BY id;

-- name: ListActiveHealthyNodes :many
SELECT * FROM judge_nodes WHERE is_active = true AND is_healthy = true AND active_jobs < max_concurrent ORDER BY is_local ASC, active_jobs ASC, id ASC;

-- name: ListHealthyNodesForLanguage :many
SELECT * FROM judge_nodes WHERE is_active = true AND is_healthy = true AND supported_languages->>sqlc.arg(language) = 'true';

-- name: CreateJudgeNode :one
INSERT INTO judge_nodes (name, base_url, api_key_encrypted, is_active, max_concurrent)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateJudgeNode :exec
UPDATE judge_nodes SET name = $2, base_url = $3, api_key_encrypted = $4, is_active = $5, max_concurrent = $6 WHERE id = $1;

-- name: UpdateJudgeNodeHealth :exec
UPDATE judge_nodes SET is_healthy = $2, supported_languages = $3, max_concurrent = $4,
    last_health_check_at = $5, consecutive_failures = $6 WHERE id = $1;

-- name: IncrementNodeActiveJobs :exec
UPDATE judge_nodes SET active_jobs = active_jobs + 1 WHERE id = $1;

-- name: DecrementNodeActiveJobs :exec
UPDATE judge_nodes SET active_jobs = GREATEST(active_jobs - 1, 0) WHERE id = $1;

-- name: DeleteJudgeNode :exec
DELETE FROM judge_nodes WHERE id = $1;

-- name: ToggleJudgeNodeActive :exec
UPDATE judge_nodes SET is_active = NOT is_active WHERE id = $1;

-- name: AcquireJudgeNode :one
UPDATE judge_nodes
SET active_jobs = active_jobs + 1
WHERE id = (
    SELECT id FROM judge_nodes
    WHERE is_active = true 
      AND is_healthy = true 
      AND active_jobs < max_concurrent
    ORDER BY is_local ASC, active_jobs ASC, id ASC
    LIMIT 1
    FOR UPDATE SKIP LOCKED
)
RETURNING *;
