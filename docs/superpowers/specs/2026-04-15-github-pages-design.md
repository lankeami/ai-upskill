# GitHub Pages Design Spec

**Date:** 2026-04-15
**Status:** Draft

## Overview

Turn the ai-upskill repo into a GitHub Pages site using Jekyll, served from the `main` branch. The site has two page types: an index page and daily report pages. Jekyll must only process the index and reports — nothing else.

## Approach

Jekyll at repo root (Approach A). Add `_config.yml`, layouts, and `index.md` at the top level. GitHub Pages builds from `main` with no additional deployment setup. Jekyll's `exclude` list prevents processing of Go source, docs, and other non-site files.

## File Structure

New files:

```
_config.yml              # Jekyll config: theme, excludes, collections
_layouts/
  default.html           # Base HTML layout (head, nav, footer)
  report.html            # Report page layout (extends default)
index.md                 # Homepage with project description + report list
scripts/
  test-jekyll-output.sh  # CI script to verify Jekyll only builds expected pages
```

Modified files:

- `reports/*.md` — Add Jekyll front matter to each report
- `.github/workflows/daily-report.yml` — Add front matter injection for new reports
- `CLAUDE.md` — Add Jekyll exclusion rule

## Jekyll Configuration

`_config.yml` must exclude all non-site content:

```yaml
title: AI Daily Report
description: Daily AI news aggregated from Reddit, Hacker News, and tech RSS feeds.
exclude:
  - cmd/
  - internal/
  - docs/
  - go.mod
  - go.sum
  - config.yaml
  - "*.go"
  - ai-report
  - scripts/
  - Makefile
  - README.md
```

Any new top-level directories or Markdown files added to the repo must be added to this exclude list.

## Index Page

`index.md` renders a minimal homepage:

- **Title:** "AI Daily Report"
- **Description:** One paragraph explaining the project (aggregates AI news from Reddit, Hacker News, and RSS feeds into daily reports organized by company)
- **Report list:** Auto-generated via Liquid, iterating over report pages sorted newest-first
- **Per-report blurb:** Each entry shows the date and a summary like "OpenAI, Google, Anthropic + 5 more — 47 items", derived from front matter fields (`companies`, `item_count`)

## Report Pages

Each `reports/YYYY-MM-DD.md` file gets Jekyll front matter:

```yaml
---
layout: report
title: "AI Daily Report — 2026-04-15"
date: 2026-04-15
companies: ["OpenAI", "Google", "Anthropic", "Meta", "Microsoft", "Mistral", "xAI", "Other/Independent"]
item_count: 47
---
```

The `report` layout provides:

- **Date header** — Human-readable format (e.g., "Tuesday, April 15, 2026")
- **Table of contents** — Auto-generated list of company sections for jump navigation
- **Collapsible company sections** — Each `## Company` heading wrapped in a `<details>` element (open by default), allowing users to collapse sections they don't care about
- **Previous/Next navigation** — Links to adjacent reports at top and bottom of page
- **Back to index** link

The existing Markdown content (bullets with bold titles and source links) renders as-is within each section.

## GitHub Actions Integration

The existing `daily-report.yml` workflow is updated to add a post-processing step after the Go tool generates a report. This step:

1. Parses the generated Markdown to extract company headings and item count
2. Prepends Jekyll front matter (`layout`, `title`, `date`, `companies`, `item_count`)

The preferred approach is to update the Go Markdown renderer to emit front matter directly, keeping report generation self-contained. If that proves too invasive, a shell post-processing script is an acceptable fallback.

## Safety: Jekyll Build Test

A test script (`scripts/test-jekyll-output.sh`) runs `jekyll build` and asserts the `_site/` output only contains:

- `index.html`
- `reports/*.html`
- Jekyll assets (CSS, etc.)

If unexpected files appear (e.g., from `docs/`), the test fails. This script runs in CI as part of the GitHub Actions workflow.

## CLAUDE.md Update

Add the following rule to CLAUDE.md:

> **Jekyll Exclusion Rule:** Jekyll must only serve `index.md` and `reports/*.md`. All other directories and top-level Markdown files must be listed in `_config.yml`'s `exclude` list. When adding new top-level directories or Markdown files to the repo, add them to the exclude list and verify with `scripts/test-jekyll-output.sh`.

## Styling

Use a minimal Jekyll theme (e.g., `minima` or a custom minimal layout). The site should be clean and readable — no heavy styling needed. The report content is the focus.

## Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Site generator | Jekyll | Native GitHub Pages support, zero deploy config |
| Deployment | Main branch | Simplest setup, no separate branch needed |
| Index style | Minimal | Title, description, report list with blurbs |
| Report restyling | Collapsible sections + TOC + nav | Makes long reports more browsable |
| Safety | CI build test + CLAUDE.md rule | Prevents Jekyll from processing unintended files |
| Summary blurb | "Company1, Company2 + N more — M items" | Computed from front matter fields |
