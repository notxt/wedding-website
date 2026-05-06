# 008 — Contact section

## Why
Ship the Contact index and its three sub-pages with verbatim copy.

## Acceptance criteria
- [ ] `GET /contact` renders an index linking to the three sub-pages
- [ ] `GET /contact/about-us` renders the History section verbatim; "The proposal" and "Gallery" rendered as placeholders ("Coming soon")
- [ ] `GET /contact/registry` renders the verbatim registry copy
- [ ] `GET /contact/get-in-touch` renders the contact email, displayed obfuscated as in source: `michaelandadrien4ever at gmail.com` (no `mailto:` link, plain text — matches the doc's intent)
- [ ] `_tasks/008-public-pages-contact.md` deleted in the merging PR

## Notes
- Don't add charity links or down-payment payment links yet — copy says "below" but no specifics provided.
