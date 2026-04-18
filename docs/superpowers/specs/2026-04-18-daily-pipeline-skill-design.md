# Daily Pipeline Skill Design

**Date:** 2026-04-18
**Status:** Draft

## Problem

The current `podcast.yml` GitHub Actions workflow times out because NotebookLM video generation can take 30+ minutes, exceeding GitHub Actions' timeout limits. The workflow has been bumped from 900s to 1800s and still fails.

## Solution

A Claude Code skill (`/daily-pipeline`) that runs the entire pipeline locally: report generation, video generation, GitHub Release creation, front matter update, and git commit/push. The skill orchestrates existing CLI tools — no new application code needed.

## Architecture

### Skill File

**Path:** `.claude/skills/daily-pipeline.md`
**Invocation:** `/daily-pipeline` or `/daily-pipeline 2026-04-15`

The skill is a Claude Code prompt file that instructs Claude to execute 6 sequential steps using existing tools. Claude handles edge cases (idempotency, missing prereqs) through conditional logic described in the prompt.

### Pipeline Steps

```
1. Build Go CLI        →  go build -o ai-report ./cmd/ai-report
2. Generate report     →  ./ai-report generate --date YYYY-MM-DD
3. Generate video      →  python scripts/generate-podcast.py --date YYYY-MM-DD --media-type video
4. Create release      →  gh release create podcast-YYYY-MM-DD ...
5. Update front matter →  Edit reports/YYYY-MM-DD.md to add podcast_url
6. Commit & push       →  git add + commit + push
```

### Step Details

#### Step 1: Build Go CLI

```bash
go build -o ai-report ./cmd/ai-report
```

Rebuilds the CLI binary to ensure it matches the current source. Quick (< 5s).

#### Step 2: Generate Report

```bash
./ai-report generate --date YYYY-MM-DD
```

**Skip condition:** If `reports/YYYY-MM-DD.md` already exists, skip this step and log that the report already exists.

**Prereqs:** `config.yaml` must exist, internet access for fetching news sources.

#### Step 3: Generate Video

```bash
python scripts/generate-podcast.py --date YYYY-MM-DD --media-type video
```

**Skip condition:** If `podcasts/YYYY-MM-DD.mp4` already exists, skip this step.

**Prereqs:**
- `NOTEBOOKLM_AUTH_JSON` environment variable must be set. If missing, stop the pipeline with a clear error message explaining how to set it.
- `notebooklm-py` Python package must be installed. If missing, install it via `pip install notebooklm-py`.
- `reports/YYYY-MM-DD.md` must exist (ensured by step 2).

**Timeout:** This step can take up to 30 minutes. The skill should inform the user that video generation is in progress and may take a while.

**Error handling:** If video generation fails, stop the pipeline. Do not proceed to release creation or commit partial state.

#### Step 4: Create GitHub Release

```bash
gh release create podcast-YYYY-MM-DD podcasts/YYYY-MM-DD.mp4 \
  --title "Podcast — YYYY-MM-DD" \
  --notes "Video podcast for AI Daily Report YYYY-MM-DD"
```

**Skip condition:** If the tag `podcast-YYYY-MM-DD` already exists (check with `gh release view podcast-YYYY-MM-DD`), skip this step.

**Prereqs:** `gh` CLI must be authenticated. `podcasts/YYYY-MM-DD.mp4` must exist.

**Output:** The release asset URL will be:
`https://github.com/lankeami/ai-upskill/releases/download/podcast-YYYY-MM-DD/YYYY-MM-DD.mp4`

#### Step 5: Update Report Front Matter

Edit `reports/YYYY-MM-DD.md` to add or update the `podcast_url` field in the YAML front matter:

```yaml
podcast_url: "https://github.com/lankeami/ai-upskill/releases/download/podcast-YYYY-MM-DD/YYYY-MM-DD.mp4"
```

**Skip condition:** If `podcast_url` is already present in the front matter with the correct URL, skip.

The skill should use the Edit tool to insert the `podcast_url` line before the closing `---` of the front matter.

#### Step 6: Commit & Push

```bash
git add reports/YYYY-MM-DD.md
git commit -m "chore: add podcast URL to YYYY-MM-DD report"
git push
```

If the report was newly generated in step 2, also stage it:
```bash
git add reports/YYYY-MM-DD.md podcasts/  # podcasts/ is gitignored, won't be added
git commit -m "chore: daily AI report and podcast for YYYY-MM-DD"
git push
```

**Note:** The video file itself is NOT committed to git — it's stored only in the GitHub Release. The `podcasts/` directory should be in `.gitignore`.

### Idempotency

Each step checks whether its output already exists before executing. This means the pipeline can be re-run safely after a partial failure — it picks up where it left off:

| State when re-run | Behavior |
|---|---|
| Report exists, no video | Skips report, generates video, creates release, updates front matter |
| Report + video exist, no release | Skips report + video, creates release, updates front matter |
| Everything exists | Skips all steps, no-op |

### Date Parameter

- **Default:** Today's date (`YYYY-MM-DD`)
- **Override:** Pass a date as argument: `/daily-pipeline 2026-04-15`
- The skill parses the first argument as the date. If no argument, uses today.

## GitHub Actions Changes

Modify `.github/workflows/podcast.yml` to remove the automatic push trigger, keeping only manual dispatch:

**Before:**
```yaml
on:
  push:
    branches: [main]
    paths: ['reports/*.md']
  workflow_dispatch:
    inputs:
      ...
```

**After:**
```yaml
on:
  workflow_dispatch:
    inputs:
      ...
```

This makes the workflow a manual fallback only.

## Scheduling

The skill documents two scheduling options. The user picks at setup time.

### Option A: Claude Code Remote Triggers

Use the `/schedule` skill to create a cron-based remote agent:

```
/schedule create "daily-pipeline" --cron "0 8 * * *" --prompt "run /daily-pipeline"
```

This runs the pipeline daily at 8am UTC via Claude Code's scheduling infrastructure.

### Option B: Local Crontab / launchd

**crontab (Linux/macOS):**
```bash
0 8 * * * cd /path/to/ai-upskill && claude -p "run /daily-pipeline"
```

**launchd (macOS):**
Create `~/Library/LaunchAgents/com.ai-upskill.daily-pipeline.plist` with:
- Program: `/path/to/claude`
- Arguments: `-p "run /daily-pipeline"`
- Working directory: `/path/to/ai-upskill`
- StartCalendarInterval: Hour 8, Minute 0

Both options are documented in the skill file itself for easy reference.

## Prerequisites

The skill will check these prerequisites at the start and fail fast with clear messages:

1. **Go** — `go version` must succeed
2. **Python 3** — `python3 --version` must succeed
3. **`notebooklm-py`** — install via pip if missing
4. **`gh` CLI** — `gh auth status` must succeed
5. **`NOTEBOOKLM_AUTH_JSON`** — environment variable must be set
6. **Git** — working directory must be clean (no uncommitted changes)

## Files Changed

| File | Change |
|---|---|
| `.claude/skills/daily-pipeline.md` | **New** — skill prompt file |
| `.github/workflows/podcast.yml` | **Modified** — remove push trigger |
| `.gitignore` | **No change needed** — `podcasts/` is already listed |

## Out of Scope

- Changing the Go CLI or Python script behavior
- Audio-only pipeline (video is the default; audio can be added later as a skill argument)
- Automatic retry on video generation failure
- Notification/alerting on success or failure
