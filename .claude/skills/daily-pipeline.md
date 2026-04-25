---
name: daily-pipeline
description: Run the full daily AI report pipeline — generate report, create video, upload to YouTube, update front matter, commit & push
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

0. **Load .env:** If a `.env` file exists in the project root, source it: `set -a && source .env && set +a`. This loads `NOTEBOOKLM_AUTH_JSON`, `YOUTUBE_CLIENT_ID`, `YOUTUBE_CLIENT_SECRET`, and any other env vars.
1. **Activate venv:** Run `source .venv/bin/activate`. If `.venv/` doesn't exist, create it first: `python3 -m venv .venv && source .venv/bin/activate`.
2. **Go:** Run `go version`. If it fails, tell the user to install Go.
3. **Python 3:** Run `python3 --version`. If it fails, tell the user to install Python 3.
4. **notebooklm-py:** Run `python3 -c "import notebooklm"`. If it fails, run `pip install notebooklm-py`.
5. **google-api-python-client:** Run `python3 -c "import googleapiclient"`. If it fails, run `pip install google-api-python-client google-auth-oauthlib`.
6. **NOTEBOOKLM_AUTH_JSON:** Check if the environment variable is set. If not, stop and tell the user:
   > Set the `NOTEBOOKLM_AUTH_JSON` environment variable with your NotebookLM credentials before running this skill.
7. **YOUTUBE_CLIENT_ID / YOUTUBE_CLIENT_SECRET:** Check both are set. If not, stop and tell the user:
   > Set `YOUTUBE_CLIENT_ID` and `YOUTUBE_CLIENT_SECRET` in `.env`. See `docs/youtube-setup.md` for instructions.
8. **.youtube-token.json:** Check the file exists in the project root. If not, stop and tell the user:
   > Run `python scripts/upload-youtube.py --auth` to complete one-time YouTube authentication.
9. **Clean git state:** Run `git status --porcelain`. If there is output, stop and tell the user to commit or stash their changes first.

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

If not, generate the video in three phases:

### Step 3a: Start Generation

Run:
```bash
python scripts/generate-podcast.py start --date DATE --media-type video
```

If this command fails, stop the pipeline and show the error.

Tell the user: "Video generation started. Polling for completion (this typically takes 10-30 minutes)."

### Step 3b: Poll for Completion

Run this command in a loop, waiting 30 seconds between each attempt:
```bash
python scripts/generate-podcast.py poll
```

- **Exit code 0** (prints `complete`): generation is done — proceed to Step 3c.
- **Exit code 1** (prints `pending` or `in_progress`): still working — wait 30 seconds and poll again.
- **Exit code 2** (prints error): generation failed — stop the pipeline and show the error.

Maximum 40 poll attempts (~20 minutes). If exceeded, stop the pipeline with: "Video generation timed out after 20 minutes of polling. The state file `.podcast-state.json` has been preserved for manual recovery."

Between polls, tell the user the current status (e.g., "Poll 5/40: in_progress").

### Step 3c: Download

Run:
```bash
python scripts/generate-podcast.py download
```

If this command fails, stop the pipeline and show the error.

Verify that `podcasts/DATE.mp4` was created.

## Step 4: Upload to YouTube

**Check:** Does `reports/DATE.md` already contain a `youtube_url` in the YAML front matter? If yes, skip this step and note "YouTube upload already done for DATE".

If not, run:
```bash
python scripts/upload-youtube.py --date DATE
```

Capture the last line of stdout — this is the YouTube URL (format: `https://youtube.com/watch?v=VIDEO_ID`). Store it as `YOUTUBE_URL`.

If the command fails, stop the pipeline and show the full error output.

## Step 5: Update Report Front Matter

**Check:** Read `reports/DATE.md` and check if `youtube_url` is already present in the YAML front matter. If it already contains the correct URL, skip this step.

If not, use the Edit tool to add the following line to the YAML front matter (before the closing `---`):

```
youtube_url: "YOUTUBE_URL"
```

Where `YOUTUBE_URL` is the URL captured in Step 4.

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
- YouTube: uploaded or already existed
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
