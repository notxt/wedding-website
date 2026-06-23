# 027 — Apply v3 wedding copy

## Why
The couple delivered a fresh end-to-end copy doc (`untracked/copy_doc.pdf`). This sweep applies the new copy to every page except Home, Travel, and RSVP (explicitly excluded), and fills in the two big stubs: Wedding itinerary and Things to Do.

## Acceptance criteria
- [ ] `faqs.html` matches the PDF's question list and ordering, including the new "Where is the wedding…" Q, the split indoor/outdoor + weather Qs, and the real "How do I get to and from the wedding venues?" answer.
- [ ] `information-itinerary.html` has Day-of timeline and Parking sections (no more "Details coming soon.").
- [ ] `information-things-to-do.html` has Activities, Museums, Shopping, Day trips, and Bars & Restaurants sections (with sub-lists for Bars, Coffee & snacks, New Mexican cuisine, Take-out, Other), plus the Matador/Evangelos post-midnight P.S.
- [ ] `contact.html` has the new "Drop us a line / share a note of how much you adore us" intro above the existing sub-page links.
- [ ] `contact-registry.html` has the new Charities and Cash funds sub-sections.
- [ ] No changes to `home.html`, `information-travel.html`, or any `rsvp*.html` file.

## Notes
- Source copy lives in `untracked/copy_doc.pdf` (gitignored). Read the PDF directly for the source-of-truth wording.
- The charity list in the PDF marks the entries as links but doesn't include URLs; ship as plain text and follow up if links are needed.
