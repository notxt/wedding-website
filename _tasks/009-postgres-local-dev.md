# 009 — Postgres migrations and DB wiring

## Why
Wire Postgres into the server and apply schema migrations on startup. Storage is required before the RSVP store (010) and form (012). The Postgres service itself already exists at `dev/compose.yml` (added in task 002), so this task is about the application-side integration.

## Acceptance criteria
- [ ] `migrations/001_init.sql` creates the `rsvps` table per the schema in `SPEC.md` (uses `CREATE TABLE IF NOT EXISTS` so it's idempotent)
- [ ] `internal/store/store.go` opens a `*pgxpool.Pool` from `DATABASE_URL`, exposes a `Ping(ctx)` method
- [ ] On server startup: open pool → run all `.sql` files in `migrations/` in lexical order → log applied migrations
- [ ] `GET /healthz` now also pings the DB and returns 503 on DB failure
- [ ] `go.mod` adds `github.com/jackc/pgx/v5`; `go.sum` committed
- [ ] Inside the dev container: `go run ./cmd/server` connects to the sibling `postgres` service and applies the schema on a fresh pgdata volume
- [ ] `_tasks/009-postgres-local-dev.md` deleted in the merging PR

## Notes
- Migration runner can be naive: read directory, sort, execute each file. No tracking table yet — schema files use `IF NOT EXISTS`.
- If we outgrow that pattern, switch to tracked migrations in a later task.
- Server should fail fast on DB errors at startup.
- The dev compose file already exposes 5432 to the host, so a host-side `go run` also works using the `localhost:5432` default in `.env.example`.
