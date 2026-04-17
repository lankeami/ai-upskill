# Video Media Support Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Switch the daily report media pipeline to generate video by default (configurable to audio) using NotebookLM's Video Overview, with an HTML5 video player in the Jekyll layout.

**Architecture:** Add a `--media-type` flag to the existing `generate-podcast.py` script that branches between audio and video generation/download calls. The GitHub Actions workflow exposes this as a `media_type` input (default: `video`). The Jekyll layout detects `.mp4` vs `.mp3` file extensions to render the appropriate HTML5 player.

**Tech Stack:** Python 3.12, notebooklm-py (VideoFormat, VideoStyle), GitHub Actions, Jekyll/Liquid

---

### Task 1: Add `--media-type` flag to `scripts/generate-podcast.py`

**Files:**
- Modify: `scripts/generate-podcast.py`

- [ ] **Step 1: Add VideoFormat and VideoStyle imports and media_type argument**

In `scripts/generate-podcast.py`, update the imports and `parse_args()`:

```python
# At top of file, no changes to existing imports needed yet.
# VideoFormat/VideoStyle are imported lazily inside generate_podcast()
# alongside the existing lazy notebooklm imports.

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
        help="Media type to generate (default: video)",
    )
    return parser.parse_args()
```

- [ ] **Step 2: Update `generate_podcast()` to accept and use media_type**

Replace the entire `generate_podcast` function with this version that branches on media type:

```python
async def generate_podcast(report_date: str, media_type: str) -> Path:
    """Generate a podcast MP3 or video MP4 from the report for the given date."""
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

    print(f"Generating {media_type} for {report_date}...")

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

                # Wait for completion (longer timeout for video)
                await client.artifacts.wait_for_completion(nb.id, status.task_id, timeout=600.0)
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
```

- [ ] **Step 3: Update `main()` to pass media_type**

```python
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
```

- [ ] **Step 4: Verify script parses arguments correctly**

Run: `python scripts/generate-podcast.py --help`

Expected output should show both `--date` and `--media-type` flags, with `--media-type` defaulting to `video` and accepting `audio` or `video`.

- [ ] **Step 5: Commit**

```bash
git add scripts/generate-podcast.py
git commit -m "feat: add --media-type flag to generate-podcast script"
```

---

### Task 2: Update GitHub Actions workflow with `media_type` input

**Files:**
- Modify: `.github/workflows/podcast.yml`

- [ ] **Step 1: Add media_type input to workflow_dispatch and env default**

Replace the entire `.github/workflows/podcast.yml` with:

