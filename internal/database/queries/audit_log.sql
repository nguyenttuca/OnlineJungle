-- name: CreateAuditLog :exec
INSERT INTO admin_audit_logs (admin_id, action, target_type, target_id, details, ip_address)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: ListAuditLogs :many
SELECT al.*, u.username as admin_username
FROM admin_audit_logs al
JOIN users u ON u.id = al.admin_id
ORDER BY al.created_at DESC
LIMIT sqlc.arg(limit_val) OFFSET sqlc.arg(offset_val);

-- name: ListAuditLogsByAdmin :many
SELECT * FROM admin_audit_logs WHERE admin_id = $1
ORDER BY created_at DESC
LIMIT sqlc.arg(limit_val) OFFSET sqlc.arg(offset_val);

-- name: ListAuditLogsByTarget :many
SELECT * FROM admin_audit_logs WHERE target_type = $1 AND target_id = $2
ORDER BY created_at DESC;
