# Modern Theme Redesign — Design Spec

**Date:** 2026-05-01
**Approach:** Option B — Dark Header + Indigo Accent (light mode only)

## Scope

Modify exactly two files:
- `_layouts/default.html`
- `_layouts/report.html`

No other files are touched. No new files, Jekyll plugins, or external dependencies are added (except a Google Fonts `<link>` tag).

---

## Color System

CSS custom properties declared in `default.html`'s `<style>` block, available in both layouts since `report.html` renders inside `default.html`.

| Token | Value | Usage |
|---|---|---|
| `--bg` | `#ffffff` | Page body background |
| `--surface` | `#f8fafc` | Cards, TOC panel, podcast player |
| `--border` | `#e2e8f0` | All borders |
| `--text` | `#1e293b` | Body text |
| `--text-muted` | `#64748b` | Secondary text, dates, descriptions |
| `--accent` | `#6366f1` | Links, chips, play button, open details border |
| `--accent-hover` | `#4f46e5` | Hover state for accent elements |
| `--header-bg` | `#0f172a` | Nav bar and footer background |
| `--header-text` | `#f1f5f9` | Site title in nav bar |

---

## Typography

- **Font:** Inter (weights 400, 500, 600, 700) loaded via `<link>` from Google Fonts in `<head>`
- **Fallback:** `system-ui, -apple-system, sans-serif`
- **Body:** 16px / 1.7 line-height
- **H1:** 2rem / 700 weight / `-0.02em` letter-spacing
- **H2:** 1.4rem / 600 weight
- **H3:** 1.1rem / 600 weight

---

## `default.html` Layout

### Header
- Sticky (`position: sticky; top: 0; z-index: 100`)
- Full-width, `--header-bg` background, 60px tall
- Inner wrapper capped at 860px, centered
- Site title: left-aligned, bold, `--header-text` color, links to `/`
- Title hover: `#a5b4fc` (soft indigo-tinted white)
- Subtle bottom border: `1px solid rgba(255,255,255,0.08)`

### Body
- `display: flex; flex-direction: column; min-height: 100vh` — footer always sticks to bottom
- `.site-main` container: max-width 860px, centered, `2.5rem 1.5rem` padding, `flex: 1`

### Footer
- Full-width, `--header-bg` background, `#94a3b8` text, centered, `1.5rem` padding
- Mirrors header visually with matching top border

### Index Page Cards
The index `{{ content }}` renders as a `<ul>`. CSS targets `.site-main > ul`:
- Remove list-style, set `flex-direction: column; gap: 0.75rem`

Each `<li>` becomes a card:
- `--surface` background, `1px solid --border` border, `10px` border-radius, `1rem 1.25rem` padding
- Hover: `translateY(-2px)` lift, `0 4px 16px rgba(0,0,0,0.08)` shadow, border-color → `--accent`
- Link inside: `font-weight: 600`

### Responsive
- At ≤600px: H1 drops to 1.5rem, `.site-main` padding reduces to `1.5rem 1rem`

---

## `report.html` Layout

### Prev/Next Navigation
- Pill-shaped buttons: `6px` border-radius, `1px solid --border`, `0.4rem 0.75rem` padding
- Hover: fill with `--accent`, white text, border → `--accent`
- Disabled spans: same shape in muted gray (no layout shift)

### Report Date
- `--text-muted`, `1rem`, `font-weight: 500`, `1.5rem` bottom margin

### TOC Company Chips
- Container: `--surface` bg, `1px solid --border`, `10px` radius, `1.25rem 1.5rem` padding
- Label: `0.75rem`, uppercase, `0.08em` letter-spacing, `--text-muted`
- Each chip: pill (`border-radius: 20px`), white bg, `--border` border, `0.8rem` font, `font-weight: 500`
- Chip hover: fill `--accent`, white text, border → `--accent`

### Company `<details>` Sections
- Container: `1px solid --border` border, `8px` radius
- Summary: `--surface` bg, `0.875rem 1.25rem` padding, CSS `::before` chevron (`▶`) in `--accent`
- Chevron rotates 90° when open via CSS `transform` — no HTML or JS changes
- `[open]` state: border-color → `--accent`, summary text → `--accent`
- Summary hover: text → `--accent`, bg lightens to `#f1f5f9`

### News Item Descriptions (`li p`)
- `0.875rem`, `--text-muted`, `1.4` line-height (tighter than body for visual density)

### Podcast Player
- Remove inline `style=` from `.podcast-player` container div; handled via CSS class
- Container CSS: `--surface` bg, `1px solid --border`, `12px` radius, `1.5rem` padding
- Play button: `#btn-play { background: var(--accent) !important; }` — overrides existing inline style
- Rewind/forward buttons: retain existing `#e1e8f0` background

---

## Hard Constraints

- NEVER change Liquid template logic, `{% %}` tags, front matter, or JavaScript behavior
- NEVER rename or restructure HTML elements targeted by JS IDs: `podcast-audio`, `btn-play`, `btn-rewind`, `btn-forward`, `podcast-seek`, `podcast-current`, `podcast-duration`, `podcast-speed`, `report-content`
- NEVER add new files, Jekyll plugins, or dependencies beyond the Google Fonts `<link>`
- NEVER modify any file other than `_layouts/default.html` and `_layouts/report.html`
- Stop and ask before touching any other file

---

## Done Criteria

- Both layout files saved with new CSS
- Site title renders in `--header-bg` sticky nav bar
- Index report items display as hover-lift cards
- Report page TOC chips and prev/next nav use `--accent`
- Company `<details>` sections have left-border accent and chevron rotation
- Podcast player card uses `--accent` play button
- All existing JS (collapsible sections, podcast player, external link targeting) still functions
- Renders correctly at 375px and 1440px viewport widths
