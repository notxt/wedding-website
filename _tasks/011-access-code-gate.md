# 011 — Access code gate

## Why
The RSVP section requires a shared code. This task adds the cookie-based gate per the auth design in `SPEC.md`.

## Acceptance criteria
- [ ] `internal/session/cookie.go` provides `Sign(marker string) string` and `Verify(value string) bool` using `crypto/hmac` + `crypto/sha256` keyed on `SESSION_SECRET`
- [ ] `internal/middleware/auth.go` provides `RequireRSVPAuth(next http.Handler) http.Handler` — checks the `rsvp_auth` cookie, redirects to the code-entry form if invalid
- [ ] `internal/templates/rsvp_auth.html` renders the access-code entry form (POSTs to `/rsvp/auth`)
- [ ] `POST /rsvp/auth` compares submitted code to `ACCESS_CODE` in **constant time** (`crypto/subtle.ConstantTimeCompare`); on success sets the signed cookie and redirects to `/rsvp`; on failure re-renders the form with an error
- [ ] Cookie attributes: `HttpOnly`, `SameSite=Strict`, `Path=/rsvp`, `Max-Age=2592000` (30 days). `Secure` set when `r.TLS != nil` or when behind a `Forwarded`/`X-Forwarded-Proto: https` header
- [ ] `GET /rsvp` (unauthenticated) renders the entry form; (authenticated) renders a placeholder "RSVP form goes here" — the real form is task 012
- [ ] `_tasks/011-access-code-gate.md` deleted in the merging PR

## Notes
- The cookie value can be just `Sign("ok")` — there's no per-user identity to encode.
- Don't gate `/rsvp/auth` itself — chicken/egg.
