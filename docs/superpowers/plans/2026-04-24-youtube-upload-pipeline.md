# YouTube Upload Pipeline Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a YouTube upload step to the daily pipeline, replacing GitHub Releases as the video host, so videos are embedded on report pages via a working YouTube iframe player.

**Architecture:** New `scripts/upload-youtube.py` handles OAuth2 auth (one-time browser flow, token persisted to `.youtube-token.json`) and YouTube Data API v3 uploads. The daily-pipeline skill drops the GitHub Release step and adds a YouTube upload step that captures the returned video URL. The report layout swaps the broken `<video>` element for a YouTube iframe embed using the `youtube_url` front matter field.

**Tech Stack:** Python 3, `google-api-python-client`, `google-auth-oauthlib`, YouTube Data API v3, Jekyll/Liquid templates

**Spec:** `docs/superpowers/specs/2026-04-24-youtube-upload-pipeline-design.md`

---

## File Structure

| File | Action | Responsibility |
|------|--------|----------------|
| `scripts/upload-youtube.py` | Create | OAuth2 auth flow, idempotency check, upload with metadata, print YouTube URL |
| `docs/youtube-setup.md` | Create | Step-by-step Google Cloud + OAuth setup instructions |
| `.gitignore` | Modify | Add `.youtube-token.json` |
| `_layouts/report.html` | Modify | Replace `podcast_url` video element with `youtube_url` iframe embed |
| `.claude/skills/daily-pipeline.md` | Modify | Remove GitHub Release step, add YouTube upload step, update prerequisites and summary |
| `tests/test_upload_youtube.py` | Create | Unit tests for description extraction, idempotency logic, metadata building |

---

### Task 1: Add `.youtube-token.json` to `.gitignore`

**Files:**
- Modify: `.gitignore`

- [ ] **Step 1: Append the token file to `.gitignore`**

Add `.youtube-token.json` as a new line in `.gitignore` (after the existing `.podcast-state.json` entry):

```
.youtube-token.json
```

- [ ] **Step 2: Commit**

```bash
git add .gitignore
git commit -m "chore: add .youtube-token.json to .gitignore"
```

---

### Task 2: Create `docs/youtube-setup.md`

**Files:**
- Create: `docs/youtube-setup.md`

- [ ] **Step 1: Create the setup documentation file**

Create `docs/youtube-setup.md` with the following content:

```markdown
# YouTube Upload Setup

One-time setup to enable the daily pipeline to upload videos to the Toiletpaper Press YouTube channel.

## 1. Create a Google Cloud project

1. Go to [console.cloud.google.com](https://console.cloud.google.com) → **New Project**
2. Name it (e.g., `ai-upskill`) and create it
3. Go to **APIs & Services → Library** → search for **YouTube Data API v3** → **Enable**

## 2. Create OAuth2 credentials

1. Go to **APIs & Services → Credentials → Create Credentials → OAuth 2.0 Client ID**
2. If prompted, configure the OAuth consent screen first:
   - User type: **External**
   - Add `lankeami@gmail.com` as a test user
3. Back in Credentials → Create Credentials → OAuth 2.0 Client ID:
   - Application type: **Desktop app**
   - Name: anything (e.g., `ai-upskill-upload`)
4. Click **Download JSON** — open it and note `client_id` and `client_secret`

## 3. Add credentials to `.env`

Add the following to your `.env` file in the project root:

```
YOUTUBE_CLIENT_ID=your_client_id_here
YOUTUBE_CLIENT_SECRET=your_client_secret_here
```

## 4. Install dependencies

```bash
source .venv/bin/activate
pip install google-api-python-client google-auth-oauthlib
```

## 5. Run the one-time auth flow

```bash
source .venv/bin/activate
python scripts/upload-youtube.py --auth
```

This opens a browser window. Sign in as `lankeami@gmail.com` and grant the YouTube upload permission. The token is saved to `.youtube-token.json` in the project root. **This only needs to be done once.** The token auto-refreshes on subsequent runs.

## Troubleshooting

- **"Access blocked"**: The OAuth app is in test mode. Make sure `lankeami@gmail.com` is listed as a test user in the OAuth consent screen.
- **"Token file missing"**: Run `python scripts/upload-youtube.py --auth` again.
- **Quota errors**: The YouTube Data API allows ~6 uploads per day (10,000 quota units, ~1,600 per upload). If you hit the limit, wait 24 hours.
```

