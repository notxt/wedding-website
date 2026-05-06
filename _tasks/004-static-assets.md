# 004 — Static asset serving

## Why
Make the invite + envelope SVGs (and future CSS) reachable from the server.

## Acceptance criteria
- [ ] `static/img/invitation.svg` and `static/img/envelope-return.svg` copied from `context/`
- [ ] `static/` embedded via `embed.FS`
- [ ] `GET /static/*` serves files via `http.FileServer` over the embedded FS, with `Cache-Control: public, max-age=3600` for development
- [ ] Home page references `/static/img/invitation.svg` (visible on `/`)
- [ ] Unknown paths under `/static/` return 404 (default `FileServer` behavior is fine)
- [ ] `_tasks/004-static-assets.md` deleted in the merging PR

## Notes
- Keep the originals in `context/` untouched — they're reference assets.
- No CSS file yet; that arrives in task 012.
