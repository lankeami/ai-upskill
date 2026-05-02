# Modern Theme Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace CSS in `_layouts/default.html` and `_layouts/report.html` with a modern dark-header / indigo-accent design using Inter font.

**Architecture:** All styling lives in `<style>` blocks within the two layout files. CSS custom properties (`:root` vars) are declared in `default.html` and inherited by `report.html` since it renders inside the default layout. No external stylesheets, no new files.

**Tech Stack:** Jekyll/GitHub Pages, plain CSS, Google Fonts (Inter)

---

## File Map

| File | Change |
|---|---|
| `_layouts/default.html` | Full rewrite: new `<style>` block, Inter font `<link>`, header inner wrapper div, `<main class="site-main">` |
| `_layouts/report.html` | Full rewrite of `<style>` block; remove inline `style=` from `.podcast-player` div |

---

## Task 1: Rewrite `_layouts/default.html`

**Files:**
- Modify: `_layouts/default.html`

- [ ] **Step 1: Verify current build is clean (baseline)**

```bash
cd /path/to/ai-upskill
bundle exec jekyll build 2>&1 | tail -5
```
Expected: `done in X seconds` with no errors. If it errors, stop and investigate before proceeding.

- [ ] **Step 2: Replace `_layouts/default.html` with the new version**