- [ ] **Step 2: Add `docs/youtube-setup.md` to Jekyll exclude list**

Per the project's Jekyll Exclusion Rule, any new top-level directories or Markdown files must be added to `_config.yml`'s `exclude` list. The `docs/` directory should already be excluded — verify:

```bash
grep "docs" _config.yml
```

If `docs/` or `docs` is already listed, no change needed. If not, add it.

- [ ] **Step 3: Commit**

```bash
git add docs/youtube-setup.md _config.yml
git commit -m "docs: add YouTube upload setup guide"
```

---

### Task 3: Create `scripts/upload-youtube.py` — core helpers

**Files:**
- Create: `scripts/upload-youtube.py`
- Create: `tests/test_upload_youtube.py`

This task adds the pure-Python helper functions (no API calls): description extraction from report markdown, title builder, and token file path. These are testable without mocking the YouTube API.

- [ ] **Step 1: Write failing tests**

Create `tests/test_upload_youtube.py`:

```python
"""Tests for upload-youtube.py helper functions."""

import sys
from pathlib import Path

import pytest

# Load module (hyphen in filename)
import importlib.util

def load_module():
    spec = importlib.util.spec_from_file_location(
        "upload_youtube",
        Path(__file__).resolve().parent.parent / "scripts" / "upload-youtube.py",
    )
    mod = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mod)
    return mod


@pytest.fixture
def mod():
    return load_module()


class TestBuildTitle:
    def test_formats_date_correctly(self, mod):
        assert mod.build_title("2026-04-22") == "Toiletpaper Press — AI Daily 2026-04-22"

    def test_different_date(self, mod):
        assert mod.build_title("2026-01-01") == "Toiletpaper Press — AI Daily 2026-01-01"


class TestExtractDescription:
    def test_extracts_first_five_bullets(self, mod, tmp_path):
        report = tmp_path / "2026-04-22.md"
        report.write_text(
            "---\ntitle: Test\ndate: 2026-04-22\n---\n\n"
            "## Section One\n"
            "- **Item 1** — Source 1\n"
            "  Summary one.\n"
            "- **Item 2** — Source 2\n"
            "  Summary two.\n"
            "- **Item 3** — Source 3\n"
            "  Summary three.\n"
            "- **Item 4** — Source 4\n"
            "  Summary four.\n"
            "- **Item 5** — Source 5\n"
            "  Summary five.\n"
            "- **Item 6** — Source 6\n"
            "  Summary six.\n"
        )
        desc = mod.extract_description("2026-04-22", reports_dir=tmp_path)
        lines = desc.split("\n")
        # Should have 5 bullets + blank line + full report link
        bullet_lines = [l for l in lines if l.startswith("- ")]
        assert len(bullet_lines) == 5

    def test_appends_report_link(self, mod, tmp_path):
        report = tmp_path / "2026-04-22.md"
        report.write_text(
            "---\ntitle: Test\n---\n\n- **Item 1** — Source\n  Summary.\n"
        )
        desc = mod.extract_description("2026-04-22", reports_dir=tmp_path)
        assert "https://lankeami.github.io/ai-upskill/reports/2026-04-22" in desc

    def test_handles_fewer_than_five_bullets(self, mod, tmp_path):
        report = tmp_path / "2026-04-22.md"
        report.write_text(
            "---\ntitle: Test\n---\n\n- **Item 1** — Source\n  Summary.\n"
            "- **Item 2** — Source\n  Summary.\n"
        )
        desc = mod.extract_description("2026-04-22", reports_dir=tmp_path)
        bullet_lines = [l for l in desc.split("\n") if l.startswith("- ")]
        assert len(bullet_lines) == 2

    def test_strips_front_matter(self, mod, tmp_path):
        report = tmp_path / "2026-04-22.md"
        report.write_text(
            "---\ntitle: Test\npodcast_url: http://example.com\n---\n\n"
            "- **Item 1** — Source\n  Summary.\n"
        )
        desc = mod.extract_description("2026-04-22", reports_dir=tmp_path)
        assert "podcast_url" not in desc
        assert "title:" not in desc
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
source .venv/bin/activate && python -m pytest tests/test_upload_youtube.py -v
```

