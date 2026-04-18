# Daily Pipeline Skill Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create a Claude Code skill that runs the full daily pipeline locally — report generation, video generation, GitHub Release, front matter update, commit & push.

**Architecture:** A single skill prompt file (`.claude/skills/daily-pipeline.md`) that instructs Claude to execute existing CLI tools in sequence. One workflow file modification to disable automatic triggers. No new application code.

**Tech Stack:** Claude Code skills (Markdown prompt), GitHub CLI (`gh`), Go CLI (`ai-report`), Python (`generate-podcast.py`)

---

## File Structure

| File | Action | Responsibility |
|---|---|---|
| `.claude/skills/daily-pipeline.md` | Create | Skill prompt — orchestrates the 6-step pipeline |
| `.github/workflows/podcast.yml` | Modify | Remove push trigger, keep manual dispatch only |

---

### Task 1: Create the Skill Prompt File

**Files:**
- Create: `.claude/skills/daily-pipeline.md`

- [ ] **Step 1: Create the `.claude/skills/` directory**

Run:
```bash
mkdir -p .claude/skills
```

- [ ] **Step 2: Write the skill file**

Create `.claude/skills/daily-pipeline.md` with the following content:

````markdown
---
name: daily-pipeline
description: Run the full daily AI report pipeline — generate report, create video, publish release, update front matter, commit & push
arguments:
  - name: date
    description: Report date in YYYY-MM-DD format (default: today)
    required: false
---

# Daily Pipeline

Run the full daily pipeline for the ai-upskill project. This generates a daily AI report, creates a video podcast via NotebookLM, publishes it as a GitHub Release, updates the report front matter with the video URL, and commits everything.

## Date

Use the provided `$ARGUMENTS` as the target date. If no argument was provided, use today's date in `YYYY-MM-DD` format.

Store the resolved date — all subsequent steps reference it as `DATE`.

## Prerequisites Check

Before running the pipeline, verify all prerequisites. Run these checks and stop immediately with a clear error if any fail:

1. **Go:** Run `go version`. If it fails, tell the user to install Go.
2. **Python 3:** Run `python3 --version`. If it fails, tell the user to install Python 3.
3. **notebooklm-py:** Run `python3 -c "import notebooklm"`. If it fails, run `pip install notebooklm-py`.
4. **gh CLI:** Run `gh auth status`. If it fails, tell the user to run `gh auth login`.
5. **NOTEBOOKLM_AUTH_JSON:** Check if the environment variable is set. If not, stop and tell the user:
   > Set the `NOTEBOOKLM_AUTH_JSON` environment variable with your NotebookLM credentials before running this skill.
6. **Clean git state:** Run `git status --porcelain`. If there is output, stop and tell the user to commit or stash their changes first.

## Step 1: Build Go CLI

Run:
```bash
go build -o ai-report ./cmd/ai-report
```

If the build fails, stop and show the error.

## Step 2: Generate Report

**Check:** Does `reports/DATE.md` already exist? If yes, skip this step and note "Report already exists for DATE".

If not, run:
```bash
./ai-report generate --date DATE
```

Verify that `reports/DATE.md` was created. If not, stop and show the error.

## Step 3: Generate Video

**Check:** Does `podcasts/DATE.mp4` already exist? If yes, skip this step and note "Video already exists for DATE".

If not, run:
```bash
python scripts/generate-podcast.py --date DATE --media-type video
```

Tell the user: "Video generation is in progress. This typically takes 10-30 minutes."

Run this command with a timeout of at least 2400 seconds (40 minutes).

If the command fails, stop the pipeline. Do NOT proceed to release creation or commit. Show the full error output.

Verify that `podcasts/DATE.mp4` was created.

## Step 4: Create GitHub Release

**Check:** Run `gh release view podcast-DATE 2>/dev/null`. If it succeeds (exit code 0), the release already exists — skip this step and note "Release already exists for DATE".

If not, run:
```bash
gh release create "podcast-DATE" "podcasts/DATE.mp4" \
  --title "Podcast — DATE" \
  --notes "Video podcast for AI Daily Report DATE"
```

If this fails, stop and show the error.

## Step 5: Update Report Front Matter

**Check:** Read `reports/DATE.md` and check if `podcast_url` is already present in the YAML front matter. If it already contains the correct URL, skip this step.

If not, use the Edit tool to add the following line to the YAML front matter (before the closing `---`):

