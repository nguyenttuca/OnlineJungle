CREATE INDEX idx_submissions_contest_standings
ON submissions (contest_id, user_id, problem_id, submitted_at, score);
