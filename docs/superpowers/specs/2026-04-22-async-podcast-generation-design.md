# Async Podcast Generation Design

**Date:** 2026-04-22
**Status:** Draft

## Problem

The `daily-pipeline` skill calls `scripts/generate-podcast.py`, which blocks for up to 30 minutes waiting for NotebookLM to generate video/audio. Claude Code's Bash tool has a maximum timeout of 600 seconds (10 minutes), so the command times out before generation completes. Claude cannot proceed to download the artifact or continue the pipeline.

## Solution

Refactor `scripts/generate-podcast.py` to support three subcommands — `start`, `poll`, `download` — that break the long-running operation into short, discrete steps. A JSON state file persists context between calls. The `daily-pipeline` skill orchestrates the polling loop itself.

## Script Changes

### Subcommands

The script adds three subcommands via `argparse` subparsers. The original monolithic behavior (no subcommand) is preserved for backward compatibility with the GitHub Actions workflow.

#### `start`

```bash
python scripts/generate-podcast.py start --date 2026-04-22 --media-type video
```

1. Reads the report from `reports/DATE.md`
2. Creates a NotebookLM notebook, adds the report as a text source
3. Kicks off video/audio generation
4. Writes state to `.podcast-state.json` in the project root
5. Prints `Started video generation for 2026-04-22 (task: <id>)` and exits 0

**Idempotency:** If `.podcast-state.json` already exists:
- Same date and media type: skip creation, print "Generation already in progress for DATE", exit 0
- Different date: exit 1 with error "State file exists for a different date (DATE_OLD). Run `download` or delete `.podcast-state.json` first."

#### `poll`

```bash
python scripts/generate-podcast.py poll
```

Reads `.podcast-state.json`, calls `poll_status()` once (non-blocking), prints status.

| Condition | Exit code | Output |
|-----------|-----------|--------|
| Generation complete | 0 | `complete` |
| Generation pending/in-progress | 1 | `pending` or `in_progress` |
| Generation failed | 2 | Error message from `GenerationStatus.error` |
| No state file | 1 | `No generation in progress. Run 'start' first.` |

#### `download`

```bash
python scripts/generate-podcast.py download
```

1. Reads `.podcast-state.json`
2. Verifies generation is complete via `poll_status()`. If not complete, exits with error.
3. Downloads artifact to `podcasts/DATE.ext`
4. Deletes the NotebookLM notebook (logs warning on failure, does not fail the download)
5. Removes `.podcast-state.json`
6. Prints path to downloaded file, exits 0

### State File

**Location:** `.podcast-state.json` in project root (added to `.gitignore`).

```json
{
  "notebook_id": "abc123",
  "task_id": "def456",
  "date": "2026-04-22",
  "media_type": "video",
  "started_at": "2026-04-22T08:00:00Z"
}
```

### Backward Compatibility

The original invocation without a subcommand continues to work:

```bash
python scripts/generate-podcast.py --date DATE --media-type video
```

This runs the existing monolithic flow (create, wait, download, cleanup) unchanged. The GitHub Actions workflow uses this form and requires no changes.

## Skill Changes

Step 3 of the `daily-pipeline` skill changes from:

```
python scripts/generate-podcast.py --date DATE --media-type video
# (single blocking command with 2400s timeout)
```

To:

```
Step 3a: Start Generation
  python scripts/generate-podcast.py start --date DATE --media-type video

Step 3b: Poll Loop
  Repeat every 30 seconds:
    python scripts/generate-podcast.py poll
  Until exit code 0 (complete) or 2 (failed).
  Max 40 iterations (~20 minutes). If exceeded, stop pipeline with timeout error.

Step 3c: Download
  python scripts/generate-podcast.py download
```

All other steps in the skill (1, 2, 4, 5, 6) are unchanged.

## Error Handling

| Scenario | Behavior |
|----------|----------|
| `start` with existing state for same date | Skip, print "already in progress", exit 0 |
| `start` with existing state for different date | Error, exit 1 |
| `poll`/`download` with no state file | Error, exit 1 |
| `poll` returns `is_failed` | Exit 2, print error. Skill stops pipeline. |
| Poll loop exceeds 40 iterations | Skill stops pipeline with timeout error. State file left for manual recovery. |
| `download` when generation not complete | Error, exit 1 |
| Notebook cleanup fails during download | Log warning, continue. Artifact download is what matters. |

## Files Changed

| File | Change |
|------|--------|
| `scripts/generate-podcast.py` | Add `start`, `poll`, `download` subcommands; preserve monolithic mode |
| `.claude/skills/daily-pipeline.md` | Replace Step 3 with 3a/3b/3c polling flow |
| `.gitignore` | Add `.podcast-state.json` |
