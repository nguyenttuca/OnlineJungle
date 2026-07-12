-- name: GetContestByID :one
SELECT * FROM contests WHERE id = $1;

-- name: ListContests :many
SELECT * FROM contests ORDER BY start_at DESC;

-- name: CreateContest :one
INSERT INTO contests (title, start_at, end_at, ranking_type) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: AddContestProblem :exec
INSERT INTO contest_problems (contest_id, problem_id, label, points) VALUES ($1, $2, $3, $4);

-- name: ListContestProblems :many
SELECT cp.label, cp.points, p.* FROM contest_problems cp JOIN problems p ON p.id = cp.problem_id WHERE cp.contest_id = $1 ORDER BY cp.label;

-- name: RemoveContestProblem :exec
DELETE FROM contest_problems WHERE contest_id = $1 AND problem_id = $2;

-- name: UpdateContest :exec
UPDATE contests SET title = $2, start_at = $3, end_at = $4, ranking_type = $5 WHERE id = $1;

-- name: GetContestProblem :one
SELECT p.*
FROM problems p
JOIN contest_problems cp ON cp.problem_id = p.id
WHERE cp.contest_id = $1 AND p.slug = $2;

-- name: CalculateContestStandingsIOI :many
WITH best_submissions AS (
    SELECT
        s.user_id,
        s.problem_id,
        s.score,
        s.submitted_at,
        ROW_NUMBER() OVER (
            PARTITION BY s.user_id, s.problem_id
            ORDER BY s.score DESC, s.submitted_at ASC
        ) AS rn
    FROM submissions s
    WHERE s.contest_id = sqlc.arg(contest_id)
)
SELECT
    u.id AS user_id,
    u.username,
    COALESCE(SUM(bs.score), 0)::BIGINT AS total_score
FROM best_submissions bs
JOIN users u ON u.id = bs.user_id
JOIN contests c ON c.id = sqlc.arg(contest_id)
WHERE bs.rn = 1
GROUP BY u.id, u.username
ORDER BY total_score DESC, username ASC;

-- name: CalculateContestStandingsICPC :many
WITH first_ac AS (
    SELECT
        s.user_id,
        s.problem_id,
        s.submitted_at AS ac_time,
        ROW_NUMBER() OVER (
            PARTITION BY s.user_id, s.problem_id
            ORDER BY s.submitted_at ASC
        ) AS rn
    FROM submissions s
    WHERE s.contest_id = sqlc.arg(contest_id)
      AND s.score = 100
),
wrong_attempts AS (
    SELECT
        s.user_id,
        s.problem_id,
        COUNT(*) AS wrong_count
    FROM submissions s
    JOIN first_ac fa
        ON fa.user_id = s.user_id
       AND fa.problem_id = s.problem_id
       AND fa.rn = 1
    WHERE s.contest_id = sqlc.arg(contest_id)
      AND s.submitted_at < fa.ac_time
      AND s.score < 100
    GROUP BY s.user_id, s.problem_id
)
SELECT
    u.id AS user_id,
    u.username,
    COUNT(DISTINCT fa.problem_id)::BIGINT AS solved_count,
    COALESCE(SUM(
        EXTRACT(EPOCH FROM (fa.ac_time - c.start_at))/60
        + 20 * COALESCE(wa.wrong_count, 0)
    ), 0)::BIGINT AS total_penalty_minutes
FROM first_ac fa
JOIN users u ON u.id = fa.user_id
JOIN contests c ON c.id = sqlc.arg(contest_id)
LEFT JOIN wrong_attempts wa
    ON wa.user_id = fa.user_id AND wa.problem_id = fa.problem_id
WHERE fa.rn = 1
GROUP BY u.id, u.username
ORDER BY solved_count DESC, total_penalty_minutes ASC;