Expected: `ModuleNotFoundError` or `AttributeError` — file doesn't exist yet.

- [ ] **Step 3: Create `scripts/upload-youtube.py` with helper functions**

```python
#!/usr/bin/env python3
"""Upload a daily AI report video to the Toiletpaper Press YouTube channel."""

import argparse
import os
import re
import sys
from pathlib import Path

REPORTS_DIR = Path(__file__).resolve().parent.parent / "reports"
PODCASTS_DIR = Path(__file__).resolve().parent.parent / "podcasts"
TOKEN_FILE = Path(__file__).resolve().parent.parent / ".youtube-token.json"

CHANNEL_TAGS = ["AI news", "daily AI report", "Toiletpaper Press", "artificial intelligence"]
CATEGORY_ID = "28"  # Science & Technology
REPORT_BASE_URL = "https://lankeami.github.io/ai-upskill/reports"


def build_title(report_date: str) -> str:
    """Return the YouTube video title for a given date."""
    return f"Toiletpaper Press — AI Daily {report_date}"


def strip_front_matter(content: str) -> str:
    """Remove YAML front matter delimited by --- lines."""
    match = re.match(r"^---\s*\n.*?\n---\s*\n", content, re.DOTALL)
    if match:
        return content[match.end():]
    return content


def extract_description(report_date: str, reports_dir: Path = REPORTS_DIR) -> str:
    """Extract the first 3-5 bullet points from a report for the YouTube description."""
    report_path = reports_dir / f"{report_date}.md"
    if not report_path.exists():
        print(f"Error: Report not found at {report_path}", file=sys.stderr)
        sys.exit(1)

    raw = report_path.read_text(encoding="utf-8")
    content = strip_front_matter(raw)

    # Extract top-level bullet lines (lines starting with "- ")
    bullets = [line for line in content.splitlines() if line.startswith("- ")]
    top_bullets = bullets[:5]

    description = "\n".join(top_bullets)
    description += f"\n\nFull report: {REPORT_BASE_URL}/{report_date}"
    return description


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Upload daily AI report video to YouTube")
    parser.add_argument(
        "--date",
        type=str,
        help="Report date in YYYY-MM-DD format",
    )
    parser.add_argument(
        "--auth",
        action="store_true",
        help="Run one-time OAuth2 browser auth flow and save token",
    )
    return parser.parse_args()


def main() -> None:
    args = parse_args()

    if not args.auth and not args.date:
        print("Error: provide --date DATE or --auth", file=sys.stderr)
        sys.exit(1)

    if args.auth:
        run_auth_flow()
        return

    upload_video(args.date)


if __name__ == "__main__":
    main()
```

Note: `run_auth_flow()` and `upload_video()` are stubs — they will be implemented in Tasks 4 and 5.

- [ ] **Step 4: Run tests to verify they pass**

```bash
source .venv/bin/activate && python -m pytest tests/test_upload_youtube.py -v
```

Expected: all 6 tests pass.

- [ ] **Step 5: Commit**

```bash
git add scripts/upload-youtube.py tests/test_upload_youtube.py
git commit -m "feat: add upload-youtube.py with title builder and description extractor"
```

---

### Task 4: Implement OAuth2 auth flow

**Files:**
- Modify: `scripts/upload-youtube.py`
- Modify: `tests/test_upload_youtube.py`

- [ ] **Step 1: Write failing tests for auth helpers**

