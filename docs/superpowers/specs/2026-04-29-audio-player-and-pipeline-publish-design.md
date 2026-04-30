# Design: Audio Player + Pipeline Publish Step

**Date:** 2026-04-29
**Status:** Approved

## Problem

1. `_layouts/report.html` only renders a YouTube iframe (`page.youtube_url`). No audio player exists for `page.podcast_url`.
2. The local `/daily-pipeline` skill generates and downloads the MP3 but never publishes it — no GitHub Release is created, `podcast_url` is never added to the report front matter, and the site shows nothing.

## Scope

Two files changed, one skill updated:

| File | Change |
|---|---|
| `_layouts/report.html` | Replace `youtube_url` block with `podcast_url` audio player |
| `.claude/skills/daily-pipeline.md` | Insert Step 4 (publish release + inject front matter) between download and commit |

## Design

### 1. Jekyll Layout — Audio Player

Replace the existing `{% if page.youtube_url %}` block with:

```html
{% if page.podcast_url %}
<div class="podcast-player" style="margin-bottom: 1.5rem; padding: 1rem; background: #f6f8fa; border-radius: 6px;">
  <h3 style="font-size: 0.9rem; margin-bottom: 0.5rem; color: #656d76; text-transform: uppercase; letter-spacing: 0.05em;">Listen to this report</h3>
  <audio controls style="width: 100%; margin-bottom: 0.5rem;">
    <source src="{{ page.podcast_url }}" type="audio/mpeg">
  </audio>
  <p style="margin: 0; font-size: 0.85rem;"><a href="{{ page.podcast_url }}">Download MP3</a></p>
</div>
{% endif %}
```

- `youtube_url` block removed entirely (dead code once audio-only)
- Native `<audio>` element — no third-party dependencies
- Same visual container style as existing block

### 2. Daily Pipeline Skill — Publish Step

Insert new **Step 4: Publish GitHub Release** after Step 3c (download) and before the commit step.

**Step 4 logic:**

1. Check if tag `podcast-DATE` already exists: `gh release view podcast-DATE 2>/dev/null` — skip if found, note "Release already exists for DATE"
2. Create release and upload MP3:
   ```bash
   gh release create podcast-DATE podcasts/DATE.mp3 \
     --title "Podcast — DATE" \
     --notes "Audio podcast for AI Daily Report DATE"
   ```
3. Capture asset URL (deterministic — no need to parse output):
   ```
   https://github.com/lankeami/ai-upskill/releases/download/podcast-DATE/DATE.mp3
   ```
4. Inject `podcast_url` into report front matter using Python (same pattern as `podcast.yml` CI workflow):
   ```python
   parts = content.split('---', 2)
   front_matter = parts[1].rstrip('\n') + '\npodcast_url: URL\n'
   content = '---' + front_matter + '---' + parts[2]
   ```
5. Step 5 (commit) adds `reports/DATE.md` and uses message: `chore: add podcast URL to DATE report`

**Done summary gains:**
- GitHub Release: published or already existed
- Front matter: updated or already correct

## Data Flow

```
podcasts/DATE.mp3 (local)
  → gh release create → GitHub Release asset URL
  → inject into reports/DATE.md front matter as podcast_url
  → git add reports/DATE.md && git commit && git push
  → Jekyll renders <audio> player on site
```

## Out of Scope

- No changes to `podcast.yml` CI workflow (already handles publish correctly)
- No changes to `generate-podcast.py` (publish is a shell+Python operation, not NotebookLM)
- No YouTube support retained
