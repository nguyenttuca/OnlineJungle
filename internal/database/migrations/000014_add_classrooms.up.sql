CREATE TABLE classes (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    schedule_md TEXT NOT NULL DEFAULT '',
    notice_md TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE class_members (
    class_id BIGINT NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (role IN ('pending', 'student', 'teacher')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (class_id, user_id)
);

CREATE TABLE class_problems (
    class_id BIGINT NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    problem_id BIGINT NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (class_id, problem_id)
);

CREATE TABLE class_contests (
    class_id BIGINT NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    contest_id BIGINT NOT NULL REFERENCES contests(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (class_id, contest_id)
);

ALTER TABLE problems ADD COLUMN is_hidden BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE contests ADD COLUMN is_hidden BOOLEAN NOT NULL DEFAULT false;
