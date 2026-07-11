-- Performance: Index cho filter problems theo category
CREATE INDEX IF NOT EXISTS idx_problems_category ON problems(category);

-- Performance: Index cho search submissions theo status  
CREATE INDEX IF NOT EXISTS idx_submissions_status ON submissions(status);

-- Security: Audit log table — ghi lại mọi hành động admin
CREATE TABLE IF NOT EXISTS admin_audit_logs (
    id BIGSERIAL PRIMARY KEY,
    admin_id BIGINT NOT NULL REFERENCES users(id),
    action VARCHAR(100) NOT NULL,          -- e.g. 'create_problem', 'update_test', 'create_judge_node'
    target_type VARCHAR(50) NOT NULL,      -- e.g. 'problem', 'test_case', 'judge_node', 'contest'
    target_id BIGINT,                      -- ID of affected resource
    details JSONB DEFAULT '{}',            -- Additional context (old_value, new_value, etc.)
    ip_address VARCHAR(45) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_audit_admin_time ON admin_audit_logs(admin_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_target ON admin_audit_logs(target_type, target_id);
