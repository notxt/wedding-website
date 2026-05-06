# wedding-website

Michael & Adrien's wedding website. Santa Fe, NM — September 25, 2026.

A Go MPA. Most pages public; RSVP gated behind a shared access code.

## Run locally

Recommended: use the dev container.

```bash
docker compose -f dev/compose.yml run --rm dev
# inside the container:
go run ./cmd/server
```

The container has Go, the AWS CLI, `gh`, and Claude Code preinstalled, plus a sibling Postgres service. See `CLAUDE.md` for the full setup (AWS creds passthrough, GitHub App key location, trust boundary).

Direct host run also works once the container's `postgres` service is up:

```bash
cp .env.example .env
docker compose -f dev/compose.yml up -d postgres
go run ./cmd/server
```

## More

- `SPEC.md` — architecture, routes, data model
- `CLAUDE.md` — process, conventions, dev container (read first if you're picking up work)
- `_tasks/` — current backlog; lowest-numbered file is next
