# 012 — Styling pass

## Why
Site has been functional through tasks 002–011. Time for it to actually look like a wedding website that matches the invite aesthetic.

## Acceptance criteria
- [ ] `static/css/main.css` linked from `layout.html`
- [ ] Visual direction matches the invite SVG: dark charcoal background (`#231f20`), white serif type, generous letter-spacing on display headings
- [ ] Web font: choose a public-domain or Google-Fonts serif that approximates the invite (e.g. Cormorant Garamond, EB Garamond, or system serif fallback). Self-host if Google Fonts; document choice in `SPEC.md`.
- [ ] Mobile-first responsive: looks good at 375px wide and at 1280px+
- [ ] Nav becomes a hamburger / stacked menu under ~640px
- [ ] Forms are usable on mobile (44px+ tap targets, sensible input sizing)
- [ ] Focus styles visible for keyboard nav
- [ ] No CSS framework, no build step
- [ ] `_tasks/012-styling-pass.md` deleted in the merging PR

## Notes
- Don't redesign the layout — keep semantic HTML from prior tasks; add CSS only.
- Inline the invite SVG on the home page so its colors can be themed via CSS if needed.