Write the entire file with this exact content:

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{ page.title | default: site.title }}</title>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet">
  <style>
    *, *::before, *::after { margin: 0; padding: 0; box-sizing: border-box; }
    :root {
      --bg: #ffffff;
      --surface: #f8fafc;
      --border: #e2e8f0;
      --text: #1e293b;
      --text-muted: #64748b;
      --accent: #6366f1;
      --accent-hover: #4f46e5;
      --header-bg: #0f172a;
      --header-text: #f1f5f9;
    }
    body {
      font-family: 'Inter', system-ui, -apple-system, sans-serif;
      font-size: 16px;
      line-height: 1.7;
      color: var(--text);
      background: var(--bg);
      min-height: 100vh;
      display: flex;
      flex-direction: column;
    }
    .site-header {
      background: var(--header-bg);
      border-bottom: 1px solid rgba(255, 255, 255, 0.08);
      position: sticky;
      top: 0;
      z-index: 100;
    }
    .site-header-inner {
      max-width: 860px;
      margin: 0 auto;
      padding: 0 1.5rem;
      height: 60px;
      display: flex;
      align-items: center;
    }
    .site-header a {
      color: var(--header-text);
      text-decoration: none;
      font-weight: 700;
      font-size: 1.1rem;
      letter-spacing: -0.01em;
    }
    .site-header a:hover { color: #a5b4fc; text-decoration: none; }
    .site-main {
      flex: 1;
      max-width: 860px;
      margin: 0 auto;
      width: 100%;
      padding: 2.5rem 1.5rem;
    }
    .site-footer {
      background: var(--header-bg);
      border-top: 1px solid rgba(255, 255, 255, 0.08);
      padding: 1.5rem;
      color: #94a3b8;
      font-size: 0.875rem;
      text-align: center;
    }
    h1 { font-size: 2rem; font-weight: 700; line-height: 1.25; margin-bottom: 0.5rem; letter-spacing: -0.02em; }
    h2 { font-size: 1.4rem; font-weight: 600; line-height: 1.35; margin-top: 2rem; margin-bottom: 0.75rem; }
    h3 { font-size: 1.1rem; font-weight: 600; margin-top: 1.5rem; margin-bottom: 0.5rem; }
    p { margin-bottom: 1rem; }
    a { color: var(--accent); text-decoration: none; }
    a:hover { color: var(--accent-hover); text-decoration: underline; }
    hr { border: none; border-top: 1px solid var(--border); margin: 2rem 0; }
    ul { padding-left: 1.5rem; }
    li { margin-bottom: 0.25rem; }
    /* Index page: style top-level ul > li as hover-lift cards */
    .site-main > ul {
      list-style: none;
      padding: 0;
      margin-top: 1.5rem;
      display: flex;
      flex-direction: column;
      gap: 0.75rem;
    }
    .site-main > ul > li {
      background: var(--surface);
      border: 1px solid var(--border);
      border-radius: 10px;
      padding: 1rem 1.25rem;
      transition: box-shadow 0.15s ease, transform 0.15s ease, border-color 0.15s ease;
    }
    .site-main > ul > li:hover {
      box-shadow: 0 4px 16px rgba(0, 0, 0, 0.08);
      transform: translateY(-2px);
      border-color: var(--accent);
    }
    .site-main > ul > li a { font-weight: 600; }
    @media (max-width: 600px) {
      h1 { font-size: 1.5rem; }
      .site-main { padding: 1.5rem 1rem; }
    }
  </style>
</head>
<body>
  <header class="site-header">
    <div class="site-header-inner">
      <a href="{{ '/' | relative_url }}">{{ site.title }}</a>
    </div>
  </header>
  <main class="site-main">
    {{ content }}
  </main>
  <footer class="site-footer">
    <p>Generated daily from Reddit, Hacker News, and RSS feeds.</p>
  </footer>
<script>
document.addEventListener('DOMContentLoaded', function() {
  document.querySelectorAll('a[href]').forEach(function(a) {
    if (!a.getAttribute('href').startsWith('#')) {
      a.setAttribute('target', '_blank');
      a.setAttribute('rel', 'noopener noreferrer');
    }
  });
});
</script>
</body>
</html>
```

- [ ] **Step 3: Verify Jekyll builds cleanly**

```bash
bundle exec jekyll build 2>&1 | tail -5
```
Expected: `done in X seconds`, no errors or warnings about layouts.

- [ ] **Step 4: Commit**

```bash
git add _layouts/default.html
git commit -m "style: rewrite default layout with dark header and Inter font"
```

---

## Task 2: Rewrite `_layouts/report.html` CSS and podcast player HTML

**Files:**
- Modify: `_layouts/report.html`

The only HTML change is removing the inline `style=` attribute from the `.podcast-player` container div (line ~58). All Liquid, JavaScript, and element IDs remain byte-for-byte identical.

- [ ] **Step 1: Replace the `<style>` block (lines 5–20) in `_layouts/report.html`**

The current style block starts with `<style>` on line 5 and ends with `</style>` on line 20. Replace it with:

```html
<style>
  .report-nav {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 2rem;
    font-size: 0.875rem;
    gap: 1rem;
  }
  .report-nav a {
    color: var(--accent);
    font-weight: 500;
    padding: 0.4rem 0.75rem;
    border: 1px solid var(--border);
    border-radius: 6px;
    transition: background 0.15s, color 0.15s, border-color 0.15s;
  }
  .report-nav a:hover {
    background: var(--accent);
    color: #fff;
    text-decoration: none;
    border-color: var(--accent);
  }
  .report-nav .disabled {
    color: #cbd5e1;
    padding: 0.4rem 0.75rem;
    border: 1px solid var(--border);
    border-radius: 6px;
  }
  .report-date {
    font-size: 1rem;
    color: var(--text-muted);
    margin-bottom: 1.5rem;
    font-weight: 500;
  }
  .report-toc {
    margin-bottom: 2rem;
    padding: 1.25rem 1.5rem;
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 10px;
  }
  .report-toc h3 {
    font-size: 0.75rem;
    margin: 0 0 0.75rem;
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.08em;
    font-weight: 600;
  }
  .report-toc ul {
    list-style: none;
    padding: 0;
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
  }
  .report-toc li a {
    font-size: 0.8rem;
    padding: 0.3rem 0.65rem;
    background: #fff;
    border: 1px solid var(--border);
    border-radius: 20px;
    display: inline-block;
    font-weight: 500;
    transition: background 0.15s, color 0.15s, border-color 0.15s;
  }
  .report-toc li a:hover {
    background: var(--accent);
    color: #fff;
    text-decoration: none;
    border-color: var(--accent);
  }
  details {
    margin-bottom: 1rem;
    border: 1px solid var(--border);
    border-radius: 8px;
    overflow: hidden;
  }
  details summary {
    cursor: pointer;
    font-size: 1.1rem;
    font-weight: 600;
    padding: 0.875rem 1.25rem;
    background: var(--surface);
    list-style: none;
    display: flex;
    align-items: center;
    gap: 0.5rem;
    transition: background 0.15s;
    user-select: none;
  }
  details summary::-webkit-details-marker { display: none; }
  details summary::before {
    content: '▶';
    font-size: 0.65rem;
    color: var(--accent);
    transition: transform 0.2s ease;
    display: inline-block;
    flex-shrink: 0;
  }
  details[open] summary::before { transform: rotate(90deg); }
  details summary:hover { background: #f1f5f9; color: var(--accent); }
  details[open] { border-color: var(--accent); }
  details[open] summary { color: var(--accent); border-bottom: 1px solid var(--border); }
  details > ul { padding: 0.75rem 1.25rem 1rem 2.75rem; }
  #report-content li p {
    font-size: 0.875rem;
    color: var(--text-muted);
    margin-top: 0.2rem;
    line-height: 1.4;
  }
  .podcast-player {
    margin-bottom: 2rem;
    padding: 1.5rem;
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 12px;
  }
  #btn-play { background: var(--accent) !important; }
</style>
```

- [ ] **Step 2: Remove inline `style=` from the `.podcast-player` div**

Find this line (around line 58 after the style block):
```html
<div class="podcast-player" style="margin-bottom: 1.5rem; padding: 1.25rem; background: #f6f8fa; border-radius: 8px;">
```

Replace it with:
```html
<div class="podcast-player">
```

Everything else in the podcast player HTML (buttons, seek bar, speed selector, script) stays exactly as-is.

- [ ] **Step 3: Verify Jekyll builds cleanly**

```bash
bundle exec jekyll build 2>&1 | tail -5
```
Expected: `done in X seconds`, no errors.

- [ ] **Step 4: Commit**

```bash
git add _layouts/report.html
git commit -m "style: rewrite report layout with modern theme"
```

---

## Task 3: Visual Verification

**Files:** None (read-only verification)

- [ ] **Step 1: Start Jekyll dev server**

```bash
bundle exec jekyll serve --livereload 2>&1 &
```
Wait for `Server address: http://127.0.0.1:4000` in the output before proceeding.

- [ ] **Step 2: Verify index page at 1440px**

Open `http://127.0.0.1:4000/ai-upskill/` in a browser at 1440px width. Confirm:
- [ ] Dark `#0f172a` sticky nav bar visible at top with white site title "AI Daily Report"
- [ ] Report list items render as cards (rounded border, surface background)
- [ ] Hovering a card lifts it (translateY + shadow) and border turns indigo
- [ ] Footer is dark matching the header

- [ ] **Step 3: Verify index page at 375px**

Resize browser to 375px width. Confirm:
- [ ] Nav bar still visible, title readable
- [ ] Cards stack full-width with appropriate padding
- [ ] No horizontal overflow / scrollbar

- [ ] **Step 4: Verify report page at 1440px**

Open any report, e.g. `http://127.0.0.1:4000/ai-upskill/reports/2026-05-01`. Confirm:
- [ ] Prev/next nav buttons are pill-shaped, hover fills indigo
- [ ] TOC panel is a card with pill-shaped company chips; chips hover fills indigo
- [ ] Company `<details>` sections show a small `▶` chevron in indigo; chevron rotates on open
- [ ] Open `<details>` section has indigo border
- [ ] Podcast player is a card; play button is indigo

- [ ] **Step 5: Verify report page at 375px**

Resize to 375px. Confirm:
- [ ] TOC chips wrap correctly
- [ ] Podcast player controls don't overflow
- [ ] No horizontal overflow

- [ ] **Step 6: Verify JavaScript still works**

On the report page:
- [ ] Click a company `<details>` summary — section collapses/expands (JS-driven collapsible)
- [ ] Click podcast play button — audio loads and plays, seek bar moves
- [ ] Click any external link — opens in new tab (`target="_blank"` from JS hook)

- [ ] **Step 7: Stop the dev server and do a final clean build**

```bash
kill %1
bundle exec jekyll build 2>&1 | tail -5
```
Expected: clean build, no errors.

- [ ] **Step 8: Commit verification note**

```bash
git add -p  # no changes expected; just confirm working tree is clean
git status
```
Expected: `nothing to commit, working tree clean`