Add to `tests/test_upload_youtube.py`:

```python
class TestLoadCredentials:
    def test_loads_token_from_file(self, mod, tmp_path, monkeypatch):
        """load_credentials should return Credentials when token file exists."""
        from unittest.mock import MagicMock, patch

        token_data = {
            "token": "tok",
            "refresh_token": "ref",
            "token_uri": "https://oauth2.googleapis.com/token",
            "client_id": "cid",
            "client_secret": "csec",
            "scopes": ["https://www.googleapis.com/auth/youtube.upload"],
        }
        token_file = tmp_path / ".youtube-token.json"
        import json
        token_file.write_text(json.dumps(token_data))

        monkeypatch.setattr(mod, "TOKEN_FILE", token_file)
        monkeypatch.setenv("YOUTUBE_CLIENT_ID", "cid")
        monkeypatch.setenv("YOUTUBE_CLIENT_SECRET", "csec")

        with patch("google.oauth2.credentials.Credentials") as MockCreds:
            MockCreds.return_value.valid = True
            MockCreds.return_value.expired = False
            creds = mod.load_credentials()
            assert creds is not None

    def test_exits_when_token_missing(self, mod, tmp_path, monkeypatch):
        """load_credentials should exit 1 when token file is missing."""
        monkeypatch.setattr(mod, "TOKEN_FILE", tmp_path / ".youtube-token.json")
        monkeypatch.setenv("YOUTUBE_CLIENT_ID", "cid")
        monkeypatch.setenv("YOUTUBE_CLIENT_SECRET", "csec")

        with pytest.raises(SystemExit) as exc_info:
            mod.load_credentials()
        assert exc_info.value.code == 1

    def test_exits_when_client_id_missing(self, mod, tmp_path, monkeypatch):
        """load_credentials should exit 1 when YOUTUBE_CLIENT_ID is not set."""
        monkeypatch.delenv("YOUTUBE_CLIENT_ID", raising=False)
        monkeypatch.delenv("YOUTUBE_CLIENT_SECRET", raising=False)
        monkeypatch.setattr(mod, "TOKEN_FILE", tmp_path / ".youtube-token.json")

        with pytest.raises(SystemExit) as exc_info:
            mod.load_credentials()
        assert exc_info.value.code == 1
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
source .venv/bin/activate && python -m pytest tests/test_upload_youtube.py::TestLoadCredentials -v
```

Expected: `AttributeError: module has no attribute 'load_credentials'`

- [ ] **Step 3: Install dependencies**

```bash
source .venv/bin/activate && pip install google-api-python-client google-auth-oauthlib
```

- [ ] **Step 4: Implement `run_auth_flow` and `load_credentials` in `scripts/upload-youtube.py`**

Add the following imports at the top of the file (after existing imports):

```python
import json
from google.oauth2.credentials import Credentials
from google.auth.transport.requests import Request
from google_auth_oauthlib.flow import InstalledAppFlow
from googleapiclient.discovery import build
```

Add the following constants (after existing constants):

```python
SCOPES = ["https://www.googleapis.com/auth/youtube.upload",
          "https://www.googleapis.com/auth/youtube.readonly"]
```

Add these two functions (before `parse_args`):

