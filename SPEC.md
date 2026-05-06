# SPEC

Architecture and design reference for the wedding website. Process and conventions live in `CLAUDE.md`.

## Tech stack

| Concern        | Choice                                                          |
|----------------|-----------------------------------------------------------------|
| Language       | Go 1.22+ (uses new `net/http.ServeMux` method+path patterns)    |
| HTTP           | `net/http` stdlib                                               |
| Templating     | `html/template` stdlib                                          |
| Logging        | `log/slog` stdlib                                               |
| Sessions       | Signed cookies via `crypto/hmac` + `crypto/sha256` (stdlib)     |
| Static assets  | `embed.FS` served via `http.FileServer`                         |
| DB driver      | `github.com/jackc/pgx/v5` (only sanctioned external dep)        |
| Local DB       | Postgres 16 via `docker-compose.yml`                            |
| Prod DB        | Amazon RDS Postgres (later phase)                               |
| Migrations     | Plain `.sql` files in `migrations/`, applied idempotently on server start |
| Frontend       | Plain HTML + plain CSS, no JS framework, no build step          |
| Fonts          | Playfair Display (headings) + Playfair (body), self-hosted woff2 |
| Deployment     | CloudFormation → ECS Fargate (or App Runner) + RDS (later phase)|

## Routes

| Method | Path                          | Auth      | Purpose                                             |
|--------|-------------------------------|-----------|-----------------------------------------------------|
| GET    | `/`                           | public    | Home — invite SVG hero                              |
| GET    | `/faqs`                       | public    | FAQs                                                |
| GET    | `/information`                | public    | Information index                                   |
| GET    | `/information/itinerary`      | public    | Wedding itinerary                                   |
| GET    | `/information/travel`         | public    | Travel + accommodations                             |
| GET    | `/information/things-to-do`   | public    | Things to do in Santa Fe                            |
| GET    | `/contact`                    | public    | Contact index                                       |
| GET    | `/contact/about-us`           | public    | About us                                            |
| GET    | `/contact/registry`           | public    | Registry                                            |
| GET    | `/contact/get-in-touch`       | public    | Get in touch                                        |
| GET    | `/rsvp`                       | **gated** | If unauthenticated → access-code form; else RSVP form |
| POST   | `/rsvp/auth`                  | public    | Verify access code, set cookie, redirect to `/rsvp` |
| POST   | `/rsvp`                       | **gated** | Submit RSVP                                         |
| GET    | `/rsvp/thanks`                | gated     | Confirmation                                        |
| GET    | `/healthz`                    | public    | Liveness probe (DB ping included)                   |
| GET    | `/static/*`                   | public    | CSS, SVGs                                           |

## Auth design

Single shared `ACCESS_CODE` (env var, dev default `letmein`). This is a **logistics gate, not identity** — there are no per-guest accounts.

Flow:

1. Visitor hits `/rsvp` without a valid auth cookie → server renders an access-code entry form.
2. Visitor submits the code to `POST /rsvp/auth`.
3. Server compares (constant-time) against `ACCESS_CODE`. On match, sets a cookie:
   - Name: `rsvp_auth`
   - Value: HMAC-SHA256 of a fixed marker, signed with `SESSION_SECRET`
   - Attributes: `HttpOnly`, `SameSite=Strict`, `Secure` (in prod), `Path=/rsvp`, `Max-Age` ~ 30 days
4. Middleware on `/rsvp` and `POST /rsvp` verifies the cookie's HMAC. Bad/missing → re-render the entry form.
5. `/rsvp/auth` itself is **not** gated (otherwise nobody could authenticate).

CSRF: `SameSite=Strict` cookies plus same-origin form posts are sufficient for this threat model. No additional CSRF token framework.

## Data model

```sql
CREATE TABLE IF NOT EXISTS rsvps (
  id              BIGSERIAL PRIMARY KEY,
  submitted_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  attending       BOOLEAN     NOT NULL,
  party_names     TEXT        NOT NULL,
  meal_choice     TEXT,                     -- 'chicken' | 'vegetarian' | NULL when not attending
  dairy_allergy   BOOLEAN     NOT NULL DEFAULT false,
  gluten_allergy  BOOLEAN     NOT NULL DEFAULT false,
  other_allergies TEXT,
  remote_addr     TEXT                      -- audit / dedupe hint
);
```

Form fields map directly to columns. `meal_choice`, `dairy_allergy`, `gluten_allergy`, `other_allergies` are only collected when `attending = true`.

## Environment variables

| Name              | Default (dev)                                                          | Required in prod | Purpose |
|-------------------|------------------------------------------------------------------------|------------------|---------|
| `PORT`            | `8080`                                                                 | no               | HTTP listen port |
| `DATABASE_URL`    | `postgres://wedding:wedding@localhost:5432/wedding?sslmode=disable`    | yes              | Postgres connection string |
| `ACCESS_CODE`     | `letmein`                                                              | **yes — override** | Shared RSVP access code |
| `SESSION_SECRET`  | `dev-secret-do-not-use-in-prod`                                        | **yes — override** | HMAC key for the auth cookie |

## Project layout (target)

```
wedding-website/
├── CLAUDE.md
├── SPEC.md
├── README.md
├── .env.example
├── .gitignore
├── docker-compose.yml          # added in task 008
├── go.mod
├── go.sum                      # appears once we add pgx
├── _tasks/                     # numbered task files; deleted on PR merge
├── context/                    # original assets (invite SVG, envelope SVG, copy doc)
├── cmd/
│   └── server/main.go          # added in task 002
├── internal/
│   ├── config/                 # env var loading
│   ├── handlers/               # one file per route group
│   ├── middleware/             # auth, logging
│   ├── session/                # signed cookie helpers
│   ├── store/                  # Postgres-backed RSVP repo
│   └── templates/              # *.html, embedded
├── migrations/
│   └── 001_init.sql
└── static/
    ├── css/
    └── img/
```

## Out of scope (for now)

- AWS infrastructure (separate phase — task 014)
- Custom domain, TLS termination
- Email notifications on RSVP submit
- Photo gallery, the proposal page (copy doc marks these as TBD)
- Itinerary content, things-to-do content (copy doc has these as stubs)
- Per-guest authentication, edit-your-RSVP flow
