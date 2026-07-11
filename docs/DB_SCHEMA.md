# Database Schema Overview

This file explains the core Postgres SQL tables and their relationships for the Top OJ system. 

## Tables Overview

### 1. `users`
Stores all account information.
- `id` (PK, int)
- `username`, `email` (Unique)
- `password_hash` (Bcrypt)
- `role` ('user' or 'admin')

### 2. `problems`
Stores algorithmic problems.
- `id` (PK, int)
- `slug` (Unique identifier for URL, e.g. `a-plus-b`)
- `title`, `statement` (Markdown/HTML text)
- `time_limit_ms`, `memory_limit_mb` (Limits sent to judge)
- `is_public` (Boolean: visible outside contests)
- `checker_type`, `custom_checker_code` (Testlib checker support)

### 3. `test_cases`
Stores exact I/O cases for a problem.
- `id` (PK)
- `problem_id` (FK to problems)
- `input`, `expected_output` (TEXT data)
- `is_sample` (Boolean: show in problem description?)
- `points` (Partial score if partial scoring is supported in the future)

### 4. `contests`
Stores exam/contest periods.
- `id` (PK)
- `title`, `start_at`, `end_at` (Timestamps)
- `ranking_type` (`IOI` or `ICPC`)

### 5. `contest_problems`
Many-to-Many junction table linking Problems to Contests.
- `contest_id` (FK)
- `problem_id` (FK)
- `label` (e.g. 'A', 'B', 'C')
- `points` (Max points for this problem in this contest)
- *Primary Key*: `(contest_id, problem_id)`

### 6. `submissions`
Core table for all user submitted code.
- `id` (PK)
- `user_id`, `problem_id` (FK)
- `contest_id` (Optional FK, NULL if practice mode)
- `language` (e.g. 'cpp', 'py')
- `source_code`, `code_size`
- `status` (`queued`, `dispatched`, `judging`, `done`, `failed`)
- `verdict` (`AC`, `WA`, `TLE`, `MLE`, `RE`, `CE`, `SYSTEM_ERROR`)
- `score` (Integer: 0 to 100. Always 100 for AC, 0 for others currently).
- `time_ms`, `memory_kb`, `compile_output`

### 7. `judge_nodes`
Stores configuration for remote workers.
- `id` (PK)
- `name` (Identifier)
- `base_url` (API endpoint of the node)
- `api_key` (Auth for the node)
- `is_active` (Boolean)
- `max_concurrent_jobs` (Limit load per node)

## Key Constraints & Relationships
- A `Submission` ties to a `User` and `Problem`. If `contest_id` is present, it will be included in the contest's live standings.
- The `Dispatcher` loops through `submissions` where `status = 'queued'`, marks them `dispatched`, and forwards them to an active `judge_nodes`.
