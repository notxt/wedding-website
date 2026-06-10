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
| GET    | `/rsvp`                       | **gated** | If unauthenticated → email-entry form; else the guest's RSVP form |
| POST   | `/rsvp/auth`                  | public    | Match email against the guest list, set cookie, redirect to `/rsvp` |
| POST   | `/rsvp`                       | **gated** | Submit RSVP                                         |
| GET    | `/rsvp/thanks`                | gated     | Confirmation                                        |
| GET    | `/healthz`                    | public    | Liveness probe (DB ping included)                   |
| GET    | `/static/*`                   | public    | CSS, SVGs                                           |

## Auth design

The RSVP gate is **membership in the `guests` invite list**, keyed by email — there's no
shared code and no per-guest password. Low sensitivity (same trust level as the old shared
code): knowing an invited email is the "credential". The signed cookie then carries the
guest's identity, so the form is personalized and the RSVP is linked to the right guest.

Flow:

1. Visitor hits `/rsvp` without a valid auth cookie → server renders an email-entry form.
2. Visitor submits their email to `POST /rsvp/auth`.
3. Server normalizes (lower-case, trim) and looks the email up in `guests`. On a match, sets a cookie:
   - Name: `rsvp_auth`
   - Value: HMAC-SHA256 of the guest's **id** (an opaque marker, no PII), signed with `SESSION_SECRET`
   - Attributes: `HttpOnly`, `SameSite=Strict`, `Secure` (in prod), `Path=/rsvp`, `Max-Age` ~ 30 days
   - No match → re-render the form with a "we can't find that email — get in touch" message; no cookie.
4. Middleware on `POST /rsvp` and `/rsvp/thanks` verifies the cookie's HMAC; handlers parse the
   guest id from the verified marker to load the guest and link/prefill the RSVP. Bad/missing → back to the form.
5. `/rsvp/auth` itself is **not** gated (otherwise nobody could authenticate).

CSRF: `SameSite=Strict` cookies plus same-origin form posts are sufficient for this threat model. No additional CSRF token framework.

## Data model

```sql
-- The invite list = the RSVP gate. Email is normalized (lower-cased, trimmed).
-- Populated out-of-band from a gitignored seed (see below); not in committed migrations.
CREATE TABLE IF NOT EXISTS guests (
  id               BIGSERIAL PRIMARY KEY,
  email            TEXT        NOT NULL UNIQUE,
  first_name       TEXT        NOT NULL,
  last_name        TEXT        NOT NULL,
  plus_one_allowed BOOLEAN     NOT NULL DEFAULT false,
  plus_one_name    TEXT,                     -- known +1 name from the list, used as a default
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS rsvps (
  id                   BIGSERIAL PRIMARY KEY,
  submitted_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
  guest_id             BIGINT      REFERENCES guests(id),  -- UNIQUE: one RSVP per guest (upsert)
  attending            BOOLEAN     NOT NULL,
  party_names          TEXT        NOT NULL,               -- derived display string ("First Last [+ Plus1]")
  meal_choice          TEXT,                               -- 'chicken' | 'vegetarian' | NULL when not attending
  dairy_allergy        BOOLEAN     NOT NULL DEFAULT false,
  gluten_allergy       BOOLEAN     NOT NULL DEFAULT false,
  other_allergies      TEXT,
  plus_one_attending   BOOLEAN     NOT NULL DEFAULT false,
  plus_one_name        TEXT,
  plus_one_meal_choice TEXT,                               -- 'chicken' | 'vegetarian'
  remote_addr          TEXT                                -- audit / dedupe hint
);
```

The primary guest is identified by their list entry (name comes from `guests`, not a form
field). `meal_choice`, allergies, and the plus-one block are only collected when
`attending = true`; plus-one fields only when the guest is `plus_one_allowed`. A returning
guest's submission **upserts** on `guest_id`.

