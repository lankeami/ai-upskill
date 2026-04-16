# Podcast Generation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Generate audio podcasts from daily AI reports using NotebookLM and publish as GitHub Release artifacts with embedded audio players on report pages.

**Architecture:** A standalone Python script (`scripts/generate-podcast.py`) uses the `notebooklm-py` library to convert markdown reports into MP3 podcasts. A separate GitHub Actions workflow (`podcast.yml`) triggers on report commits, runs the script, uploads the MP3 as a GitHub Release asset, and updates the report's front matter with the podcast URL.

**Tech Stack:** Python 3.12, notebooklm-py, GitHub Actions, GitHub Releases, Jekyll/Liquid

---

## File Structure

| File | Purpose |
|------|---------|
| `scripts/generate-podcast.py` | **Create** — Async Python script that reads a report, sends to NotebookLM, downloads MP3 |
| `.github/workflows/podcast.yml` | **Create** — GitHub Actions workflow triggered by report commits |
| `_layouts/report.html` | **Modify** — Add conditional audio player when `podcast_url` front matter exists |
| `_config.yml` | **Modify** — Add `podcasts/` to exclude list |
| `.gitignore` | **Modify** — Add `podcasts/` entry |
| `scripts/test-jekyll-output.sh` | **Verify** — Confirm it still passes after config changes |

---

### Task 1: Configuration changes (.gitignore, _config.yml)

**Files:**
- Modify: `.gitignore`
- Modify: `_config.yml`

- [ ] **Step 1: Add `podcasts/` to `.gitignore`**

Append to `.gitignore`:

```
podcasts/
```

The file currently contains only `.worktrees/`, so after this change it will be:

```
.worktrees/
podcasts/
```

- [ ] **Step 2: Add `podcasts/` to `_config.yml` exclude list**

Add `podcasts/` to the `exclude` list in `_config.yml`. Append after the last entry (`vendor/`):

```yaml
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
  - CLAUDE.md
  - .claude/
  - ai-report/
  - Gemfile.lock
  - vendor/
  - podcasts/
```

- [ ] **Step 3: Verify Jekyll output test still passes**

Run: `./scripts/test-jekyll-output.sh`
Expected: PASS — only `index.html` and `reports/*.html` in output.

- [ ] **Step 4: Commit**

```bash
git add .gitignore _config.yml
git commit -m "chore: add podcasts/ to gitignore and Jekyll exclude list"
```

---

### Task 2: Jekyll layout — audio player

**Files:**
- Modify: `_layouts/report.html:53-54` (insert after the `<h1>` and date paragraph)

- [ ] **Step 1: Add podcast player to report layout**

In `_layouts/report.html`, insert the following block after line 54 (`<p class="report-date">...</p>`) and before line 56 (`{% if page.companies %}`):

```html
{% if page.podcast_url %}
<div class="podcast-player" style="margin-bottom: 1.5rem; padding: 1rem; background: #f6f8fa; border-radius: 6px;">
  <h3 style="font-size: 0.9rem; margin-bottom: 0.5rem; color: #656d76; text-transform: uppercase; letter-spacing: 0.05em;">Listen to this report</h3>
  <audio controls preload="none" style="width: 100%; margin-bottom: 0.5rem;">
    <source src="{{ page.podcast_url }}" type="audio/mpeg">
    Your browser does not support the audio element.
  </audio>
  <p style="margin: 0; font-size: 0.85rem;"><a href="{{ page.podcast_url }}">Download podcast</a></p>
</div>
{% endif %}
```

This goes between the date paragraph and the companies TOC. Styling matches the existing `.report-toc` block.

- [ ] **Step 2: Verify layout renders without podcast_url**

Run: `bundle exec jekyll build`
Then open `_site/reports/2026-04-16.html` and verify no audio player appears (since `podcast_url` is not in the front matter yet).

- [ ] **Step 3: Test with a mock podcast_url**

Temporarily add `podcast_url: https://example.com/test.mp3` to `reports/2026-04-16.md` front matter:

```yaml
---
layout: report
title: "AI Daily Report — 2026-04-16"
date: 2026-04-16
companies: ["Other/Independent", "anthropic", "google", "meta", "microsoft", "openai", "xai"]
item_count: 50
podcast_url: https://example.com/test.mp3
---
```

Run: `bundle exec jekyll build`
Open `_site/reports/2026-04-16.html` and verify the audio player HTML is present with the test URL.

- [ ] **Step 4: Revert the test front matter change**

Remove the `podcast_url` line from `reports/2026-04-16.md`.

- [ ] **Step 5: Run Jekyll output test**

