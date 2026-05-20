# Handoff: Wedding Landing Page — Adrien & Michael

## Overview
A single-screen landing page for a wedding website. Composition is editorial typography ("Adrien & Michael" as the centerpiece) layered over a static cosmic scene — a night sky that gradients down into a Sangre de Cristo desert horizon at dusk. A small orbital element (planet + hairline orbit) lives quietly in the upper sky; a textured moon sits between the names lockup and the invitation line as a divider; layered mountain silhouettes ground the bottom. Below the names: an invitation line, the date, the location, and a two-column ceremony/reception schedule.

Wedding details: **Friday, September 25, 2026 — Santa Fe, New Mexico**. Ceremony at the New Mexico Museum of Art (St. Francis Auditorium); reception at Meow Wolf.

## About the Design Files
The files in this bundle are **design references created in HTML/JSX** — a prototype showing the intended look and behavior, not production code to copy directly. The task is to **recreate this design in the target codebase's existing environment** (React + CSS modules, Next.js, Astro, etc.) using its established patterns and libraries. If no environment exists yet, choose what's appropriate for a wedding microsite (Astro and Next.js are both good fits) and implement there.

## Fidelity
**High-fidelity.** Final colors, typography, spacing, and composition. The hero should be reproduced pixel-near. The sky scene (stars, moon, planet, mountains) is rendered as inline SVG inside `HeroSynthesis` — feel free to keep it as SVG or rebuild as separate components/assets, but preserve the layering and atmosphere.

## Screens / Views

### Hero (single screen)

**Purpose:** First impression for guests. Communicates the names, the date, the location, and the schedule at a glance, with a mood that matches the high-desert-at-dusk vibe of the wedding itself.

**Layout (top → bottom, all centered):**
1. **Top nav** — left: A·M monogram (SVG). Right: text links (FAQs, Travel, Things to Do, Registry) and a pill-bordered RSVP link. Tracked-out small caps, hairline weight.
2. **Edge metadata** — top-left: "An invitation —". Top-right: "Santa Fe · New Mexico". Vertically rotated label up the left margin: "TWO·THOUSAND·TWENTY·SIX".
3. **Names lockup** — three rows, all in Playfair Display Black Italic (900 italic), outlined (`-webkit-text-stroke`):
   - Top row: `ADRIEN`
   - Middle row: oversized italic `&` (regular Playfair Italic, 400) — overlaps the names above and below by ~0.36em
   - Bottom row: `MICHAEL`
4. **Moon divider** — large textured crescent moon (SVG) standing alone as a visual break between the names and the invitation line.
5. **Invitation line** — italic small caps: "two stars aligning · join them under the high desert moon", set as a single tracked line.
6. **Date lockup** — italic stack: "Friday" (small) over "the twenty-fifth of september · two-thousand twenty-six" (large).
7. **Place** — italic: "Santa Fe, New Mexico".
8. **Schedule** — two columns separated by a hairline rule:
   - Left: "Five o'clock — Ceremony — New Mexico Museum of Art"
   - Right: "Eight o'clock — Reception — Meow Wolf"
9. **Note** — bottom: small italic line "Black tie optional · Reception to follow".

**Composition rules:**
- Vertical journey from full night at the top to warm dusk at the horizon is the soul of the page.
- Names are the visual anchor — everything else is supporting type.
- The page is **static**. No motion, no parallax, no JS animations. The scene is one inline SVG.

## Interactions & Behavior
- **None on the hero itself.** Static page. Nav links navigate to subpages (FAQs, Travel, etc.) — those subpages are out of scope for this handoff.
- **Hover on nav links:** opacity 0.7 → 1, 200ms transition.
- **Responsive:** the entire composition uses `clamp()` for type sizing, so the page reflows naturally from ~1200px down to ~720px. Below 720px, you may want to stack the schedule into a single column and shrink the moon — not currently designed but a sensible breakpoint.

## State Management
None — pure static render.

## Design Tokens

### Colors
```
--hE-paper: #02030a       /* page background base (mostly hidden by sky gradient) */
--hE-ink:   #f4efe6       /* primary type — warm bone */

/* sky gradient stops (top → bottom) */
#000003  0%
#010108  22%
#03030d  42%
#060716  58%
#0a081d  70%
#1a0c28  80%
#3a1428  88%
#6b2a2a  94%
#9c4828  100%

/* accent inks */
#fdf3dc   /* invitation line / date / place */
#fff7e3   /* "join them under the high desert moon" — slightly brighter for legibility on the moon */
#f6e7c8   /* feature stars */
#d8a48f   /* note dot accent */

/* mountain silhouettes (front → back) */
#0a0612, #150a1c, #221028
```

