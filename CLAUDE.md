# CLAUDE.md

Durable process notes for working on this repo. Read this first after a context reset.

## Project

Wedding website for Michael Schoenfelder & Adrien Warner. Wedding is **Friday, September 25, 2026** in Santa Fe, NM (ceremony at the New Mexico Museum of Art, reception at Meow Wolf).

It's a Go MPA. Most pages are public; **only the RSVP section is gated** behind a shared access code (env var, dev default in `.env.example`). Local-first now; AWS deployment via CloudFormation later.

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

```bash
cp .env.example .env       # first time only
docker compose up -d       # starts Postgres
go run ./cmd/server        # starts the web server on $PORT (default 8080)
```

Required environment variables are listed in `.env.example`. Sane dev defaults are baked in for `ACCESS_CODE` and `SESSION_SECRET` so the server runs without a `.env` — but they MUST be overridden in any non-local environment.

## Style

- `gofmt` everything. CI will check.
- Default to no comments. Add one only when the *why* is non-obvious.
- Templates parsed at startup, not per-request.
- Errors: wrap with `fmt.Errorf("...: %w", err)`; let handlers map to HTTP statuses.
- Logging: `log/slog` with structured fields.

## Out of scope (until called for)

- AWS infra / deployment (separate phase, see `_tasks/014-aws-cfn-infra.md`)
- Custom domain, email notifications, photo gallery
- Per-guest accounts (the access-code gate is intentionally a single shared code)
