# go-job-search — Claude Context

## What This Service Is

The job search tracking service for the personal-enterprise project. Manages job applications and recruiters. Backed by Postgres. Validates JWTs issued by go-auth — does not issue tokens.

Built from `go-service-template`. The structure, patterns, and tooling are inherited from that template — this file documents what is specific to go-job-search on top of that foundation.

---

## Architecture

```
cmd/
  server/main.go     ← wiring: config, DB, stores, server.New()
  migrate/main.go    ← golang-migrate runner
app/
  app.go             ← Application struct: JobStore, RecruiterStore
server/
  server.go          ← chi router, global middleware
  routes.go          ← route registration — all routes auth-protected
config/
  config.go          ← standard template config, no additions needed
db_connection/
  db.go              ← pgxpool setup
db/
  schema.sql         ← jobs, recruiters tables
  sqlc.yaml
  queries/
  migrations/
  generated/
health/
  handler.go
middleware/
  auth.go
  logging.go
  requestid.go
job/
  job_model.go
  job_store.go
  job_handler.go
  job_routes.go
recruiter/
  recruiter_model.go
  recruiter_store.go
  recruiter_handler.go
  recruiter_routes.go
example/             ← template reference domain, leave until real domains are built
```

---

## All Routes Are Protected

Every domain in this service requires authentication. `authMiddleware` is applied to all mounts in `server/routes.go` — no unprotected endpoints except `/health`.

---

## Patterns Carried Over from Template

### Domain Structure

Four-file pattern: `<domain>_model.go`, `_store.go`, `_handler.go`, `_routes.go`. No service layer needed — handler → store is sufficient for all domains here.

### Database

- sqlc for all queries — no raw SQL strings in handlers or stores
- golang-migrate for migrations: `go run ./cmd/migrate [up|down]`
- `db/generated/` is committed

### Logging

`slog.SetDefault(logger)` in main. JSON to stdout. All packages use `slog` directly.

### Testing

Integration tests via testcontainers — real Postgres, no mocks. `TestMain` handles container lifecycle.

### Conventions

- File naming: `job_handler.go`, `recruiter_store.go`, etc.
- Receiver names: `h` for handlers, `s` for stores
- Errors: log with `slog.Error` server-side, generic message to client
- Routes function: `Routes(store *Store) chi.Router`
- User ID extracted from context via `middleware.UserIDFromContext` — all queries scoped to the authenticated user

---

## Environment Variables

| Variable | Description |
|---|---|
| `PORT` | HTTP port (defaults to 8080) |
| `DATABASE_URL` | Postgres connection string (`postgres://user:pass@host/job_search`) |
| `JWT_PUBLIC_KEY` | RSA public key PEM for validating JWTs issued by go-auth |
| `ALLOWED_ORIGINS` | Comma-separated list of allowed CORS origins |

Copy `.env.example` to `.env.local` for local dev. Never commit `.env.local`.

---

## Current State

- Module renamed to `github.com/Strangebrewer/go-job-search`, all import paths updated
- `.env.example` updated with `job_search` schema name
- Ground zero committed to `main` — template boilerplate intact, no domain code written yet
- **Next**: write `db/schema.sql` and migrations, then `job/` and `recruiter/` domains
