# YouTube Upload Pipeline Design

**Date:** 2026-04-24
**Status:** Draft

## Problem

NotebookLM-generated videos are currently published as GitHub Release assets. GitHub serves these with `Content-Type: application/octet-stream` and `Content-Disposition: attachment`, which prevents browser-native video playback on report pages. YouTube provides proper streaming, a public channel presence, and a working embedded player.

## Solution

Add a new `scripts/upload-youtube.py` script and a new Step 4 in the daily-pipeline skill to upload each day's video to the "Toiletpaper Press" YouTube channel. Replace GitHub Releases as the video host. Update report front matter to use `youtube_url` instead of `podcast_url`. Embed the YouTube player in the report layout.

---

## Script: `scripts/upload-youtube.py`

### Interface

```bash
# Upload video for a specific date
python scripts/upload-youtube.py --date 2026-04-22

# One-time browser auth flow
python scripts/upload-youtube.py --auth
```

### Dependencies

```
google-api-python-client
google-auth-oauthlib
```

Install via: `pip install google-api-python-client google-auth-oauthlib`

### Auth

- `YOUTUBE_CLIENT_ID` and `YOUTUBE_CLIENT_SECRET` stored in `.env`
- OAuth2 token persisted to `.youtube-token.json` (gitignored) after the first `--auth` run
- Subsequent runs use the saved refresh token silently — no browser required

### Upload Behavior

1. Reads `podcasts/DATE.mp4`
2. Reads `reports/DATE.md`, extracts the first 3–5 bullet points (stripping YAML front matter) for the description
3. Uploads to YouTube with:
   - **Title:** `Toiletpaper Press — AI Daily YYYY-MM-DD`
   - **Description:** First 3–5 bullet points from the report + `\n\nFull report: https://lankeami.github.io/ai-upskill/reports/YYYY-MM-DD`
   - **Tags:** `["AI news", "daily AI report", "Toiletpaper Press", "artificial intelligence"]`
   - **Category ID:** `28` (Science & Technology)
   - **Privacy:** `public`
4. Prints the YouTube URL to stdout: `https://youtube.com/watch?v=VIDEO_ID`
5. Exits 0 on success, 1 on error

### Idempotency

Before uploading, the script lists the channel's videos and checks for an existing video with the exact title. If found, it prints the existing URL and exits 0 without re-uploading.

---

## Pipeline Changes

### Removed

**Step 4: Create GitHub Release** — eliminated. GitHub Releases are no longer used for video hosting.

### Added

**New Step 4: Upload to YouTube**

```bash
python scripts/upload-youtube.py --date DATE
```

Capture stdout to get the YouTube URL. If the command fails, stop the pipeline and show the error.

### Modified

**Step 5: Update Report Front Matter**

Replace `podcast_url` with `youtube_url`:

```yaml
youtube_url: "https://youtube.com/watch?v=VIDEO_ID"
```

**Step 6: Commit & Push** — unchanged, references `youtube_url` field instead of `podcast_url`.

### Updated pipeline order

1. Build Go CLI
2. Generate Report
3. Generate Video (start/poll/download — unchanged)
4. Upload to YouTube *(new)*
5. Update Report Front Matter with `youtube_url` *(modified)*
6. Commit & Push

---

## Report Layout Changes

In `_layouts/report.html`, replace the `<video>` element with a YouTube iframe embed.

The `youtube_url` front matter field stores the watch URL (`https://youtube.com/watch?v=VIDEO_ID`). The template extracts the video ID and builds the embed URL.

```html
{% if page.youtube_url %}
<div class="podcast-player" style="margin-bottom: 1.5rem; padding: 1rem; background: #f6f8fa; border-radius: 6px;">
  <h3 style="font-size: 0.9rem; margin-bottom: 0.5rem; color: #656d76; text-transform: uppercase; letter-spacing: 0.05em;">Watch this report</h3>
  {% assign yt_id = page.youtube_url | split: "v=" | last %}
  <iframe width="100%" height="315" style="border-radius: 4px; border: none; margin-bottom: 0.5rem;"
    src="https://www.youtube.com/embed/{{ yt_id }}"
    allowfullscreen loading="lazy">
  </iframe>
  <p style="margin: 0; font-size: 0.85rem;"><a href="{{ page.youtube_url }}">Watch on YouTube</a></p>
</div>
{% endif %}
```

---

## Prerequisites & Setup

Documented in `docs/youtube-setup.md`:

### 1. Create a Google Cloud project

1. Go to [console.cloud.google.com](https://console.cloud.google.com) → New Project
2. Name it (e.g., `ai-upskill`)
3. Go to **APIs & Services → Library** → search "YouTube Data API v3" → Enable

### 2. Create OAuth2 credentials

1. Go to **APIs & Services → Credentials → Create Credentials → OAuth 2.0 Client ID**
2. Application type: **Desktop app**
3. Download the JSON — note `client_id` and `client_secret`
4. Go to **OAuth consent screen** → set to External → add your Gmail as a test user

### 3. Add credentials to `.env`

```
YOUTUBE_CLIENT_ID=your_client_id
YOUTUBE_CLIENT_SECRET=your_client_secret
```

### 4. Install dependencies

```bash
source .venv/bin/activate
pip install google-api-python-client google-auth-oauthlib
```

### 5. Run the one-time auth flow

```bash
source .venv/bin/activate
python scripts/upload-youtube.py --auth
```

This opens a browser. Sign in as `lankeami@gmail.com`, grant the YouTube upload permission, and the token is saved to `.youtube-token.json`. This only needs to be done once.

---

## Prerequisites Check (daily-pipeline skill additions)

Add to the Prerequisites Check section:

- **YOUTUBE_CLIENT_ID / YOUTUBE_CLIENT_SECRET:** Check both env vars are set. If not, tell the user to follow `docs/youtube-setup.md`.
- **google-api-python-client:** Run `python3 -c "import googleapiclient"`. If it fails, run `pip install google-api-python-client google-auth-oauthlib`.
- **.youtube-token.json:** Check it exists. If not, tell the user to run `python scripts/upload-youtube.py --auth` first.

---

## Error Handling

| Scenario | Behavior |
|----------|----------|
| `podcasts/DATE.mp4` missing | Exit 1 with error |
| `reports/DATE.md` missing | Exit 1 with error |
| Token file missing | Exit 1: "Run `python scripts/upload-youtube.py --auth` first" |
| Upload API error | Exit 1 with full error message |
| Video already uploaded (idempotency) | Print existing URL, exit 0 |
| Quota exceeded (10,000 units/day) | Exit 1 with clear quota error message |

---

## Files Changed

| File | Action |
|------|--------|
| `scripts/upload-youtube.py` | Create |
| `docs/youtube-setup.md` | Create |
| `.claude/skills/daily-pipeline.md` | Modify: remove Step 4 (GitHub Release), add YouTube upload, update prerequisites |
| `_layouts/report.html` | Modify: replace `<video>` with YouTube iframe |
| `.gitignore` | Modify: add `.youtube-token.json` |
