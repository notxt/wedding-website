# wedding-website

Michael & Adrien's wedding website. Santa Fe, NM — September 25, 2026.

A Go MPA. Most pages public; RSVP gated behind a shared access code.

## Run locally

Recommended: use the dev container.

```bash
./bin/dev.sh
# inside the container:
go run ./cmd/server
```

`bin/dev.sh` is a thin wrapper around `docker compose run --rm --service-ports dev` — the `--service-ports` flag is what lets `http://localhost:8080` from your browser reach the server inside the container.

The container has Go, the AWS CLI, `gh`, and Claude Code preinstalled, plus a sibling Postgres service. See `CLAUDE.md` for the full setup (AWS creds passthrough, GitHub App key location, trust boundary).

Direct host run also works once the container's `postgres` service is up:

```bash
cp .env.example .env
docker compose up -d postgres
go run ./cmd/server
```

## More

- `SPEC.md` — architecture, routes, data model
- `CLAUDE.md` — process, conventions, dev container (read first if you're picking up work)
- `_tasks/` — current backlog; lowest-numbered file is next