```
podcast_url: "https://github.com/lankeami/ai-upskill/releases/download/podcast-DATE/DATE.mp4"
```

## Step 6: Commit & Push

**Check:** Run `git status --porcelain`. If there are no changes, note "Nothing to commit" and finish.

If there are changes:

```bash
git add reports/DATE.md
git commit -m "chore: daily AI report and podcast for DATE"
git push
```

## Done

Summarize what was done:
- Report: generated or already existed
- Video: generated or already existed
- Release: created or already existed
- Front matter: updated or already correct
- Commit: pushed or nothing to commit

## Scheduling (Reference)

This skill can be scheduled to run automatically. Two options:

### Option A: Claude Code Remote Triggers
```
/schedule create "daily-pipeline" --cron "0 8 * * *" --prompt "run /daily-pipeline"
```

### Option B: Local Crontab
```bash
# crontab -e
0 8 * * * cd /Users/jaychinthrajah/workspaces/_personal_/ai-upskill && claude -p "run /daily-pipeline"
```

### Option C: macOS launchd
Create `~/Library/LaunchAgents/com.ai-upskill.daily-pipeline.plist`:
```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.ai-upskill.daily-pipeline</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/claude</string>
        <string>-p</string>
        <string>run /daily-pipeline</string>
    </array>
    <key>WorkingDirectory</key>
    <string>/Users/jaychinthrajah/workspaces/_personal_/ai-upskill</string>
    <key>StartCalendarInterval</key>
    <dict>
        <key>Hour</key>
        <integer>8</integer>
        <key>Minute</key>
        <integer>0</integer>
    </dict>
    <key>StandardOutPath</key>
    <string>/tmp/ai-upskill-pipeline.log</string>
    <key>StandardErrorPath</key>
    <string>/tmp/ai-upskill-pipeline.err</string>
</dict>
</plist>
```
Load with: `launchctl load ~/Library/LaunchAgents/com.ai-upskill.daily-pipeline.plist`
````

- [ ] **Step 3: Verify the skill file was created**

Run:
```bash
cat .claude/skills/daily-pipeline.md | head -5
```

Expected output:
```
---
name: daily-pipeline
description: Run the full daily AI report pipeline — generate report, create video, publish release, update front matter, commit & push
arguments:
  - name: date
```

- [ ] **Step 4: Commit**

```bash
git add .claude/skills/daily-pipeline.md
git commit -m "feat: add daily-pipeline skill for local report+video generation"
```

---

### Task 2: Disable Automatic Trigger on podcast.yml

**Files:**
- Modify: `.github/workflows/podcast.yml:1-7`

- [ ] **Step 1: Remove the push trigger**

Edit `.github/workflows/podcast.yml` — replace the `on:` block:

Before:
```yaml
on:
  push:
    branches: [main]
    paths: ['reports/*.md']
  workflow_dispatch:
    inputs:
      date:
        description: 'Report date (YYYY-MM-DD)'
        required: true
      media_type:
        description: 'Media type to generate'
        required: false
        default: 'video'
        type: choice
        options:
          - video
          - audio
```

After:
```yaml
on:
  workflow_dispatch:
    inputs:
      date:
        description: 'Report date (YYYY-MM-DD)'
        required: true
      media_type:
        description: 'Media type to generate'
        required: false
        default: 'video'
        type: choice
        options:
          - video
          - audio
```

- [ ] **Step 2: Verify the workflow file**

Run:
```bash
head -20 .github/workflows/podcast.yml
```

Expected: The `on:` section should only contain `workflow_dispatch`, no `push` trigger.

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/podcast.yml
git commit -m "chore: disable automatic podcast trigger, keep as manual fallback"
```

---

### Task 3: Test the Skill Invocation

- [ ] **Step 1: Verify skill is discoverable**

Run:
```bash
claude -p "list skills" 2>&1 | grep -i daily-pipeline
```

Expected: `daily-pipeline` should appear in the skill list.

- [ ] **Step 2: Dry-run the prereq checks**

Invoke `/daily-pipeline` and verify that the prerequisites check runs correctly. The pipeline should reach at least the "Build Go CLI" step before you stop it (since you may not want to actually generate a video during testing).

- [ ] **Step 3: Commit all remaining changes (if any)**

If there were any adjustments needed:
```bash
git add -A
git commit -m "fix: adjust daily-pipeline skill after testing"
```
