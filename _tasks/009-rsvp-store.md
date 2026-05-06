# 009 — RSVP store

## Why
Encapsulate RSVP persistence behind a small typed API so the form handler in 011 stays focused on HTTP/validation.

## Acceptance criteria
- [ ] `internal/store/rsvp.go` defines an `RSVP` struct mirroring the table schema
- [ ] `(s *Store) InsertRSVP(ctx, RSVP) (id int64, err error)` inserts and returns the new row id
- [ ] `(s *Store) ListRSVPs(ctx) ([]RSVP, error)` returns all rows ordered by `submitted_at DESC` (used later for an admin view; ship it now)
- [ ] Table-driven test in `internal/store/rsvp_test.go` against the docker'd Postgres; skip with `t.Skip` if `DATABASE_URL` unset or unreachable
- [ ] `go test ./internal/store/...` passes against `docker compose up -d`
- [ ] `_tasks/009-rsvp-store.md` deleted in the merging PR

## Notes
- Use parameterized queries (`pgx` named args or `$1` placeholders). No string interpolation of user input.
- Keep `RSVP.MealChoice`, `RSVP.OtherAllergies` as `*string` (nullable) to match the schema.
