---
name: fetch-media
description: Poll NotebookLM for the latest audio and video files, download them, then run the release step to publish a GitHub Release and update the report front matter
arguments:
  - name: date
    description: Report date in YYYY-MM-DD format (default: today)
    required: false
---

# Fetch Media

Poll NotebookLM for an in-progress audio or video generation, download the completed artifact, then publish it as a GitHub Release and update the report's front matter with the asset URL.

## Date

Use the provided `$ARGUMENTS` as the target date. If no argument was provided, use today's date in `YYYY-MM-DD` format.

Store the resolved date — all subsequent steps reference it as `DATE`.

## Prerequisites Check

Before running, first load environment variables, then verify all prerequisites. Stop immediately with a clear error if any check fails:

0. **Load .env:** If a `.env` file exists in the project root, source it: `set -a && source .env && set +a`.
1. **Activate venv:** Run `source .venv/bin/activate`. If `.venv/` doesn't exist, stop and tell the user to create it first: `python3 -m venv .venv && pip install notebooklm-py pyyaml`.
2. **Python 3:** Run `python3 --version`. If it fails, tell the user to install Python 3.
3. **notebooklm-py / pyyaml:** Run `python3 -c "import notebooklm, yaml"`. If it fails, run `pip install notebooklm-py pyyaml`.
4. **NOTEBOOKLM_AUTH_JSON:** Check if the environment variable is set. If not, stop and tell the user:
   > Set the `NOTEBOOKLM_AUTH_JSON` environment variable with your NotebookLM credentials before running this skill.
5. **State file:** Check if `.podcast-state.json` exists. If it does not exist, stop and tell the user:
   > No in-progress NotebookLM generation found. Run `/daily-pipeline` to start one, or use `python scripts/generate-podcast.py start --date DATE --media-type audio` (or `video`) to kick one off manually.

Read `.podcast-state.json` to determine the `media_type` (`audio` or `video`) and confirm the `date` matches DATE. If the dates do not match, warn the user:
> Warning: State file is for a different date (`STATE_DATE`). Proceeding with the state file's date instead of DATE.

Update DATE to match the state file's date.

## Step 1: Poll for Completion

Run this command in a loop, waiting 30 seconds between each attempt:
```bash
python scripts/generate-podcast.py poll
```

- **Exit code 0** (prints `complete`): generation is done — proceed to Step 2.
- **Exit code 1** (prints `pending` or `in_progress`): still working — wait 30 seconds and poll again.
- **Exit code 2** (prints error): generation failed — stop and show the error.

Maximum 60 poll attempts (~30 minutes). If exceeded, stop with:
> Generation timed out after 30 minutes of polling. The state file `.podcast-state.json` has been preserved for manual recovery.

Between polls, tell the user the current status, e.g. "Poll 5/60: in_progress".

## Step 2: Download

Run:
```bash
python scripts/generate-podcast.py download
```

If this command fails, stop and show the error.

The expected output file path depends on `media_type`:
- `audio` → `podcasts/DATE.mp3`
- `video` → `podcasts/DATE.mp4`

Verify the expected file was created. If not, stop and show the error.

Tell the user which file was downloaded.

## Step 3: Publish GitHub Release

**Check:** Run `gh release view podcast-DATE 2>/dev/null >/dev/null && echo "exists" || echo "missing"`. If it prints `exists`, skip this step and note "Release already exists for DATE".

If not, determine the assets to attach based on which files exist:
- If `podcasts/DATE.mp3` exists, include it.
- If `podcasts/DATE.mp4` exists, include it.

Run:
```bash
gh release create podcast-DATE [FILES] \
  --title "Podcast — DATE" \
  --notes "Audio/video podcast for AI Daily Report DATE"
```

Where `[FILES]` is the space-separated list of existing podcast files for DATE.

If this command fails, stop and show the error.

The asset URLs are deterministic:
```
AUDIO_URL=https://github.com/lankeami/ai-upskill/releases/download/podcast-DATE/DATE.mp3
VIDEO_URL=https://github.com/lankeami/ai-upskill/releases/download/podcast-DATE/DATE.mp4
```

## Step 4: Update Report Front Matter

For each asset URL that was published, inject it into `reports/DATE.md` if not already present.

**Check for audio URL:**
Run `grep -q 'podcast_url:' reports/DATE.md && echo "found" || echo "missing"`.

If `missing` and `podcasts/DATE.mp3` was published, inject using Python:
```bash
python3 -c "
path = 'reports/DATE.md'
url = 'AUDIO_URL'
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

**Check for video URL:**
Run `grep -q 'video_url:' reports/DATE.md && echo "found" || echo "missing"`.

If `missing` and `podcasts/DATE.mp4` was published, inject using Python:
```bash
python3 -c "
path = 'reports/DATE.md'
url = 'VIDEO_URL'
content = open(path).read()
parts = content.split('---', 2)
if len(parts) >= 3:
    parts[1] = parts[1].rstrip('\n') + '\nvideo_url: \"' + url + '\"\n'
    open(path, 'w').write('---'.join(parts))
    print('Added video_url to', path)
else:
    import sys; print('Error: could not parse front matter', file=sys.stderr); sys.exit(1)
"
```

If either Python command fails, stop and show the error.

## Step 5: Commit & Push

**Check:** Run `git status --porcelain`. If there are no changes, note "Nothing to commit" and finish.

If there are changes:
```bash
git add reports/DATE.md
git commit -m "chore: add podcast URL to DATE report"
git push
```

If the commit or push fails, stop and show the error.

## Done

Summarize what was done:
- Poll: number of polls before completion
- Download: file(s) downloaded
- GitHub Release: published or already existed
- Front matter: fields injected or already correct
- Commit: pushed or nothing to commit
