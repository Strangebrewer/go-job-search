# go-job-search — Claude Context

## What This Service Is

The job search tracking service for the personal-enterprise project. Manages job applications and recruiters. Backed by MongoDB Atlas. Validates JWTs issued by go-auth — does not issue tokens.

Built from `go-service-template`. The structure, patterns, and tooling are inherited from that template — this file documents what is specific to go-job-search on top of that foundation.

---

## Architecture

```
cmd/
  server/main.go     ← wiring: config, DB, stores, server.New()
app/
  app.go             ← Application struct: JobStore, RecruiterStore
server/
  server.go          ← chi router, global middleware
  routes.go          ← route registration — all routes auth-protected
config/
  config.go          ← standard template config, no additions needed
db_connection/
  db.go              ← MongoDB Connect(), creates indexes for jobs + recruiters
health/
  handler.go
middleware/
  auth.go
  logging.go
  requestid.go
job/
  job_model.go       ← Job domain type + request/filter types
  job_store.go       ← jobDoc (bson), Store, CRUD + dynamic filter for List
  job_handler.go
  job_routes.go
recruiter/
  recruiter_model.go ← Recruiter domain type + request types
  recruiter_store.go ← recruiterDoc (bson), Store, CRUD; Delete checks jobs collection
  recruiter_handler.go
  recruiter_routes.go
```

---

## All Routes Are Protected

Every domain in this service requires authentication. `authMiddleware` is applied to all mounts in `server/routes.go` — no unprotected endpoints except `/health`.

---

## Patterns Carried Over from Template

### Domain Structure

Four-file pattern: `<domain>_model.go`, `_store.go`, `_handler.go`, `_routes.go`. No service layer needed — handler → store is sufficient for all domains here.

### Database

- MongoDB Atlas via mongo-driver v2; no ORM; no migrations
- `db_connection.Connect()` returns `(*mongo.Client, *mongo.Database)`; indexes created at startup
- Store pattern: private `<domain>Doc` struct with `bson` tags in `_store.go`; exported domain type with `json` tags in `_model.go`; `toDomain()` converts between them
- IDs stored as UUID v7 strings (`uuid.NewV7().String()`)
- Recruiter existence validated in `job.Store.Create` by counting documents in the recruiters collection (existence only — not ownership, matching the original FK constraint behavior)
- `recruiter.Store.Delete` checks the jobs collection before removing to enforce `ErrHasJobs`

### Logging

`slog.SetDefault(logger)` in main. JSON to stdout. All packages use `slog` directly.

### Testing

Integration tests via testcontainers — real MongoDB (`mongo:6`), no mocks. `TestMain` handles container lifecycle. `recruiter_test.go` holds a `testJobStore` to set up FK-equivalent state for `Delete_BlockedByJobs`.

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
| `DATABASE_URL` | MongoDB Atlas URI (`mongodb+srv://user:pass@cluster.mongodb.net/`) — database name `job_search` is hardcoded in `db_connection` |
| `JWT_PUBLIC_KEY` | RSA public key PEM for validating JWTs issued by go-auth |
| `ALLOWED_ORIGINS` | Comma-separated list of allowed CORS origins |
| `TRACER_SERVICE_URL` | go-tracer service URL (optional) |
| `TRACER_SERVICE_KEY` | go-tracer auth key (optional) |

Copy `.env.example` to `.env.local` for local dev. Never commit `.env.local`.

---

## Current State

- Migrated from Postgres (sqlc + golang-migrate) to MongoDB Atlas (mongo-driver v2)
- `job/` and `recruiter/` domains complete — full CRUD, integration tests passing
- Tracing wired on all `job/` and `recruiter/` endpoints (all handlers emit spans on success and error paths)
- Deployed to dev: `https://go-job-search-dev-213672305641.us-central1.run.app`
- **GCP TODO**: update `db-url-job-search` secret in Secret Manager to MongoDB Atlas URI; remove Cloud SQL attachment from Cloud Run service
