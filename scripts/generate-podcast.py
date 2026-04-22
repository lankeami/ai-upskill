#!/usr/bin/env python3
"""Generate an audio or video podcast from a daily AI report using NotebookLM."""

import argparse
import asyncio
import json
import re
import sys
from datetime import date, datetime, timezone
from pathlib import Path

REPORTS_DIR = Path(__file__).resolve().parent.parent / "reports"
PODCASTS_DIR = Path(__file__).resolve().parent.parent / "podcasts"
STATE_FILE = Path(__file__).resolve().parent.parent / ".podcast-state.json"


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Generate podcast from AI daily report")

    # Legacy mode args (top-level, used when no subcommand is given)
    parser.add_argument(
        "--date", type=str, default=date.today().isoformat(),
        help="Report date in YYYY-MM-DD format (default: today)",
    )
    parser.add_argument(
        "--media-type", type=str, choices=["audio", "video"], default="video",
        help="Media type to generate: audio or video (default: video)",
    )

    subparsers = parser.add_subparsers(dest="command")

    # start subcommand
    start_parser = subparsers.add_parser("start", help="Start podcast generation")
    start_parser.add_argument(
        "--date", type=str, default=date.today().isoformat(),
        help="Report date in YYYY-MM-DD format (default: today)",
    )
    start_parser.add_argument(
        "--media-type", type=str, choices=["audio", "video"], default="video",
        help="Media type to generate (default: video)",
    )

    # poll subcommand — no arguments needed (reads state file)
    subparsers.add_parser("poll", help="Check generation status")

    # download subcommand — no arguments needed (reads state file)
    subparsers.add_parser("download", help="Download completed artifact")

    return parser.parse_args()


def strip_front_matter(content: str) -> str:
    """Remove YAML front matter delimited by --- lines."""
    match = re.match(r"^---\s*\n.*?\n---\s*\n", content, re.DOTALL)
    if match:
        return content[match.end():]
    return content


def write_state(notebook_id: str, task_id: str, report_date: str, media_type: str) -> None:
    """Write generation state to the state file."""
    state = {
        "notebook_id": notebook_id,
        "task_id": task_id,
        "date": report_date,
        "media_type": media_type,
        "started_at": datetime.now(timezone.utc).isoformat(),
    }
    STATE_FILE.write_text(json.dumps(state, indent=2))


def read_state() -> dict | None:
    """Read generation state from the state file. Returns None if no file."""
    if not STATE_FILE.exists():
        return None
    return json.loads(STATE_FILE.read_text())


def delete_state() -> None:
    """Remove the state file if it exists."""
    if STATE_FILE.exists():
        STATE_FILE.unlink()


async def generate_podcast(report_date: str, media_type: str) -> Path:
    """Generate a podcast from the report for the given date."""
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
    ext = ".mp4" if media_type == "video" else ".mp3"
    output_path = PODCASTS_DIR / f"{report_date}{ext}"

    print(f"Generating {media_type} podcast for {report_date}...")

    from notebooklm import NotebookLMClient, AudioFormat, AudioLength, VideoFormat, VideoStyle

    async with await NotebookLMClient.from_storage() as client:
        # Create notebook
        notebook_title = f"AI Daily Report — {report_date}"
        nb = await client.notebooks.create(notebook_title)
        print(f"Created notebook: {notebook_title} ({nb.id})")

        try:
            # Add report as text source
            await client.sources.add_text(nb.id, notebook_title, content)
            print("Added report content as source")

            if media_type == "video":
                # Generate video
                status = await client.artifacts.generate_video(
                    nb.id,
                    video_format=VideoFormat.EXPLAINER,
                    video_style=VideoStyle.AUTO_SELECT,
                )
                print(f"Video generation started (task: {status.task_id})")

                # Wait for completion (video takes longer — up to 30 min for large reports)
                await client.artifacts.wait_for_completion(nb.id, status.task_id, timeout=1800.0)
                print("Video generation complete")

                # Download MP4
                await client.artifacts.download_video(nb.id, str(output_path))
                print(f"Downloaded video to {output_path}")
            else:
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

    output = asyncio.run(generate_podcast(args.date, args.media_type))
    print(f"\n{args.media_type.title()} generated: {output}")


if __name__ == "__main__":
    main()
