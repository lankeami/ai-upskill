# Audio Podcast Generation via NotebookLM

**Date:** 2026-04-16
**Status:** Draft
**Approach:** Python script called from GitHub Actions

## Overview

Add automated audio podcast generation to the daily AI report pipeline using Google NotebookLM's "Audio Overview" feature via the unofficial `notebooklm-py` Python library. Each daily report gets a companion podcast episode published as a GitHub Release asset, with an embedded audio player on the Jekyll report page.

## Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Audio format | Deep Dive (two hosts discuss) | Most engaging NotebookLM format |
| Audio length | Short | Daily reports don't need long episodes |
| Hosting | GitHub Release artifacts | Persistent, doesn't bloat repo, free |
| CI integration | Separate workflow | Isolates podcast failures from report publishing |
| Implementation | Standalone Python script | Simple, keeps Go codebase unchanged |

## Architecture

```
daily-report.yml                    podcast.yml
┌──────────────┐                   ┌──────────────────────────┐
│ Generate      │   push to main   │ Trigger on reports/*.md  │
│ report .md    │ ─────────────►   │ change                   │
│ Commit & push │                  │                          │
└──────────────┘                   │ 1. Install notebooklm-py │
                                   │ 2. Run generate-podcast  │
                                   │ 3. Create GitHub Release │
                                   │ 4. Update report front   │
                                   │    matter + commit       │
                                   └──────────────────────────┘
```

## Components

### 1. Python Script: `scripts/generate-podcast.py`

**Input:** Report date (CLI argument, defaults to today)
**Output:** MP3 file at `podcasts/YYYY-MM-DD.mp3`

**Flow:**
1. Parse CLI args for `--date YYYY-MM-DD`
2. Read `reports/YYYY-MM-DD.md`
3. Strip YAML front matter, extract clean markdown content
4. Create a NotebookLM notebook titled "AI Daily Report — YYYY-MM-DD"
5. Add the report content as a text source via `client.sources.add_text()`
6. Generate audio with `AudioFormat.DEEP_DIVE` and `AudioLength.SHORT`
7. Poll for completion via `client.artifacts.wait_for_completion()`
8. Download MP3 to `podcasts/YYYY-MM-DD.mp3`
9. Delete the notebook to avoid clutter in the NotebookLM account

**Dependencies:** `notebooklm-py` (installed via `pip install notebooklm-py` in CI)

**Authentication:** Uses `NOTEBOOKLM_AUTH_JSON` environment variable. In CI, this is stored as a GitHub Actions secret. Locally, run `notebooklm login` to authenticate via browser.

**Error handling:** If any step fails, the script exits with a non-zero code. The workflow handles this gracefully — a failed podcast does not block report publishing.

### 2. GitHub Actions Workflow: `.github/workflows/podcast.yml`

**Trigger:** `push` to `main` branch, filtered to `reports/*.md` path changes.

**Steps:**
1. **Checkout** the repository
2. **Set up Python 3.12**
3. **Install** `notebooklm-py` via pip
4. **Extract report date** from the changed file path (e.g., `2026-04-16` from `reports/2026-04-16.md`)
5. **Run** `python scripts/generate-podcast.py --date YYYY-MM-DD`
6. **Create GitHub Release** tagged `podcast-YYYY-MM-DD` with the MP3 as an attached asset using `gh release create`
7. **Update report front matter** — add `podcast_url` field pointing to the release asset URL
8. **Commit and push** the updated report markdown

**Secrets required:**
- `NOTEBOOKLM_AUTH_JSON` — Google authentication cookies JSON from `notebooklm login`

**Failure handling:** Workflow failure does not affect the already-published report. Can be re-triggered manually via `workflow_dispatch`.

### 3. Report Front Matter Update

After the podcast is generated and uploaded, the workflow adds a `podcast_url` field to the report's YAML front matter:

```yaml
---
layout: report
title: "AI Daily Report — 2026-04-16"
date: 2026-04-16
companies:
  - OpenAI
  - Google
item_count: 50
podcast_url: https://github.com/lankeami/ai-upskill/releases/download/podcast-2026-04-16/podcast-2026-04-16.mp3
---
```

Reports without a `podcast_url` field display no audio player.

### 4. Jekyll Layout Update: `_layouts/report.html`

Add a conditional audio player block to `report.html` that renders when `podcast_url` is present:

```html
{% if page.podcast_url %}
<div class="podcast-player">
  <h3>Listen to this report</h3>
  <audio controls preload="none">
    <source src="{{ page.podcast_url }}" type="audio/mpeg">
    Your browser does not support the audio element.
  </audio>
  <p><a href="{{ page.podcast_url }}">Download podcast</a></p>
</div>
{% endif %}
```

Placed at the top of the report content, before the company sections.

### 5. Configuration Changes

**`_config.yml`:** Add `podcasts/` to the `exclude` list (temporary local staging directory).

**`.gitignore`:** Add `podcasts/` to prevent MP3s from being committed to the repo.

## Auth Session Management

The `notebooklm-py` library authenticates via browser cookies that expire every few weeks. When cookies expire:

1. The podcast workflow will fail
2. A developer runs `notebooklm login` locally
3. Copies the resulting JSON from `~/.notebooklm/profiles/default/storage_state.json`
4. Updates the `NOTEBOOKLM_AUTH_JSON` GitHub Actions secret

This is a known maintenance burden of using an unofficial library with no API key auth.

## Release Asset URL Pattern

```
https://github.com/lankeami/ai-upskill/releases/download/podcast-{YYYY-MM-DD}/podcast-{YYYY-MM-DD}.mp3
```

This URL is deterministic and permanent — assets persist across all releases independently.

## Files Changed

| File | Change |
|------|--------|
| `scripts/generate-podcast.py` | New — podcast generation script |
| `.github/workflows/podcast.yml` | New — podcast CI workflow |
| `_layouts/report.html` | Modified — add conditional audio player |
| `_config.yml` | Modified — add `podcasts/` to exclude list |
| `.gitignore` | Modified — add `podcasts/` |