Run: `./scripts/test-jekyll-output.sh`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add _layouts/report.html
git commit -m "feat: add conditional podcast audio player to report layout"
```

---

### Task 3: Python podcast generation script

**Files:**
- Create: `scripts/generate-podcast.py`

- [ ] **Step 1: Create the podcast generation script**

Create `scripts/generate-podcast.py`:

```python
#!/usr/bin/env python3
"""Generate an audio podcast from a daily AI report using NotebookLM."""

import argparse
import asyncio
import re
import sys
from datetime import date, datetime
from pathlib import Path

REPORTS_DIR = Path(__file__).resolve().parent.parent / "reports"
PODCASTS_DIR = Path(__file__).resolve().parent.parent / "podcasts"


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Generate podcast from AI daily report")
    parser.add_argument(
        "--date",
        type=str,
        default=date.today().isoformat(),
        help="Report date in YYYY-MM-DD format (default: today)",
    )
    return parser.parse_args()


def strip_front_matter(content: str) -> str:
    """Remove YAML front matter delimited by --- lines."""
    match = re.match(r"^---\s*\n.*?\n---\s*\n", content, re.DOTALL)
    if match:
        return content[match.end():]
    return content


async def generate_podcast(report_date: str) -> Path:
    """Generate a podcast MP3 from the report for the given date."""
    from notebooklm import NotebookLMClient
    from notebooklm.enums import AudioFormat, AudioLength

    report_path = REPORTS_DIR / f"{report_date}.md"
    if not report_path.exists():
        print(f"Error: Report not found at {report_path}", file=sys.stderr)
        sys.exit(1)

    raw_content = report_path.read_text(encoding="utf-8")
    content = strip_front_matter(raw_content)

    if not content.strip():
        print(f"Error: Report {report_path} is empty after stripping front matter", file=sys.stderr)
        sys.exit(1)

    PODCASTS_DIR.mkdir(exist_ok=True)
    output_path = PODCASTS_DIR / f"{report_date}.mp3"

    print(f"Generating podcast for {report_date}...")

    async with await NotebookLMClient.from_storage() as client:
        # Create notebook
        notebook_title = f"AI Daily Report — {report_date}"
        nb = await client.notebooks.create(notebook_title)
        print(f"Created notebook: {notebook_title} ({nb.id})")

        try:
            # Add report as text source
            await client.sources.add_text(nb.id, notebook_title, content)
            print("Added report content as source")

            # Generate audio
            status = await client.artifacts.generate_audio(
                nb.id,
                audio_format=AudioFormat.DEEP_DIVE,
                audio_length=AudioLength.SHORT,
            )
            print(f"Audio generation started (task: {status.task_id})")

            # Wait for completion
            await client.artifacts.wait_for_completion(nb.id, status.task_id)
            print("Audio generation complete")

            # Download MP3
            await client.artifacts.download_audio(nb.id, str(output_path))
            print(f"Downloaded podcast to {output_path}")

        finally:
            # Clean up notebook
            await client.notebooks.delete(nb.id)
            print(f"Cleaned up notebook {nb.id}")

    return output_path


def main() -> None:
    args = parse_args()

    # Validate date format
    try:
        datetime.strptime(args.date, "%Y-%m-%d")
    except ValueError:
        print(f"Error: Invalid date format '{args.date}'. Use YYYY-MM-DD.", file=sys.stderr)
        sys.exit(1)

    output = asyncio.run(generate_podcast(args.date))
    print(f"\nPodcast generated: {output}")


if __name__ == "__main__":
    main()
```

- [ ] **Step 2: Make the script executable**

Run: `chmod +x scripts/generate-podcast.py`

- [ ] **Step 3: Verify script parses arguments correctly**

Run: `python3 scripts/generate-podcast.py --help`

Expected output:
```
usage: generate-podcast.py [-h] [--date DATE]

Generate podcast from AI daily report

options:
  -h, --help   show this help message and exit
  --date DATE  Report date in YYYY-MM-DD format (default: today)
```

- [ ] **Step 4: Verify script fails gracefully for missing report**

Run: `python3 scripts/generate-podcast.py --date 1999-01-01`

Expected: `Error: Report not found at .../reports/1999-01-01.md` and exit code 1.

- [ ] **Step 5: Commit**

```bash
git add scripts/generate-podcast.py
git commit -m "feat: add podcast generation script using NotebookLM"
```

---

### Task 4: GitHub Actions workflow

**Files:**
- Create: `.github/workflows/podcast.yml`

- [ ] **Step 1: Create the podcast workflow**

Create `.github/workflows/podcast.yml`:

```yaml
name: Generate Podcast

on:
  push:
    branches: [main]
    paths: ['reports/*.md']
  workflow_dispatch:
    inputs:
      date:
        description: 'Report date (YYYY-MM-DD)'
        required: true

permissions:
  contents: write

