# go-job-search

Job search tracking service for the personal-enterprise project. Manages job applications and recruiters.

> Active development — endpoints and schema documented once stable.

---

## Stack

- **Language**: Go
- **Router**: [chi](https://github.com/go-chi/chi)
- **Database**: MongoDB
- **Auth**: Stateless RSA JWT validation — tokens issued by go-auth, verified independently here
- **Logging**: `slog` with JSON output

---

## Running Locally

Copy `.env.example` to `.env.local` and fill in values.

```bash
# Start the server
go run ./cmd/server

# Run tests
go test ./...
```

---

## Environment Variables

| Variable          | Description                                              |
| ----------------- | -------------------------------------------------------- |
| `PORT`            | HTTP port (defaults to 8080)                             |
| `DATABASE_URL`    | MongoDB connection string                                |
| `JWT_PUBLIC_KEY`  | RSA public key PEM for validating JWTs issued by go-auth |
| `ALLOWED_ORIGINS` | Comma-separated list of allowed CORS origins             |