```python
def load_credentials() -> Credentials:
    """Load OAuth2 credentials from token file, refreshing if needed.

    Exits with code 1 if env vars are missing or token file doesn't exist.
    """
    client_id = os.environ.get("YOUTUBE_CLIENT_ID")
    client_secret = os.environ.get("YOUTUBE_CLIENT_SECRET")

    if not client_id or not client_secret:
        print(
            "Error: YOUTUBE_CLIENT_ID and YOUTUBE_CLIENT_SECRET must be set in .env.\n"
            "See docs/youtube-setup.md for instructions.",
            file=sys.stderr,
        )
        sys.exit(1)

    if not TOKEN_FILE.exists():
        print(
            "Error: YouTube token file not found. Run:\n"
            "  python scripts/upload-youtube.py --auth",
            file=sys.stderr,
        )
        sys.exit(1)

    creds = Credentials.from_authorized_user_file(str(TOKEN_FILE), SCOPES)

    if creds.expired and creds.refresh_token:
        creds.refresh(Request())
        TOKEN_FILE.write_text(creds.to_json())

    return creds


def run_auth_flow() -> None:
    """Run one-time browser OAuth2 flow and save token to TOKEN_FILE."""
    client_id = os.environ.get("YOUTUBE_CLIENT_ID")
    client_secret = os.environ.get("YOUTUBE_CLIENT_SECRET")

    if not client_id or not client_secret:
        print(
            "Error: YOUTUBE_CLIENT_ID and YOUTUBE_CLIENT_SECRET must be set in .env.\n"
            "See docs/youtube-setup.md for instructions.",
            file=sys.stderr,
        )
        sys.exit(1)

    client_config = {
        "installed": {
            "client_id": client_id,
            "client_secret": client_secret,
            "auth_uri": "https://accounts.google.com/o/oauth2/auth",
            "token_uri": "https://oauth2.googleapis.com/token",
            "redirect_uris": ["urn:ietf:wg:oauth:2.0:oob", "http://localhost"],
        }
    }

    flow = InstalledAppFlow.from_client_config(client_config, SCOPES)
    creds = flow.run_local_server(port=0)
    TOKEN_FILE.write_text(creds.to_json())
    print(f"Auth complete. Token saved to {TOKEN_FILE}")
```

- [ ] **Step 5: Run tests to verify they pass**

```bash
source .venv/bin/activate && python -m pytest tests/test_upload_youtube.py::TestLoadCredentials -v
```

Expected: all 3 pass.

- [ ] **Step 6: Commit**

```bash
git add scripts/upload-youtube.py tests/test_upload_youtube.py
git commit -m "feat: implement OAuth2 auth flow for YouTube upload"
```

---

### Task 5: Implement idempotency check and upload

**Files:**
- Modify: `scripts/upload-youtube.py`
- Modify: `tests/test_upload_youtube.py`

- [ ] **Step 1: Write failing tests for idempotency and upload**

Add to `tests/test_upload_youtube.py`:

```python
class TestFindExistingVideo:
    def test_returns_url_when_title_matches(self, mod):
        """find_existing_video should return URL if a video with matching title exists."""
        from unittest.mock import MagicMock

        mock_youtube = MagicMock()
        mock_youtube.search().list().execute.return_value = {
            "items": [
                {
                    "id": {"videoId": "abc123"},
                    "snippet": {"title": "Toiletpaper Press — AI Daily 2026-04-22"},
                }
            ]
        }
        # Reset call chain for the actual call
        mock_youtube.search.return_value.list.return_value.execute.return_value = {
            "items": [
                {
                    "id": {"videoId": "abc123"},
                    "snippet": {"title": "Toiletpaper Press — AI Daily 2026-04-22"},
                }
            ]
        }

        result = mod.find_existing_video(mock_youtube, "Toiletpaper Press — AI Daily 2026-04-22")
        assert result == "https://youtube.com/watch?v=abc123"

    def test_returns_none_when_no_match(self, mod):
        """find_existing_video should return None when no matching video exists."""
        from unittest.mock import MagicMock

        mock_youtube = MagicMock()
        mock_youtube.search.return_value.list.return_value.execute.return_value = {
            "items": [
                {
                    "id": {"videoId": "xyz999"},
                    "snippet": {"title": "Some other video"},
                }
            ]
        }

        result = mod.find_existing_video(mock_youtube, "Toiletpaper Press — AI Daily 2026-04-22")
        assert result is None

    def test_returns_none_when_empty_results(self, mod):
        """find_existing_video should return None when search returns no items."""
        from unittest.mock import MagicMock

        mock_youtube = MagicMock()
        mock_youtube.search.return_value.list.return_value.execute.return_value = {"items": []}

        result = mod.find_existing_video(mock_youtube, "Toiletpaper Press — AI Daily 2026-04-22")
        assert result is None
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
source .venv/bin/activate && python -m pytest tests/test_upload_youtube.py::TestFindExistingVideo -v
```

