DROP TABLE IF EXISTS class_contests;
DROP TABLE IF EXISTS class_problems;
DROP TABLE IF EXISTS class_members;
DROP TABLE IF EXISTS classes;

ALTER TABLE problems DROP COLUMN IF EXISTS is_hidden;
ALTER TABLE contests DROP COLUMN IF EXISTS is_hidden;