jobs:
  generate-podcast:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Set up Python
        uses: actions/setup-python@v5
        with:
          python-version: '3.12'

      - name: Install notebooklm-py
        run: pip install notebooklm-py

      - name: Determine report date
        id: date
        run: |
          if [ -n "${{ github.event.inputs.date }}" ]; then
            echo "report_date=${{ github.event.inputs.date }}" >> "$GITHUB_OUTPUT"
          else
            # Find the most recently changed report file
            REPORT_FILE=$(git diff --name-only HEAD~1 HEAD -- 'reports/*.md' | head -1)
            if [ -z "$REPORT_FILE" ]; then
              echo "No report file changed in last commit"
              exit 1
            fi
            # Extract date from filename: reports/YYYY-MM-DD.md -> YYYY-MM-DD
            REPORT_DATE=$(basename "$REPORT_FILE" .md)
            echo "report_date=$REPORT_DATE" >> "$GITHUB_OUTPUT"
          fi

      - name: Generate podcast
        env:
          NOTEBOOKLM_AUTH_JSON: ${{ secrets.NOTEBOOKLM_AUTH_JSON }}
        run: python scripts/generate-podcast.py --date ${{ steps.date.outputs.report_date }}

      - name: Create GitHub Release
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: |
          REPORT_DATE="${{ steps.date.outputs.report_date }}"
          TAG="podcast-${REPORT_DATE}"
          MP3_FILE="podcasts/${REPORT_DATE}.mp3"

          gh release create "$TAG" "$MP3_FILE" \
            --title "Podcast — ${REPORT_DATE}" \
            --notes "Audio podcast for AI Daily Report ${REPORT_DATE}"

      - name: Update report front matter with podcast URL
        run: |
          REPORT_DATE="${{ steps.date.outputs.report_date }}"
          REPORT_FILE="reports/${REPORT_DATE}.md"
          PODCAST_URL="https://github.com/lankeami/ai-upskill/releases/download/podcast-${REPORT_DATE}/podcast-${REPORT_DATE}.mp3"

          # Use Python for reliable front matter editing
          python3 -c "
          import re, sys

          path = '${REPORT_FILE}'
          with open(path, 'r') as f:
              content = f.read()

          # Match closing --- of front matter (second occurrence)
          parts = content.split('---', 2)
          if len(parts) >= 3:
              front_matter = parts[1]
              rest = parts[2]
              front_matter = front_matter.rstrip('\n') + '\npodcast_url: ${PODCAST_URL}\n'
              content = '---' + front_matter + '---' + rest
              with open(path, 'w') as f:
                  f.write(content)
              print(f'Added podcast_url to {path}')
          else:
              print(f'Error: Could not parse front matter in {path}', file=sys.stderr)
              sys.exit(1)
          "

      - name: Commit updated report
        run: |
          git config user.name "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"
          git add reports/
          if git diff --cached --quiet; then
            echo "No changes to commit"
          else
            git commit -m "chore: add podcast URL to ${{ steps.date.outputs.report_date }} report"
            git push
          fi
```

- [ ] **Step 2: Validate workflow YAML syntax**

Run: `python3 -c "import yaml; yaml.safe_load(open('.github/workflows/podcast.yml'))" && echo "YAML valid"`

Expected: `YAML valid`

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/podcast.yml
git commit -m "feat: add podcast generation GitHub Actions workflow"
```

---

### Task 5: End-to-end local test

- [ ] **Step 1: Authenticate with NotebookLM locally**

Run: `pip install "notebooklm-py[browser]" && playwright install chromium && notebooklm login`

This opens a browser — sign in with your Google account. Credentials are saved to `~/.notebooklm/profiles/default/storage_state.json`.

- [ ] **Step 2: Generate a test podcast**

Run: `python3 scripts/generate-podcast.py --date 2026-04-16`

Expected:
```
Generating podcast for 2026-04-16...
Created notebook: AI Daily Report — 2026-04-16 (...)
Added report content as source
Audio generation started (task: ...)
Audio generation complete
Downloaded podcast to .../podcasts/2026-04-16.mp3

Podcast generated: .../podcasts/2026-04-16.mp3
```

Verify the MP3 file exists and plays correctly.

- [ ] **Step 3: Verify podcasts/ is gitignored**

Run: `git status`
Expected: `podcasts/` directory does NOT appear in untracked files.

- [ ] **Step 4: Run full Jekyll test**

Run: `./scripts/test-jekyll-output.sh`
Expected: PASS

---

### Task 6: Set up GitHub Actions secret

- [ ] **Step 1: Extract auth JSON**

Run: `cat ~/.notebooklm/profiles/default/storage_state.json`

Copy the full JSON content.

- [ ] **Step 2: Add GitHub Actions secret**

Run:
```bash
gh secret set NOTEBOOKLM_AUTH_JSON < ~/.notebooklm/profiles/default/storage_state.json
```

Expected: `✓ Set secret NOTEBOOKLM_AUTH_JSON for lankeami/ai-upskill`