### Typography
- **Family:** Playfair Display (Google Fonts) — weights 400, 700, 900; italics for all three.
- **Names** (`.heroE-word`): Playfair Display 900 italic, outlined via `-webkit-text-stroke: 1.6px var(--hE-ink)`, transparent fill. `font-size: clamp(96px, 17vw, 280px)`. `line-height: 0.92`. `letter-spacing: 0.02em`.
- **Ampersand** (`.heroE-amp`): Playfair Display 400 italic. `font-size: clamp(120px, 15vw, 240px)`. `line-height: 1`. Sits in the row between names with `margin: -0.36em 0` so it overlaps the descenders of "ADRIEN" and ascenders of "MICHAEL".
- **Date — large line:** Playfair Display 400 italic. `clamp(34px, 4.4vw, 60px)`. `letter-spacing: 0.005em`.
- **Date — day-of-week:** Playfair Display 400 italic. `clamp(13px, 1.2vw, 17px)`. Tracked `0.32em`. Uppercase. `opacity: 0.72`.
- **Schedule time:** Playfair Display 400 italic. `clamp(20px, 2.2vw, 28px)`.
- **Schedule label:** Playfair Display 400 italic. `clamp(12px, 1.05vw, 15px)`. Tracked `0.18em`. Uppercase.
- **Edge metadata / nav links:** Playfair Display 400. 11px. `letter-spacing: 0.32em`. Uppercase.
- **Invitation line + place:** Playfair Display 400 italic. Mid sizes, tracked `0.04em`.

### Spacing
- Page padding: `clamp(140px, 16vh, 200px) 40px clamp(120px, 14vh, 170px)`.
- Stage gap (between zones): `clamp(48px, 8vh, 100px)`.
- Names row leading: `-0.36em` (ampersand overlaps neighboring names).
- Moon divider margin: `clamp(20px, 4vh, 56px)` top and bottom.
- Hairline divider margin: `clamp(20px, 4vh, 56px)` top and bottom.
- Min hero height: `190vh` (the page is intentionally taller than viewport — the composition wants room to breathe).
- Nav padding: `22px 36px`.
- Edge label inset: `36px` from edges, `90px` from top.

### Border / Shadow
- Names text-shadow: `0 2px 32px rgba(0, 0, 0, 0.55)` (subtle halo for legibility against the night sky).
- RSVP pill: `1px solid currentColor`, `border-radius: 999px`, padding `9px 16px`.
- Hairline rules (dividers, mountain edges): 1px, semi-transparent ink.

## Assets
- **`assets/ma-mark.svg`** — A·M monogram lockup used in the nav. Custom mark, included.
- **Sky scene** — entirely inline SVG inside the `HeroSynthesis` component. Stars are deterministically seeded (PRNG with seed `91827`); a few feature sparkles use `HeroESparkle`. Moon, planet, orbit ring, milky way band, and mountain ridges are all inline SVG paths within the component.
- **Fonts** — Playfair Display via Google Fonts. Imported in `preview-synthesis.html`:
  ```
  https://fonts.googleapis.com/css2?family=Playfair+Display:ital,wght@0,400;0,700;0,900;1,400;1,700;1,900&display=swap
  ```

## Files
- `preview-synthesis.html` — entry point. Loads React + Babel + heroes.jsx and renders `<HeroSynthesis />`.
- `heroes.jsx` — contains `MockNav`, `HeroESparkle`, the deterministic `HERO_E_STARS` array, and `HeroSynthesis` (the full hero component with inline SVG sky and mountains).
- `heroes.css` — all styling. Selectors are scoped under `.heroE` and `.hNav`.
- `assets/ma-mark.svg` — monogram in the nav.

## Implementation Notes
- The "moon as divider" and the "names lockup" are the two non-negotiable design beats. Every other element is supporting — feel free to adjust the sky's exact stops or rework the SVG mountains to match the codebase's conventions.
- Stars are intentionally **denser at the top** (Math.pow(rand(), 2.4) bias) so the field reads as deep space at the top of the page and tapers as you approach the horizon.
- The `&` glyph overlap (`-0.36em` margin on the middle row) is the typographic detail that makes the names feel like a proper editorial lockup rather than three stacked words. Preserve it.
- Outlined Black Italic (Playfair 900 italic with `-webkit-text-stroke`) is supported in all modern browsers including Safari, so no fallback is required, but if you want belt-and-suspenders, a fallback solid fill at `#f4efe6` is fine.
