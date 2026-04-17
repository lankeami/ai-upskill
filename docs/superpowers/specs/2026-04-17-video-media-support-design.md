# Configurable Video/Audio Media Generation

**Date:** 2026-04-17
**Status:** Approved
**Extends:** 2026-04-16-podcast-generation-design.md

## Overview

Switch the daily report media pipeline from audio-only to video by default, while keeping audio as a configurable fallback. Uses NotebookLM's Video Overview feature (`generate_video` / `download_video`) via `notebooklm-py`. The Jekyll report layout renders an HTML5 `<video>` or `<audio>` player based on the file extension of the media URL.

## Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Default media type | Video | More engaging than audio-only |
| Video format | Explainer | Best fit for daily news summaries; cinematic is too slow |
| Video style | Auto | Let NotebookLM choose the best visual style |
| Configuration method | Workflow input parameter | Visible in GitHub Actions UI, overridable per run |
| Hosting | GitHub Release assets (MP4) | Same pattern as existing MP3 hosting |
| Backward compatibility | Extension-based detection in layout | Existing `.mp3` reports keep working |

## Components

### 1. Python Script: `scripts/generate-podcast.py`

**New CLI flag:** `--media-type` (choices: `video`, `audio`; default: `video`)

**When `--media-type video`:**
- Imports `VideoFormat`, `VideoStyle` from notebooklm
- Calls `client.artifacts.generate_video()` with `VideoFormat.EXPLAINER`, `VideoStyle.AUTO_SELECT`
- Uses 600s timeout (video generation takes longer than audio)
- Downloads via `client.artifacts.download_video()` to `podcasts/YYYY-MM-DD.mp4`

**When `--media-type audio`:**
- Existing behavior unchanged
- `AudioFormat.DEEP_DIVE`, `AudioLength.SHORT`
- Downloads via `client.artifacts.download_audio()` to `podcasts/YYYY-MM-DD.mp3`

**Shared flow (unchanged):** Create notebook, add text source, generate, wait for completion, download, delete notebook.

### 2. GitHub Actions Workflow: `.github/workflows/podcast.yml`

**New `workflow_dispatch` input:**
```yaml
media_type:
  description: 'Media type to generate'
  required: false
  default: 'video'
  type: choice
  options:
    - video
    - audio
```

**Push trigger default:** `video` (set via environment variable since push triggers have no inputs).

**Changes:**
- Pass `--media-type` flag to the script
- Release asset file extension: `.mp4` for video, `.mp3` for audio
- Release tag stays `podcast-YYYY-MM-DD`
- Front matter field stays `podcast_url` for backward compatibility

### 3. Jekyll Layout: `_layouts/report.html`

**Extension-based media detection using Liquid:**

```liquid
{% if page.podcast_url contains '.mp4' %}
  <h3>Watch this report</h3>
  <video controls preload="none" style="width: 100%;">
    <source src="{{ page.podcast_url }}" type="video/mp4">
    Your browser does not support the video element.
  </video>
{% else %}
  <h3>Listen to this report</h3>
  <audio controls preload="none" style="width: 100%;">
    <source src="{{ page.podcast_url }}" type="audio/mpeg">
    Your browser does not support the audio element.
  </audio>
{% endif %}
```

Existing reports with `.mp3` URLs continue to render the audio player. New reports with `.mp4` URLs get the video player.

## Files Changed

| File | Change |
|------|--------|
| `scripts/generate-podcast.py` | Add `--media-type` flag, video generation path |
| `.github/workflows/podcast.yml` | Add `media_type` input, pass to script, handle MP4 assets |
| `_layouts/report.html` | Extension-based video/audio player rendering |
