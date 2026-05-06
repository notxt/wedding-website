# 011 — RSVP form and submit

## Why
Wire up the actual RSVP form (gated by 010) and the POST handler that persists it via the store from 009.

## Acceptance criteria
- [ ] `internal/templates/rsvp.html` renders the form with the fields from the copy doc:
  - Will you be attending? (radio: Yes / No, required)
  - Names of party (textarea, required)
  - Meal preference (radio: Chicken / Vegetarian, shown only when attending=Yes)
  - Dairy allergy (radio: Yes / No)
  - Gluten allergy (radio: Yes / No)
  - Other allergies (textarea, optional)
- [ ] `GET /rsvp` (authenticated) renders the form
- [ ] `POST /rsvp` parses the form, validates server-side, calls `store.InsertRSVP`, redirects to `/rsvp/thanks` on success
- [ ] `GET /rsvp/thanks` renders a confirmation page
- [ ] On validation failure: re-render the form with the user's previous values + error messages
- [ ] If `attending=false`, meal/allergy fields are ignored (set to nil/false in the row)
- [ ] `remote_addr` populated from `r.RemoteAddr` (or `X-Forwarded-For` first hop if set)
- [ ] Conditional showing of meal-preference fields can be CSS-only (`:has()`) or progressive — no JS required to submit a valid form
- [ ] `_tasks/011-rsvp-form-and-submit.md` deleted in the merging PR

## Notes
- Keep validation in a small `validate(form url.Values) (RSVP, errs)` helper — testable without HTTP.
- "Names of party" is free text; trim whitespace, reject empty.