```yaml
name: Generate Podcast

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

env:
  DEFAULT_MEDIA_TYPE: video

permissions:
  contents: write

jobs:
  generate-podcast:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 2

      - name: Set up Python
        uses: actions/setup-python@v5
        with:
          python-version: '3.12'

      - name: Install notebooklm-py
        run: pip install notebooklm-py

      - name: Determine report date
        id: date
        run: |
          if [ -n "${{ github.event.inputs.date }}" ]; then
            echo "report_date=${{ github.event.inputs.date }}" >> "$GITHUB_OUTPUT"
          else
            # Find the most recently changed report file
            REPORT_FILE=$(git diff --name-only HEAD~1 HEAD -- 'reports/*.md' | head -1)
            if [ -z "$REPORT_FILE" ]; then
              echo "No report file changed in last commit, skipping"
              echo "skip=true" >> "$GITHUB_OUTPUT"
              exit 0
            fi
            # Extract date from filename: reports/YYYY-MM-DD.md -> YYYY-MM-DD
            REPORT_DATE=$(basename "$REPORT_FILE" .md)
            echo "report_date=$REPORT_DATE" >> "$GITHUB_OUTPUT"
          fi

      - name: Determine media type
        id: media
        run: |
          MEDIA_TYPE="${{ github.event.inputs.media_type || env.DEFAULT_MEDIA_TYPE }}"
          echo "media_type=$MEDIA_TYPE" >> "$GITHUB_OUTPUT"
          if [ "$MEDIA_TYPE" = "video" ]; then
            echo "ext=mp4" >> "$GITHUB_OUTPUT"
            echo "mime=video/mp4" >> "$GITHUB_OUTPUT"
          else
            echo "ext=mp3" >> "$GITHUB_OUTPUT"
            echo "mime=audio/mpeg" >> "$GITHUB_OUTPUT"
          fi

      - name: Generate podcast
        if: steps.date.outputs.skip != 'true'
        env:
          NOTEBOOKLM_AUTH_JSON: ${{ secrets.NOTEBOOKLM_AUTH_JSON }}
        run: python scripts/generate-podcast.py --date ${{ steps.date.outputs.report_date }} --media-type ${{ steps.media.outputs.media_type }}

      - name: Create GitHub Release
        if: steps.date.outputs.skip != 'true'
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: |
          REPORT_DATE="${{ steps.date.outputs.report_date }}"
          EXT="${{ steps.media.outputs.ext }}"
          TAG="podcast-${REPORT_DATE}"
          MEDIA_FILE="podcasts/${REPORT_DATE}.${EXT}"

          gh release create "$TAG" "$MEDIA_FILE" \
            --title "Podcast — ${REPORT_DATE}" \
            --notes "${{ steps.media.outputs.media_type }} for AI Daily Report ${REPORT_DATE}"

      - name: Update report front matter with podcast URL
        if: steps.date.outputs.skip != 'true'
        run: |
          REPORT_DATE="${{ steps.date.outputs.report_date }}"
          REPORT_FILE="reports/${REPORT_DATE}.md"
          EXT="${{ steps.media.outputs.ext }}"
          PODCAST_URL="https://github.com/lankeami/ai-upskill/releases/download/podcast-${REPORT_DATE}/${REPORT_DATE}.${EXT}"

          # Use Python for reliable front matter editing
          python3 -c "
          import re, sys

          path = '${REPORT_FILE}'
          with open(path, 'r') as f:
              content = f.read()

          # Match closing --- of front matter (second occurrence)
          parts = content.split('---', 2)
          if len(parts) >= 3:
              front_matter = parts[1]
              rest = parts[2]
              front_matter = front_matter.rstrip('\n') + '\npodcast_url: ${PODCAST_URL}\n'
              content = '---' + front_matter + '---' + rest
              with open(path, 'w') as f:
                  f.write(content)
              print(f'Added podcast_url to {path}')
          else:
              print(f'Error: Could not parse front matter in {path}', file=sys.stderr)
              sys.exit(1)
          "

      - name: Commit updated report
        if: steps.date.outputs.skip != 'true'
        run: |
          git config user.name "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"
          git add reports/
          if git diff --cached --quiet; then
            echo "No changes to commit"
          else
            git commit -m "chore: add podcast URL to ${{ steps.date.outputs.report_date }} report"
            git push
          fi
```

- [ ] **Step 2: Verify YAML syntax**

Run: `python3 -c "import yaml; yaml.safe_load(open('.github/workflows/podcast.yml'))"`

Expected: No errors (exits silently).

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/podcast.yml
git commit -m "feat: add media_type workflow input, default to video"
```

---

### Task 3: Update Jekyll layout for video/audio player

**Files:**
- Modify: `_layouts/report.html`

- [ ] **Step 1: Replace the podcast player block**

In `_layouts/report.html`, replace the existing `{% if page.podcast_url %}` block (lines 57-66) with:

```liquid
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

- [ ] **Step 2: Verify layout renders correctly with existing reports**

Run: `bundle exec jekyll build 2>&1 | tail -5`

Expected: Build succeeds. Existing reports with `.mp3` URLs still work (the `contains '.mp4'` check is false, so they fall through to the audio player).

- [ ] **Step 3: Commit**

```bash
git add _layouts/report.html
git commit -m "feat: add HTML5 video player with audio fallback in report layout"
```
