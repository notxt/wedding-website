# 012 — Styling pass

## Why
Site has been functional through tasks 002–011. Time for it to actually look like a wedding website that matches the invite aesthetic.

## Acceptance criteria
- [ ] `static/css/main.css` linked from `layout.html`
- [ ] Visual direction matches the invite SVG: dark charcoal background (`#231f20`), white serif type, generous letter-spacing on display headings
- [ ] Web fonts: **Playfair Display** for headings/display, **Playfair** for body text. Both via Google Fonts (`https://fonts.google.com/specimen/Playfair+Display`, `https://fonts.google.com/specimen/Playfair`). Self-host the woff2 files under `static/fonts/` rather than using a `<link>` to fonts.googleapis.com — keeps the site self-contained and avoids the third-party request.
- [ ] Mobile-first responsive: looks good at 375px wide and at 1280px+
- [ ] Nav becomes a hamburger / stacked menu under ~640px
- [ ] Forms are usable on mobile (44px+ tap targets, sensible input sizing)
- [ ] Focus styles visible for keyboard nav
- [ ] No CSS framework, no build step
- [ ] `_tasks/012-styling-pass.md` deleted in the merging PR

## Notes
- Don't redesign the layout — keep semantic HTML from prior tasks; add CSS only.
- Inline the invite SVG on the home page so its colors can be themed via CSS if needed.
