#!/usr/bin/env python3
"""Upload a daily AI report video to the Toiletpaper Press YouTube channel."""

import argparse
import json
import os
import re
import sys
from pathlib import Path

from google.auth.transport.requests import Request
from google.oauth2.credentials import Credentials
from google_auth_oauthlib.flow import InstalledAppFlow
from googleapiclient.discovery import build

REPORTS_DIR = Path(__file__).resolve().parent.parent / "reports"
PODCASTS_DIR = Path(__file__).resolve().parent.parent / "podcasts"
TOKEN_FILE = Path(__file__).resolve().parent.parent / ".youtube-token.json"

SCOPES = [
    "https://www.googleapis.com/auth/youtube.upload",
    "https://www.googleapis.com/auth/youtube.readonly",
]
CHANNEL_TAGS = ["AI news", "daily AI report", "Toiletpaper Press", "artificial intelligence"]
CATEGORY_ID = "28"  # Science & Technology
REPORT_BASE_URL = "https://lankeami.github.io/ai-upskill/reports"


def build_title(report_date: str) -> str:
    """Return the YouTube video title for a given date."""
    return f"Toiletpaper Press \u2014 AI Daily {report_date}"


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
