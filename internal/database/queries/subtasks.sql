-- name: ListSubtasksByProblem :many
SELECT * FROM subtasks WHERE problem_id = $1 ORDER BY subtask_index;

-- name: CreateSubtask :one
INSERT INTO subtasks (problem_id, subtask_index, points, description)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: DeleteSubtask :exec
DELETE FROM subtasks WHERE id = $1;
