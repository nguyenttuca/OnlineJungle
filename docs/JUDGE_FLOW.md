# Judge System Architecture & Flow

This document details how a user's code goes from clicking "Submit" to receiving a Verdict (AC, WA, etc.) in Top OJ.

## The Async Pipeline

Top OJ uses an asynchronous Producer-Consumer pipeline to prevent web server blocking and to gracefully queue thousands of submissions if judges are busy.

### Step 1: HTTP Submit (Web Server)
1. User submits code via the browser.
2. `SubmitPostHandler` or `ContestSubmitPostHandler` (in `internal/handlers`) inserts a new record into the `submissions` table with `status = 'queued'` and `score = 0`.
3. The HTTP response immediately redirects the user to the submission detail page. At this point, the submission is pending.

### Step 2: The Dispatcher Producer (Internal)
1. When the server boots (`cmd/server/main.go`), it calls `dispatcher.Start()`.
2. The `producer` goroutine runs on a 500ms ticker (`time.Ticker`).
3. It calls `d.Queries.DequeueSubmission()` which runs an atomic SQL query:
   ```sql
   WITH next AS (
       SELECT id FROM submissions WHERE status = 'queued' ORDER BY id ASC LIMIT 1 FOR UPDATE SKIP LOCKED
   )
   UPDATE submissions SET status = 'dispatched' WHERE id = (SELECT id FROM next) RETURNING *;
   ```
4. If a submission is found, it sends it to the `d.jobs` channel.

### Step 3: The Dispatcher Consumer (Worker Goroutines)
1. The dispatcher spawns N `worker` goroutines (currently N=10 by default) that listen to the `d.jobs` channel.
2. When a worker receives a `dispatched` submission:
   - Fetches the Problem and its TestCases from the Database.
   - Finds an active/healthy remote Judge Node (from `judge_nodes` table) via `getHealthyNode()`.
   - Calls `d.Queries.UpdateSubmissionJudging()` to update status to `judging`.
   - Sends an HTTP POST payload to the remote Judge Node containing source code, limits, and test cases via `internal/judgepool/client.go`.
3. The worker blocks waiting for the Judge Node to finish (up to 30s timeout).

### Step 4: The Result (Completion)
1. The Judge Node returns a JSON response containing `Verdict`, `CompileOutput`, and an array of `TestResult`.
2. The worker extracts the overall Verdict (e.g., `AC`, `WA`) and stats (Max Time, Max Memory).
3. The worker calls `d.Queries.UpdateSubmissionResult()` to update the database:
   - `status = 'done'`
   - `verdict = 'AC'`
   - `score = 100` (if AC, else 0)
   - Time/Memory metrics
4. The user refreshes their page and sees the final result.

### CRITICAL RULES
- **Never call `DequeueSubmission` from a Web Handler.** Doing so breaks the pipeline and leaves the submission stuck in `dispatched` state forever, because handlers don't have access to the `d.jobs` channel to pass it to the workers.
- **Worker pool sizing:** Controlled by `dispatcher.NewDispatcher(...)`.
- **Score Calculation:** Currently binary. The `UpdateSubmissionResult` SQL query handles setting `score = 100` for `AC` and `0` for others natively in SQL. Do not try to update `score` from Go unless you are implementing partial scoring based on `TestResults`.