Expected: `AttributeError: module has no attribute 'find_existing_video'`

- [ ] **Step 3: Implement `find_existing_video` and `upload_video` in `scripts/upload-youtube.py`**

Add `find_existing_video` (before `parse_args`):

```python
def find_existing_video(youtube, title: str) -> str | None:
    """Search the channel for a video with the exact title. Returns URL or None."""
    response = youtube.search().list(
        part="snippet",
        forMine=True,
        type="video",
        q=title,
        maxResults=10,
    ).execute()

    for item in response.get("items", []):
        if item["snippet"]["title"] == title:
            video_id = item["id"]["videoId"]
            return f"https://youtube.com/watch?v={video_id}"

    return None
```

Add `upload_video` (before `parse_args`):

```python
def upload_video(report_date: str) -> None:
    """Upload the video for the given date to YouTube and print the URL."""
    from googleapiclient.http import MediaFileUpload

    video_path = PODCASTS_DIR / f"{report_date}.mp4"
    if not video_path.exists():
        print(f"Error: Video not found at {video_path}", file=sys.stderr)
        sys.exit(1)

    title = build_title(report_date)
    description = extract_description(report_date)

    creds = load_credentials()
    youtube = build("youtube", "v3", credentials=creds)

    existing = find_existing_video(youtube, title)
    if existing:
        print(f"Video already uploaded: {existing}")
        print(existing)
        return

    body = {
        "snippet": {
            "title": title,
            "description": description,
            "tags": CHANNEL_TAGS,
            "categoryId": CATEGORY_ID,
        },
        "status": {
            "privacyStatus": "public",
        },
    }

    media = MediaFileUpload(str(video_path), mimetype="video/mp4", resumable=True)
    request = youtube.videos().insert(part="snippet,status", body=body, media_body=media)

    print(f"Uploading {video_path} to YouTube...")
    response = None
    while response is None:
        status, response = request.next_chunk()
        if status:
            pct = int(status.progress() * 100)
            print(f"  Upload progress: {pct}%")

    video_id = response["id"]
    url = f"https://youtube.com/watch?v={video_id}"
    print(f"Upload complete: {url}")
    print(url)
```

- [ ] **Step 4: Run all tests**

```bash
source .venv/bin/activate && python -m pytest tests/test_upload_youtube.py -v
```

Expected: all tests pass.

- [ ] **Step 5: Commit**

```bash
git add scripts/upload-youtube.py tests/test_upload_youtube.py
git commit -m "feat: implement idempotency check and video upload for YouTube"
```

---

### Task 6: Update `_layouts/report.html` — YouTube iframe embed

**Files:**
- Modify: `_layouts/report.html:57-74`

- [ ] **Step 1: Replace the `podcast_url` player block with `youtube_url` iframe**

In `_layouts/report.html`, replace lines 57–74:

```html
{% if page.podcast_url %}
<div class="podcast-player" style="margin-bottom: 1.5rem; padding: 1rem; background: #f6f8fa; border-radius: 6px;">
  {% if page.podcast_url contains '.mp4' %}
  <h3 style="font-size: 0.9rem; margin-bottom: 0.5rem; color: #656d76; text-transform: uppercase; letter-spacing: 0.05em;">Watch this report</h3>
  <video controls preload="none" style="width: 100%; max-width: 100%; border-radius: 4px; margin-bottom: 0.5rem;">
    <source src="{{ page.podcast_url }}" type="video/mp4">
    Your browser does not support the video element.
  </video>
  {% else %}
  <h3 style="font-size: 0.9rem; margin-bottom: 0.5rem; color: #656d76; text-transform: uppercase; letter-spacing: 0.05em;">Listen to this report</h3>
  <audio controls preload="none" style="width: 100%; margin-bottom: 0.5rem;">
    <source src="{{ page.podcast_url }}" type="audio/mpeg">
    Your browser does not support the audio element.
  </audio>
  {% endif %}
  <p style="margin: 0; font-size: 0.85rem;"><a href="{{ page.podcast_url }}">Download</a></p>
</div>
{% endif %}
```

