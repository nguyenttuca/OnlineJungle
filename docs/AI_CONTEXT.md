# Project Context & Overview for AI Agents

**Target Audience:** AI Coding Assistants / LLMs.
**Purpose:** Provide a highly dense, token-optimized summary of the Top OJ system architecture, tech stack, and codebase structure. Read this file first to understand the project without needing to grep through the entire repository.

## 1. System Architecture

Top OJ is a monolithic web server with a built-in asynchronous Worker Pool for code evaluation.
- **Frontend:** Server-Side Rendered (SSR) HTML pages using Go's `html/template`. Styled with Bootstrap 5 (CDN) + custom `style.css`.
- **Backend:** Go (`net/http` + `chi` router).
- **Database:** PostgreSQL.
- **ORM / DB Access:** `sqlc`. Queries are written in raw SQL (`internal/database/queries/*.sql`) and compiled to Go structs via `sqlc generate`. **Do NOT use GORM or raw `database/sql` queries in Go code; always write `.sql` files and run `sqlc generate`.**
- **Judge Engine:** A remote-execution model. The web server itself does not compile/run user code. It sends payloads (source code, memory/time limits, test cases) via HTTP to remote Judge Nodes (managed in `internal/judgepool/client.go`).

## 2. Directory Structure

```text
oj-web/
├── cmd/
│   └── server/          # Entry point (main.go). Wires up DB, env, dispatcher, router.
├── internal/
│   ├── database/        # Contains SQL queries (.sql) and sqlc-generated Go code (sqlcdb/)
│   ├── dispatcher/      # Core logic for background grading. Manages queued submissions -> workers.
│   ├── handlers/        # HTTP controllers (chi router). Handles web requests & SSR.
│   └── judgepool/       # HTTP Client interacting with the external Judge API.
├── static/              # CSS, JS, Images (served at /static/)
├── templates/           # Go HTML templates (layouts/ for wrappers, pages/ for views, partials/)
├── .env.example         # Example environment variables
├── DEPLOYMENT.md        # Deployment instructions for human developers
├── go.mod / go.sum      # Go module dependencies
└── sqlc.yaml            # sqlc configuration file
```

## 3. Core Tech Patterns (Agent Rules)

1. **Routing:** `chi` router is defined in `internal/handlers/routes.go`.
2. **Context & Environment:** Handlers receive an `Env` struct containing database `Queries` (from sqlc) and session management.
3. **Authentication:** Managed via HTTP-only Cookies and standard Middleware (e.g., `RequireAuth`, `RequireAdmin`). User IDs are stored in `context.Context`.
4. **Security:** `security.go` sets strict HTTP headers (CSP, X-Frame-Options, etc.). If adding new CDNs or scripts, CSP must be updated here.
5. **Score & Standings:** Standings are computed dynamically using SQL `WITH` (CTE) clauses in `contests.sql` rather than caching in memory.
    - ICPC Style: Ranks by `solved_count` DESC, `total_penalty_minutes` ASC.
    - IOI Style: Ranks by `total_score` DESC, `username` ASC (No time penalty).

## 4. How to Extend the System
- **Adding a DB Table/Query:**
  1. Add/modify `CREATE TABLE` schema (if needed, ensure it matches DB state).
  2. Write standard Postgres SQL query in `internal/database/queries/<domain>.sql`.
  3. Run `sqlc generate` in the terminal.
  4. Use the newly generated Go function (`env.Queries.YourNewFunction(ctx, args)`) in Handlers.
- **Modifying UI:** Edit `.html` files in `templates/pages/`. Pass data from Handlers via the `render` function's `data map[string]interface{}` argument.
