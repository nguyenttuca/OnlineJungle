-- name: CreateClass :one
INSERT INTO classes (name, description, weekly_schedule, notice_md)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetClassByID :one
SELECT * FROM classes WHERE id = $1;

-- name: ListClasses :many
SELECT * FROM classes ORDER BY created_at DESC;

-- name: UpdateClass :exec
UPDATE classes SET name = $2, description = $3, weekly_schedule = $4, notice_md = $5 WHERE id = $1;

-- name: AddClassMember :exec
INSERT INTO class_members (class_id, user_id, role)
VALUES ($1, $2, $3)
ON CONFLICT (class_id, user_id) DO UPDATE SET role = EXCLUDED.role;

-- name: UpdateClassMemberRole :exec
UPDATE class_members SET role = $3 WHERE class_id = $1 AND user_id = $2;

-- name: RemoveClassMember :exec
DELETE FROM class_members WHERE class_id = $1 AND user_id = $2;

-- name: GetClassMembers :many
SELECT u.id, u.username, u.display_name, cm.role, cm.created_at
FROM class_members cm
JOIN users u ON u.id = cm.user_id
WHERE cm.class_id = $1
ORDER BY (CASE WHEN cm.role = 'teacher' THEN 1 WHEN cm.role = 'student' THEN 2 ELSE 3 END), u.username;

-- name: GetUserRoleInClass :one
SELECT role FROM class_members WHERE class_id = $1 AND user_id = $2;

-- name: AddProblemToClass :exec
INSERT INTO class_problems (class_id, problem_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;

-- name: RemoveProblemFromClass :exec
DELETE FROM class_problems WHERE class_id = $1 AND problem_id = $2;

-- name: AddContestToClass :exec
INSERT INTO class_contests (class_id, contest_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;

-- name: RemoveContestFromClass :exec
DELETE FROM class_contests WHERE class_id = $1 AND contest_id = $2;

-- name: ListClassProblems :many
SELECT p.id, p.slug, p.title, p.category, p.time_limit_ms, p.memory_limit_mb, p.created_at, p.is_hidden
FROM class_problems cp
JOIN problems p ON p.id = cp.problem_id
WHERE cp.class_id = $1
ORDER BY p.title;

-- name: ListClassContests :many
SELECT c.id, c.title, c.start_at, c.end_at, c.created_at, c.ranking_type, c.is_hidden
FROM class_contests cc
JOIN contests c ON c.id = cc.contest_id
WHERE cc.class_id = $1
ORDER BY c.start_at DESC;

-- name: CheckProblemInClass :one
SELECT EXISTS (
    SELECT 1 FROM class_problems WHERE class_id = $1 AND problem_id = $2
);

-- name: CheckContestInClass :one
SELECT EXISTS (
    SELECT 1 FROM class_contests WHERE class_id = $1 AND contest_id = $2
);

-- name: GetClassesForProblem :many
SELECT class_id FROM class_problems WHERE problem_id = $1;

-- name: GetClassesForContest :many
SELECT class_id FROM class_contests WHERE contest_id = $1;

-- name: GetUserClasses :many
SELECT c.*, cm.role
FROM classes c
JOIN class_members cm ON cm.class_id = c.id
WHERE cm.user_id = $1
ORDER BY c.created_at DESC;
