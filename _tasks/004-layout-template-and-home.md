# 004 — Layout template and home page

## Why
Establish the templating pattern (base layout + page) and ship the first real route. Every later page extends this.

## Acceptance criteria
- [ ] `internal/templates/layout.html` defines `{{ block "title" }}`, `{{ block "content" }}`, and a shared nav matching the menu in the copy doc
- [ ] `internal/templates/home.html` renders inside the layout
- [ ] Templates loaded once at startup via `embed.FS` and `template.ParseFS`; no per-request reparsing
- [ ] `GET /` renders the home page and returns 200
- [ ] Helper for rendering: takes a template name + data, writes to `http.ResponseWriter`, surfaces errors as 500
- [ ] `_tasks/004-layout-template-and-home.md` deleted in the merging PR

## Notes
- Keep the home page minimal — task 013 (styling pass) is where the invite SVG hero gets its real treatment. For now, a heading + names + date is fine.
- Nav links can point to routes that don't exist yet; that's OK, they'll be added in subsequent tasks.
