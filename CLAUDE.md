# CLAUDE.md

Durable process notes for working on this repo. Read this first after a context reset.

## Project

Wedding website for Michael Schoenfelder & Adrien Warner. Wedding is **Friday, September 25, 2026** in Santa Fe, NM (ceremony at the New Mexico Museum of Art, reception at Meow Wolf).

It's a Go MPA. Most pages are public; **only the RSVP section is gated** — by matching a guest's email against the `guests` invite list (seeded from a gitignored spreadsheet via `untracked/seed_guests.sql`; the signed cookie carries the guest id). Local-first now; AWS deployment via CloudFormation later.

See `SPEC.md` for the architecture, route table, and data model.

## Task management protocol

All planned work lives in `_tasks/NNN-<slug>.md`. Each file is one PR-sized unit of work.

**Strict rules:**

1. **One task at a time, in numeric order.** Do not start task N+1 until task N's PR is merged into `main`. No parallel branches, no stacked PRs.
2. **One task = one PR.** PR title matches the task title. Branch name matches the task slug (e.g. `002-base-server-and-config`).
3. **Delete the task file in the same PR that completes it.** The merge that ships the work also removes its task file from `_tasks/`.
4. **Keep PRs under 20 files when at all possible.** If a task naturally exceeds that, split it before starting: add an interleaved task file (e.g. `008a-...`, `008b-...`) and re-number anything downstream so the order is still strict. Better to split work than to ship a sprawling PR.

**Picking up work (e.g., after a context reset):**

```bash
git status                 # working tree should be clean
git branch --show-current  # should be on main (or the open feature branch)
git fetch && git status    # confirm in sync with origin
ls _tasks/                 # lowest-numbered file is the next task
```

If a feature branch is checked out with an open PR, finish and merge that PR before starting anything new.

**Task file format:**

```markdown
# NNN — Title

## Why
1–2 lines on the motivation.

## Acceptance criteria
- [ ] Concrete, verifiable item
- [ ] Another concrete item

## Notes
Any links, gotchas, or design hints.
```

## Dependency policy

Go standard library first. Part of the reason we chose Go is the strength of its stdlib — leverage it.

Adding a new external dependency requires updating `SPEC.md` with a justification. Currently sanctioned exceptions:

- `github.com/jackc/pgx/v5` — Postgres driver. No realistic stdlib alternative.

Front-end has no build step: plain HTML, plain CSS, no JS framework. Add JS only if a feature genuinely needs it.

## Local development

The recommended workflow is the **dev container** (see below). Direct host runs still work for poking around:

```bash
cp .env.example .env       # first time only
go run ./cmd/server        # starts the web server on $PORT (default 8080)
```

Required environment variables are listed in `.env.example`. Sane dev defaults are baked in for `ACCESS_CODE` and `SESSION_SECRET` so the server runs without a `.env` — but they MUST be overridden in any non-local environment.

## Dev container

A containerized dev environment lives in `dev/`. Run inside it instead of installing Go / AWS CLI / `gh` / Claude Code on the host. It also makes `claude --dangerously-skip-permissions` safe by isolating the filesystem.

**Run it:**

```bash
./bin/dev.sh
```

This is a thin wrapper around `docker compose run --rm --service-ports dev`. The `--service-ports` flag is what makes `localhost:8080` from the host browser reach `go run ./cmd/server` inside the container — plain `docker compose run` ignores `compose.yml`'s `ports:` block.

It drops you into a bash shell at `/workspace` (the repo, bind-mounted), with a sibling `postgres` service reachable at `postgres:5432`.

**AWS credentials**: authenticate on the host (`aws login`); the container's own AWS CLI then resolves the profile and mints short-lived creds from the cached login session, refreshing them itself as they expire. No env vars to wrangle, and the region rides along from `~/.aws/config`. When the login session itself expires, re-run `aws login` on the host. Mechanically, `~/.aws` is mounted in two pieces: `~/.aws/config` **read-only** (the host's config can't be tampered with from inside the container), and `~/.aws/login` **read-write** (shared live with the host). The read-write `login` mount is required, not incidental: an `aws login` session rotates its refresh token on every refresh and rewrites `~/.aws/login/cache`, so host and container *must* share the one cache file — a read-only mount breaks every `aws` call, and a snapshot/copy would invalidate the host's login the first time either side refreshes.

**GitHub App key** is mounted read-only from `${CLAUDE_GH_APP_DIR:-../claude_gh_app}` (relative to the repo root). The container fetches a fresh installation token on every `git push` via a credential helper — no token wrangling needed.

**Container detection:** `WEDDING_DEV_CONTAINER=1` is set inside the container. Scripts can branch on it: `[ -n "${WEDDING_DEV_CONTAINER:-}" ]`.

**Database access (prod):** the prod RDS is private (no public access, no Data API — it's a plain RDS instance, not Aurora), so reach it via an SSM port-forward through the app EC2 instance. `bin/db-tunnel.sh` does this end-to-end — it resolves the instance + DB facts, opens the tunnel, and runs `psql`:

```bash
./bin/db-tunnel.sh -c "select count(*) from rsvps;"   # one-shot, prints to stdout
./bin/db-tunnel.sh -f report.sql                       # run a file
echo "select * from guests;" | ./bin/db-tunnel.sh      # SQL from stdin
./bin/db-tunnel.sh                                     # interactive psql (on a TTY)
```

It's **read-only by default** (so an agent can't accidentally mutate the guest list); pass `--write` to allow writes. Notes:
- The container ships `session-manager-plugin`, but it only lands after a **host image rebuild** (`docker compose build dev`, or the next `./bin/dev.sh`). Until then the script errors with a hint.
- The connecting AWS identity needs `ssm:StartSession` on the instance + the `AWS-StartPortForwardingSessionToRemoteHost` document, and `secretsmanager:GetSecretValue` on the RDS master secret. Creds come from that master secret today; the script isolates credential fetching so a later move to RDS IAM auth is a one-function change.

**Trust boundary:** the only host paths visible inside the container are:
- the repo (read-write — Claude needs to edit it)
- `~/.claude` and `~/.claude.json` (read-write — preserves your Claude Code login + settings)
- the GH App key dir (read-only)
- `~/.aws/config` (read-only — the container's AWS CLI reads the profile but can't inject a `credential_process` into the host's config) and `~/.aws/login` (read-write — the shared, rotating credential cache the CLI must rewrite to refresh)

No docker socket. No host network. Runs as a non-root `dev` user. This is the boundary that makes `--dangerously-skip-permissions` acceptable.

## Style

- `gofmt` everything. CI will check.
- Default to no comments. Add one only when the *why* is non-obvious.
- Templates parsed at startup, not per-request.
- Errors: wrap with `fmt.Errorf("...: %w", err)`; let handlers map to HTTP statuses.
- Logging: `log/slog` with structured fields.

## Out of scope (until called for)

- AWS infra / deployment (separate phase, see `_tasks/015-aws-cfn-infra.md`)
- Custom domain, email notifications, photo gallery
- Per-guest accounts (the access-code gate is intentionally a single shared code)