**Seeding the guest list:** the list lives in a gitignored spreadsheet; a one-time
`untracked/seed_guests.sql` (generated from it, `ON CONFLICT (email) DO UPDATE`) is applied
with `psql` — locally for dev, against RDS for prod. No PII in committed migrations. NOTE:
`go test ./internal/store/...` truncates `guests`/`rsvps` on the shared dev DB — re-run the
seed after testing.

## Environment variables

| Name              | Default (dev)                                                          | Required in prod | Purpose |
|-------------------|------------------------------------------------------------------------|------------------|---------|
| `PORT`            | `8080`                                                                 | no               | HTTP listen port |
| `DATABASE_URL`    | `postgres://wedding:wedding@localhost:5432/wedding?sslmode=disable`    | yes              | Postgres connection string |
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
│   ├── 001_init.sql
│   └── 002_guests_and_rsvp_link.sql
└── static/
    ├── css/
    └── img/
```

## Production deployment

Four CFN stacks in `infra/`, deployed in this order (see `infra/README.md`):

1. **`wedding-bootstrap`** — S3 bucket for CFN artifacts and CodeDeploy revision bundles.
2. **`wedding-network`** — VPC `10.0.0.0/16`, two public + two private subnets across us-west-2a/b, IGW, route tables. No NAT (EC2 lives in public subnets with SG-locked ingress).
3. **`wedding-data`** — RDS Postgres (`db.t4g.micro`, single-AZ, encrypted, deletion-protected, 7-day backups). Master credentials are a **self-managed** Secrets Manager secret defined in CFN (`DbMasterSecret`) with `GenerateSecretString` and **no rotation schedule** — see "Master secret rotation policy" below. App-level `SESSION_SECRET` and `ACCESS_CODE` secrets are created outside CFN by the deploy script. (The guest list is seeded into Postgres via `untracked/seed_guests.sql` — see Data model.)

### Master secret rotation policy

The DB master credentials are **deliberately not rotated**. The userdata bakes the password into `/etc/wedding-website/env` at boot, so any rotation event invalidates the running instance's credentials and causes an outage. This happened in production on 2026-06-09.

The threat model that justifies "no rotation":

- The DB is in a private subnet with no public access (`PubliclyAccessible: false`).
- The only path in is via the app SG (locked to the ALB) or an SSM port-forward through the app EC2 instance (IAM-gated, audited via CloudTrail).
- The master secret is read only at instance boot (userdata) and by `bin/db-tunnel.sh` (operator workflow, IAM-gated).
- There is no third-party integration that could leak the secret, no broad IAM read access, no log emission of the secret value.

If rotation ever becomes necessary, the correct fix is **runtime credential fetching** in the Go app (pgxpool `BeforeConnect` hook that pulls from Secrets Manager with a short TTL + force-refresh on auth failure), not flipping the secret to RDS-managed.
4. **`wedding-app`** — ACM cert (DNS-validated, apex + www), ALB with HTTPS→target-group + HTTP→301 listeners, target group with `/healthz` health check, Route53 alias records, EC2 launch template (AL2023 arm64 `t4g.micro`, instance role with Secrets Manager + CW Logs + SSM + S3 read on the bootstrap bucket), ASG `1/1/1` self-healing, CodeDeploy (IN_PLACE, AllAtOnce).

Instance userdata installs the CodeDeploy + CloudWatch agents, fetches secrets, and writes `/etc/wedding-website/env` with `DATABASE_URL` (including `sslmode=require`), `SESSION_SECRET`, and `PORT=8080`. The Go binary + systemd unit + lifecycle hook scripts ship via CodeDeploy out of `deploy/` in this repo (`bin/deploy-app.sh` builds, bundles, uploads to S3, and creates a deployment).

## Out of scope (for now)

- Email notifications on RSVP submit
- Photo gallery, the proposal page (copy doc marks these as TBD)
- Itinerary content, things-to-do content (copy doc has these as stubs)
- Per-guest authentication, edit-your-RSVP flow
- CloudWatch alarms / WAF / blue-green CodeDeploy / multi-AZ RDS
