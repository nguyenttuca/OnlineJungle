ALTER TABLE test_cases DROP CONSTRAINT IF EXISTS uq_test_cases_problem_ordinal;
ALTER TABLE test_cases DROP COLUMN IF EXISTS description;

ALTER TABLE problems DROP CONSTRAINT IF EXISTS problems_checker_type_check;
ALTER TABLE problems ADD CONSTRAINT problems_checker_type_check 
CHECK (checker_type IN ('diff', 'custom'));
