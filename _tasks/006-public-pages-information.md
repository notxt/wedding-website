# 006 — Information section

## Why
Ship the Information index and its three sub-pages. Travel has full copy; itinerary and things-to-do are placeholders for now (copy doc marks them as TBD).

## Acceptance criteria
- [ ] `GET /information` renders an index linking to the three sub-pages
- [ ] `GET /information/itinerary` renders a placeholder ("Details coming soon")
- [ ] `GET /information/travel` renders the verbatim travel + accommodations copy from the copy doc, including the altitude warning, airport options, Rail Runner notes, and the $70 shuttle reference
- [ ] `GET /information/things-to-do` renders a placeholder ("Details coming soon")
- [ ] External links (RioMetro.org, ABQ airport, etc.) open in a new tab with `rel="noopener"`
- [ ] `_tasks/006-public-pages-information.md` deleted in the merging PR

## Notes
- The `here` link references in the travel copy point to URLs not provided in the doc — link them as plain text "here" with no `href` for now, or put `href="#"` and a TODO comment. Don't invent URLs.
