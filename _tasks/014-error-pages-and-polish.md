# 014 — Error pages and polish

## Why
Round off the local site before AWS work. Catch the rough edges so the site degrades gracefully.

## Acceptance criteria
- [ ] Custom 404 page (uses layout, friendly copy, link back to `/`)
- [ ] Custom 500 page; server errors logged with `slog` but not surfaced to the user (no stack traces in HTML)
- [ ] All routes set sensible response headers: `Content-Type`, `X-Content-Type-Options: nosniff`, `Referrer-Policy: strict-origin-when-cross-origin`
- [ ] Favicon route (use a generic placeholder — task 013 designer can swap later)
- [ ] Trailing-slash handling consistent across routes (decide one way and stick to it)
- [ ] Basic accessibility audit: alt text on images, form labels, heading hierarchy, sufficient contrast
- [ ] Manual smoke test: every route in `SPEC.md` returns expected status both authenticated and unauthenticated where applicable
- [ ] `_tasks/014-error-pages-and-polish.md` deleted in the merging PR

## Notes
- This is the "looks-good locally" gate before task 015 begins AWS work.
