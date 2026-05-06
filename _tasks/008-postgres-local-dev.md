# 008 — Postgres for local dev

## Why
Stand up Postgres in Docker, wire it into the server, and apply schema migrations on startup. Storage is required before the RSVP store (009) and form (011).

## Acceptance criteria
- [ ] `docker-compose.yml` defines a `postgres:16` service with named volume, exposing `5432`, with creds matching the `DATABASE_URL` default in `.env.example`
- [ ] `migrations/001_init.sql` creates the `rsvps` table per the schema in `SPEC.md` (uses `CREATE TABLE IF NOT EXISTS` so it's idempotent)
- [ ] `internal/store/store.go` opens a `*pgxpool.Pool` from `DATABASE_URL`, exposes a `Ping(ctx)` method
- [ ] On server startup: open pool → run all `.sql` files in `migrations/` in lexical order → log applied migrations
- [ ] `GET /healthz` now also pings the DB and returns 503 on DB failure
- [ ] `go.mod` adds `github.com/jackc/pgx/v5`; `go.sum` committed
- [ ] `docker compose up -d && go run ./cmd/server` works end-to-end on a clean machine
- [ ] `_tasks/008-postgres-local-dev.md` deleted in the merging PR

## Notes
- Migration runner can be naive: read directory, sort, execute each file. No tracking table yet — schema files use `IF NOT EXISTS`.
- If we outgrow that pattern, switch to tracked migrations in a later task.
- Server should fail fast on DB errors at startup.
