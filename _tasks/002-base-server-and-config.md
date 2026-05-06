# 002 — Base server and config

## Why
Stand up the smallest runnable Go web server so every later task has somewhere to plug in. No business logic yet — just process bones.

## Acceptance criteria
- [ ] `cmd/server/main.go` boots an `http.Server` on `$PORT`
- [ ] `internal/config/config.go` loads `PORT`, `DATABASE_URL`, `ACCESS_CODE`, `SESSION_SECRET` from env with the documented dev defaults
- [ ] Structured logging via `log/slog` (JSON handler), one startup line listing the bound address
- [ ] `GET /healthz` returns `200 OK` with body `ok`
- [ ] Graceful shutdown on SIGINT/SIGTERM with a 10s drain timeout
- [ ] `go build ./...` and `go vet ./...` pass
- [ ] `_tasks/002-base-server-and-config.md` deleted in the merging PR

## Notes
- Use `net/http.ServeMux` (Go 1.22+ patterns). No router dependency.
- Don't load `.env` from disk — `os.Getenv` only. Operators set the vars themselves.
- Healthz must NOT depend on Postgres yet (DB is wired up in task 008).
