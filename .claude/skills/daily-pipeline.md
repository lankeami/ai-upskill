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

Before running the pipeline, first load environment variables, then verify all prerequisites. Stop immediately with a clear error if any check fails:

0. **Load .env:** If a `.env` file exists in the project root, source it: `set -a && source .env && set +a`. This loads `NOTEBOOKLM_AUTH_JSON` and any other env vars.
1. **Activate venv:** Run `source .venv/bin/activate`. If `.venv/` doesn't exist, create it first: `python3 -m venv .venv && source .venv/bin/activate`.
2. **Go:** Run `go version`. If it fails, tell the user to install Go.
3. **Python 3:** Run `python3 --version`. If it fails, tell the user to install Python 3.
4. **notebooklm-py:** Run `python3 -c "import notebooklm"`. If it fails, run `pip install notebooklm-py`.
5. **gh CLI:** Run `gh auth status`. If it fails, tell the user to run `gh auth login`.
6. **NOTEBOOKLM_AUTH_JSON:** Check if the environment variable is set. If not, stop and tell the user:
   > Set the `NOTEBOOKLM_AUTH_JSON` environment variable with your NotebookLM credentials before running this skill.
7. **Clean git state:** Run `git status --porcelain`. If there is output, stop and tell the user to commit or stash their changes first.

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
```

Use the appropriate commit message:
- If the report was **newly generated** in Step 2: `git commit -m "chore: daily AI report and podcast for DATE"`
- If the report **already existed** and only the front matter was updated: `git commit -m "chore: add podcast URL to DATE report"`

Then push:
```bash
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

This skill can be scheduled to run automatically. Three options:

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
