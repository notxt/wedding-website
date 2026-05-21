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
| Local DB       | Postgres 16 via `compose.yml` (sibling of the dev container)    |
| Prod DB        | Amazon RDS Postgres (later phase)                               |
| Migrations     | Plain `.sql` files in `migrations/`, applied idempotently on server start |
| Frontend       | Plain HTML + plain CSS, no JS framework, no build step          |
| Fonts          | Playfair Display (headings) + Playfair (body), self-hosted woff2 |
| Deployment     | CloudFormation → ALB + EC2 ASG (1/1/1) + CodeDeploy + RDS       |

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
├── compose.yml                 # dev + postgres services — task 002
├── dev/                        # dev container Dockerfile + scripts — task 002
├── go.mod
├── go.sum                      # appears once we add pgx
├── _tasks/                     # numbered task files; deleted on PR merge
├── context/                    # original assets (invite SVG, envelope SVG, copy doc)
├── cmd/
│   └── server/main.go          # added in task 003
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

## Production deployment

Four CFN stacks in `infra/`, deployed in this order (see `infra/README.md`):

1. **`wedding-bootstrap`** — S3 bucket for CFN artifacts and CodeDeploy revision bundles.
2. **`wedding-network`** — VPC `10.0.0.0/16`, two public + two private subnets across us-west-2a/b, IGW, route tables. No NAT (EC2 lives in public subnets with SG-locked ingress).
3. **`wedding-data`** — RDS Postgres (`db.t4g.micro`, single-AZ, encrypted, deletion-protected, 7-day backups). Master credentials managed by RDS in Secrets Manager. App-level `ACCESS_CODE` and `SESSION_SECRET` secrets created outside CFN by the deploy script.
4. **`wedding-app`** — ACM cert (DNS-validated, apex + www), ALB with HTTPS→target-group + HTTP→301 listeners, target group with `/healthz` health check, Route53 alias records, EC2 launch template (AL2023 arm64 `t4g.micro`, instance role with Secrets Manager + CW Logs + SSM + S3 read on the bootstrap bucket), ASG `1/1/1` self-healing, CodeDeploy (IN_PLACE, AllAtOnce).

Instance userdata installs the CodeDeploy + CloudWatch agents, fetches secrets, and writes `/etc/wedding-website/env` with `DATABASE_URL` (including `sslmode=require`), `ACCESS_CODE`, `SESSION_SECRET`, and `PORT=8080`. The Go binary + systemd unit + lifecycle hook scripts ship via CodeDeploy out of `deploy/` in this repo (`bin/deploy-app.sh` builds, bundles, uploads to S3, and creates a deployment).

## Out of scope (for now)

- Email notifications on RSVP submit
- Photo gallery, the proposal page (copy doc marks these as TBD)
- Itinerary content, things-to-do content (copy doc has these as stubs)
- Per-guest authentication, edit-your-RSVP flow
- CloudWatch alarms / WAF / blue-green CodeDeploy / multi-AZ RDS