With:

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

- [ ] **Step 2: Verify Jekyll still builds**

```bash
bundle exec jekyll build 2>&1 | tail -5
```

Expected: `done in X seconds` with no errors.

- [ ] **Step 3: Commit**

```bash
git add _layouts/report.html
git commit -m "feat: replace podcast video player with YouTube iframe embed"
```

---

### Task 7: Update `.claude/skills/daily-pipeline.md`

**Files:**
- Modify: `.claude/skills/daily-pipeline.md`

Three changes: (a) add YouTube prerequisites, (b) replace Step 4 (GitHub Release) with YouTube upload, (c) update the Done summary.

- [ ] **Step 1: Add YouTube prerequisites**

In `.claude/skills/daily-pipeline.md`, replace the Prerequisites Check section (lines 20–32) with:

```markdown
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
```

- [ ] **Step 2: Replace Step 4 (GitHub Release) with YouTube upload**

Replace the Step 4 section (lines 97–108) with:

```markdown
## Step 4: Upload to YouTube

**Check:** Does `reports/DATE.md` already contain a `youtube_url` in the YAML front matter? If yes, skip this step and note "YouTube upload already done for DATE".

If not, run:
```bash
python scripts/upload-youtube.py --date DATE
```

Capture the last line of stdout — this is the YouTube URL (format: `https://youtube.com/watch?v=VIDEO_ID`). Store it as `YOUTUBE_URL`.

If the command fails, stop the pipeline and show the full error output.
```

- [ ] **Step 3: Update Step 5 to use `youtube_url`**

Replace the Step 5 section (lines 110–118) with:

```markdown
## Step 5: Update Report Front Matter

**Check:** Read `reports/DATE.md` and check if `youtube_url` is already present in the YAML front matter. If it already contains the correct URL, skip this step.

If not, use the Edit tool to add the following line to the YAML front matter (before the closing `---`):

```
youtube_url: "YOUTUBE_URL"
```

Where `YOUTUBE_URL` is the URL captured in Step 4.
```

- [ ] **Step 4: Update the Done summary**

Replace the Done summary section (lines 139–146) with:

```markdown
## Done

Summarize what was done:
- Report: generated or already existed
- Video: generated or already existed
- YouTube: uploaded or already existed
- Front matter: updated or already correct
- Commit: pushed or nothing to commit
```

- [ ] **Step 5: Verify the skill file is well-formed**

Read back `.claude/skills/daily-pipeline.md` and confirm:
- No references to `gh release` remain in the active pipeline steps
- No references to `podcast_url` remain in the active pipeline steps
- Prerequisites now include YouTube checks
- Steps 1, 2, 3, 6 are unchanged

- [ ] **Step 6: Commit**

```bash
git add .claude/skills/daily-pipeline.md
git commit -m "feat: update daily-pipeline skill to upload to YouTube instead of GitHub Releases"
```

---

### Task 8: Final verification

**Files:** None (verification only)

- [ ] **Step 1: Run all tests**

```bash
source .venv/bin/activate && python -m pytest tests/test_upload_youtube.py tests/test_generate_podcast.py -v
```

Expected: all tests pass.

- [ ] **Step 2: Verify CLI help**

```bash
source .venv/bin/activate && python scripts/upload-youtube.py --help
```

Expected output includes `--date` and `--auth` options.

- [ ] **Step 3: Verify `.youtube-token.json` is gitignored**

```bash
echo "test" > .youtube-token.json && git status --porcelain .youtube-token.json && rm .youtube-token.json
```

Expected: no output (file is ignored).

- [ ] **Step 4: Verify Jekyll builds cleanly**

```bash
bundle exec jekyll build 2>&1 | tail -5
```

Expected: build completes without errors.
