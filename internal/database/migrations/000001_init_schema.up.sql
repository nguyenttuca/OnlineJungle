-- Online Judge Schema
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    display_name VARCHAR(100) NOT NULL DEFAULT '',
    country VARCHAR(50) NOT NULL DEFAULT '',
    role VARCHAR(20) NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'admin')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE problems (
    id BIGSERIAL PRIMARY KEY,
    slug VARCHAR(100) UNIQUE NOT NULL,
    title VARCHAR(255) NOT NULL,
    category VARCHAR(100) NOT NULL DEFAULT 'Uncategorized',
    time_limit_ms INT NOT NULL DEFAULT 1000,
    memory_limit_mb INT NOT NULL DEFAULT 256,
    statement_md TEXT NOT NULL DEFAULT '',
    input_desc TEXT NOT NULL DEFAULT '',
    output_desc TEXT NOT NULL DEFAULT '',
    constraints_desc TEXT NOT NULL DEFAULT '',
    examples JSONB NOT NULL DEFAULT '[]',
    checker_type VARCHAR(20) NOT NULL DEFAULT 'diff' CHECK (checker_type IN ('diff', 'custom')),
    custom_checker_code TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE subtasks (
    id BIGSERIAL PRIMARY KEY,
    problem_id BIGINT NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
    subtask_index INT NOT NULL,
    points INT NOT NULL DEFAULT 0,
    description TEXT NOT NULL DEFAULT '',
    UNIQUE(problem_id, subtask_index)
);

CREATE TABLE test_cases (
    id BIGSERIAL PRIMARY KEY,
    problem_id BIGINT NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
    order_index INT NOT NULL DEFAULT 0,
    input TEXT NOT NULL DEFAULT '',
    expected_output TEXT NOT NULL DEFAULT '',
    is_sample BOOLEAN NOT NULL DEFAULT false,
    subtask_id BIGINT REFERENCES subtasks(id) ON DELETE SET NULL
);
CREATE INDEX idx_test_cases_problem ON test_cases(problem_id, order_index);

CREATE TABLE judge_nodes (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    base_url VARCHAR(500) NOT NULL,
    api_key_encrypted VARCHAR(500) NOT NULL DEFAULT '',
    is_active BOOLEAN NOT NULL DEFAULT true,
    supported_languages JSONB NOT NULL DEFAULT '{"cpp": false, "c": false, "python": false, "pascal": false}',
    max_concurrent INT NOT NULL DEFAULT 3,
    active_jobs INT NOT NULL DEFAULT 0,
    is_healthy BOOLEAN NOT NULL DEFAULT false,
    last_health_check_at TIMESTAMPTZ,
    consecutive_failures INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE submissions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    problem_id BIGINT NOT NULL REFERENCES problems(id),
    language VARCHAR(20) NOT NULL,
    source_code TEXT NOT NULL,
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status VARCHAR(20) NOT NULL DEFAULT 'queued' CHECK (status IN ('queued', 'dispatched', 'judging', 'done', 'failed')),
    verdict VARCHAR(20) NOT NULL DEFAULT '',
    compile_output TEXT NOT NULL DEFAULT '',
    time_ms INT NOT NULL DEFAULT 0,
    memory_kb INT NOT NULL DEFAULT 0,
    code_size INT NOT NULL DEFAULT 0,
    run_all_tests BOOLEAN NOT NULL DEFAULT true,
    judge_node_id BIGINT REFERENCES judge_nodes(id),
    attempts INT NOT NULL DEFAULT 0
);
CREATE INDEX idx_submissions_queue ON submissions(status, submitted_at) WHERE status = 'queued';
CREATE INDEX idx_submissions_user_problem ON submissions(user_id, problem_id);
CREATE INDEX idx_submissions_problem ON submissions(problem_id, submitted_at DESC);

CREATE TABLE submission_test_results (
    id BIGSERIAL PRIMARY KEY,
    submission_id BIGINT NOT NULL REFERENCES submissions(id) ON DELETE CASCADE,
    test_index INT NOT NULL,
    verdict VARCHAR(20) NOT NULL DEFAULT '',
    time_ms INT NOT NULL DEFAULT 0,
    memory_kb INT NOT NULL DEFAULT 0,
    stdout TEXT NOT NULL DEFAULT '',
    stderr TEXT NOT NULL DEFAULT ''
);
CREATE INDEX idx_str_submission ON submission_test_results(submission_id, test_index);

CREATE TABLE contests (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    start_at TIMESTAMPTZ NOT NULL,
    end_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE contest_problems (
    contest_id BIGINT NOT NULL REFERENCES contests(id) ON DELETE CASCADE,
    problem_id BIGINT NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
    label VARCHAR(10) NOT NULL DEFAULT '',
    points INT NOT NULL DEFAULT 0,
    PRIMARY KEY (contest_id, problem_id)
);

CREATE TABLE sessions (
    token TEXT PRIMARY KEY,
    data BYTEA NOT NULL,
    expiry TIMESTAMPTZ NOT NULL
);
CREATE INDEX sessions_expiry_idx ON sessions(expiry);
