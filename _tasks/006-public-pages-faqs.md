# 006 — FAQs page

## Why
Wire up the FAQs route with the verbatim copy from the copy doc.

## Acceptance criteria
- [ ] `internal/templates/faqs.html` extends layout, contains all questions and answers verbatim from `context/M+A website copy doc.pdf`
- [ ] `GET /faqs` returns 200 and renders the page
- [ ] Nav highlights the active page (basic — can be a CSS class set from template data)
- [ ] `_tasks/006-public-pages-faqs.md` deleted in the merging PR

## Notes
Q/As to include (copy verbatim from page 1–2 of the copy doc):
- When do I need to RSVP by?
- Is the wedding indoors or outdoors?
- What's the dress code? (with footwear + "do not wear" sub-bullets)
- How do I get to and from the wedding venues?
- What time should I arrive for the ceremony?
- What time should I arrive for the reception?
- What if I have food allergies or dietary restrictions?
- Are children allowed?
- Wait a minute…I thought you guys were married already?

The "Contact us" link inside the RSVP-deadline answer should link to `/contact/get-in-touch`.
