# Top OJ (Online Judge)

Welcome to **Top OJ**, a modern, high-performance Online Judge system built with Golang, PostgreSQL, and Bootstrap 5. It is designed to host competitive programming contests, manage algorithmic problems, and securely grade user submissions in real-time.

## 🚀 Key Features

- **Asynchronous Judging Pipeline**: Uses a robust Producer-Consumer dispatcher to manage thousands of submissions via background worker goroutines, ensuring the web server is never blocked.
- **Multiple Contest Formats**: Supports real-time standings for both **ICPC** (penalty-based) and **IOI** (max-score-based) ranking systems.
- **Remote Execution Model**: Seamlessly delegates code execution to isolated external Judge Nodes via a standardized HTTP API.
- **Type-Safe Database Access**: Powered by `sqlc`, generating raw PostgreSQL queries into completely type-safe Go structs.
- **Secure by Default**: Built-in HTTP-only sessions, bcrypt password hashing, and strict Content Security Policy (CSP) headers.
- **No-Bloat Frontend**: Server-Side Rendering (SSR) via `html/template` combined with standard CSS (Bootstrap), meaning zero heavy JavaScript frameworks to manage.

## 📁 Repository Structure

- `cmd/server/`: Main application entrypoint.
- `internal/`: Core business logic (Handlers, Dispatcher, Judge Client, Database).
- `static/`: Static assets (CSS, JS, Fonts).
- `templates/`: Go HTML templates for rendering pages.
- `docs/`: In-depth documentation for architecture and AI agents.

## 📖 Documentation

Everything you need to know to develop, deploy, and extend the system is documented in the repository:

- **[Deployment Guide](DEPLOYMENT.md)**: Standard guide to set up the environment, database, and run the server.
- **[VPS Deployment Guide](docs/VPS_DEPLOYMENT.md)**: A step-by-step tutorial on deploying to a fresh Ubuntu VPS with Nginx and SSL.
- **[AI & System Context](docs/AI_CONTEXT.md)**: Read this first if you are an AI assistant or a new developer trying to understand the core design patterns.
- **[Database Schema](docs/DB_SCHEMA.md)**: Explains the PostgreSQL tables and their relationships.
- **[Judge Flow Architecture](docs/JUDGE_FLOW.md)**: Details the asynchronous worker pool and dispatcher mechanism.

## 🛠 Tech Stack

- **Language**: Go (Golang) 1.21+
- **Database**: PostgreSQL 15+
- **Router**: `go-chi/chi`
- **DB Generator**: `sqlc`
- **Frontend**: Bootstrap 5 + FontAwesome

## 💻 Getting Started (Quickstart)

1. Clone the repository.
2. Run `go mod tidy`.
3. Set up a PostgreSQL database and apply your SQL schema.
4. Set the environment variable: `export DATABASE_URL="postgres://user:pass@localhost:5432/oj?sslmode=disable"`.
5. Start the server: `go run ./cmd/server`.

## 🤝 Contributing

When contributing, ensure that all SQL changes are made in the `internal/database/queries/` directory and compiled using `sqlc generate`. Do not write raw SQL queries directly inside Go files.
