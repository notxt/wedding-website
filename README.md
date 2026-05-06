# wedding-website

Michael & Adrien's wedding website. Santa Fe, NM — September 25, 2026.

A Go MPA. Most pages public; RSVP gated behind a shared access code.

## Run locally

```bash
cp .env.example .env       # first time
docker compose up -d       # Postgres
go run ./cmd/server
```

Server listens on `$PORT` (default 8080).

## More

- `SPEC.md` — architecture, routes, data model
- `CLAUDE.md` — process and conventions (read first if you're picking up work)
- `_tasks/` — current backlog; lowest-numbered file is next
