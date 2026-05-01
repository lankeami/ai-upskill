---
name: daily-pipeline
description: Run the full daily AI report pipeline — generate report, create audio podcast, publish as GitHub Release, commit & push
arguments:
  - name: date
    description: Report date in YYYY-MM-DD format (default: today)
    required: false
---

# Daily Pipeline

Run the full daily pipeline for the ai-upskill project. This generates a daily AI report, creates an audio podcast via NotebookLM, publishes it as a GitHub Release, and commits everything.

## Date

Use the provided `$ARGUMENTS` as the target date. If no argument was provided, use today's date in `YYYY-MM-DD` format.

Store the resolved date — all subsequent steps reference it as `DATE`.

## Prerequisites Check

Before running the pipeline, first load environment variables, then verify all prerequisites. Stop immediately with a clear error if any check fails:

0. **Load .env:** If a `.env` file exists in the project root, source it: `set -a && source .env && set +a`. This loads `NOTEBOOKLM_AUTH_JSON` and any other env vars.
1. **Activate venv:** Run `source .venv/bin/activate`. If `.venv/` doesn't exist, create it first: `python3 -m venv .venv && source .venv/bin/activate`.
2. **Go:** Run `go version`. If it fails, tell the user to install Go.
3. **Python 3:** Run `python3 --version`. If it fails, tell the user to install Python 3.
4. **notebooklm-py / pyyaml:** Run `python3 -c "import notebooklm, yaml"`. If it fails, run `pip install notebooklm-py pyyaml`.
5. **NOTEBOOKLM_AUTH_JSON:** Check if the environment variable is set. If not, stop and tell the user:
   > Set the `NOTEBOOKLM_AUTH_JSON` environment variable with your NotebookLM credentials before running this skill.
6. **Clean git state:** Run `git status --porcelain`. If there is output, stop and tell the user to commit or stash their changes first.

## Step 0: Sync Main

Run:
```bash
git pull origin main
```

If this fails (e.g. merge conflict, diverged branches), stop and tell the user to resolve the conflict manually before running the pipeline.

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

## Step 3: Generate Audio

**Check:** Does `podcasts/DATE.mp3` already exist? If yes, skip this step and note "Audio already exists for DATE".

If not, generate the audio in three phases:

### Step 3a: Start Generation

Run:
```bash
python scripts/generate-podcast.py start --date DATE --media-type audio
```

If this command fails, stop the pipeline and show the error.

Tell the user: "Audio generation started. Polling for completion (this typically takes 5-10 minutes)."

### Step 3b: Poll for Completion

Run this command in a loop, waiting 30 seconds between each attempt:
```bash
python scripts/generate-podcast.py poll
```

- **Exit code 0** (prints `complete`): generation is done — proceed to Step 3c.
- **Exit code 1** (prints `pending` or `in_progress`): still working — wait 30 seconds and poll again.
- **Exit code 2** (prints error): generation failed — stop the pipeline and show the error.

Maximum 40 poll attempts (~20 minutes). If exceeded, stop the pipeline with: "Audio generation timed out after 20 minutes of polling. The state file `.podcast-state.json` has been preserved for manual recovery."

Between polls, tell the user the current status (e.g., "Poll 5/40: in_progress").

### Step 3c: Download

Run:
```bash
python scripts/generate-podcast.py download
```

If this command fails, stop the pipeline and show the error.

Verify that `podcasts/DATE.mp3` was created.

## Step 4: Publish GitHub Release

**Check:** Run `gh release view podcast-DATE 2>/dev/null >/dev/null && echo "exists" || echo "missing"`. If it prints `exists`, skip this step and note "Release already exists for DATE".

If not, run:
```bash
gh release create podcast-DATE podcasts/DATE.mp3 \
  --title "Podcast — ${DATE}" \
  --notes "Audio podcast for AI Daily Report ${DATE}"
```

If this command fails, stop the pipeline and show the error.

The asset URL is deterministic — no need to parse output:
```
PODCAST_URL=https://github.com/lankeami/ai-upskill/releases/download/podcast-${DATE}/${DATE}.mp3
```

**Check:** Run `grep -q 'podcast_url:' reports/DATE.md && echo "found" || echo "missing"`. If it prints `found`, skip the injection and note "Front matter already has podcast_url for DATE".

If not, inject it using Python:
```bash
python3 -c "
path = 'reports/DATE.md'
url = 'PODCAST_URL'
content = open(path).read()
parts = content.split('---', 2)
if len(parts) >= 3:
    parts[1] = parts[1].rstrip('\n') + '\npodcast_url: \"' + url + '\"\n'
    open(path, 'w').write('---'.join(parts))
    print('Added podcast_url to', path)
else:
    import sys; print('Error: could not parse front matter', file=sys.stderr); sys.exit(1)
"
```

If the Python command fails, stop the pipeline and show the error.

## Step 5: Commit & Push

**Check:** Run `git status --porcelain`. If there are no changes, note "Nothing to commit" and finish.

If there are changes:

```bash
git add reports/DATE.md
```

Use the appropriate commit message:
- If the report was **newly generated** in Step 2: `git commit -m "chore: daily AI report for DATE"`
- If the report **already existed** and only the front matter was updated: `git commit -m "chore: add podcast URL to DATE report"`

Then push:
```bash
git push
```

## Done

Summarize what was done:
- Sync: pulled latest main
- Report: generated or already existed
- Audio: generated or already existed
- GitHub Release: published or already existed
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
0 8 * * * claude -p "run /daily-pipeline"
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
