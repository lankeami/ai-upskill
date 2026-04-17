#!/usr/bin/env python3
"""Generate an audio or video podcast from a daily AI report using NotebookLM."""

import argparse
import asyncio
import re
import sys
from datetime import date, datetime
from pathlib import Path

REPORTS_DIR = Path(__file__).resolve().parent.parent / "reports"
PODCASTS_DIR = Path(__file__).resolve().parent.parent / "podcasts"


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Generate podcast from AI daily report")
    parser.add_argument(
        "--date",
        type=str,
        default=date.today().isoformat(),
        help="Report date in YYYY-MM-DD format (default: today)",
    )
    parser.add_argument(
        "--media-type",
        type=str,
        choices=["audio", "video"],
        default="video",
        help="Media type to generate: audio or video (default: video)",
    )
    return parser.parse_args()


def strip_front_matter(content: str) -> str:
    """Remove YAML front matter delimited by --- lines."""
    match = re.match(r"^---\s*\n.*?\n---\s*\n", content, re.DOTALL)
    if match:
        return content[match.end():]
    return content


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
