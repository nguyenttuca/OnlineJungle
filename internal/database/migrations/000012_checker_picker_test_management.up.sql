-- Modify problems table checker_type constraint
ALTER TABLE problems DROP CONSTRAINT IF EXISTS problems_checker_type_check;

ALTER TABLE problems ADD CONSTRAINT problems_checker_type_check 
CHECK (checker_type IN ('diff', 'fcmp', 'hcmp', 'lcmp', 'ncmp', 'nyesno', 'rcmp4', 'rcmp6', 'rcmp9', 'wcmp', 'yesno', 'custom'));

-- Add description to test_cases
ALTER TABLE test_cases ADD COLUMN IF NOT EXISTS description TEXT NULL;

-- Add unique constraint for problem_id and order_index
ALTER TABLE test_cases ADD CONSTRAINT uq_test_cases_problem_ordinal UNIQUE (problem_id, order_index);
