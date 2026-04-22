# Async Podcast Generation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Split the blocking `generate-podcast.py` script into start/poll/download subcommands so the daily-pipeline skill can manage long-running NotebookLM generations without hitting Claude Code's 10-minute Bash timeout.

**Architecture:** Add argparse subparsers to the existing script for `start`, `poll`, and `download` commands. A `.podcast-state.json` file in the project root persists notebook/task IDs between calls. The original no-subcommand invocation is preserved for backward compatibility with the GitHub Actions workflow.

**Tech Stack:** Python 3, `notebooklm-py` library, argparse subparsers, JSON state file

**Spec:** `docs/superpowers/specs/2026-04-22-async-podcast-generation-design.md`

---

## File Structure

| File | Action | Responsibility |
|------|--------|---------------|
| `scripts/generate-podcast.py` | Modify | Add `start`, `poll`, `download` subcommands; keep monolithic mode |
| `.claude/skills/daily-pipeline.md` | Modify | Replace Step 3 with 3a/3b/3c polling flow |
| `.gitignore` | Modify | Add `.podcast-state.json` |
| `tests/test_generate_podcast.py` | Create | Unit tests for state file logic, arg parsing, subcommand routing |

---

### Task 1: Add `.podcast-state.json` to `.gitignore`

**Files:**
- Modify: `.gitignore:1-9`

- [ ] **Step 1: Add the state file to `.gitignore`**

Add `.podcast-state.json` to `.gitignore`:

```
.podcast-state.json
```

Append it after the existing entries.

- [ ] **Step 2: Commit**

```bash
git add .gitignore
git commit -m "chore: add .podcast-state.json to .gitignore"
```

---

### Task 2: Add state file helpers and subcommand arg parsing

**Files:**
- Modify: `scripts/generate-podcast.py:1-30`
- Create: `tests/test_generate_podcast.py`

This task adds the state file read/write helpers, the `STATE_FILE` constant, and refactors `parse_args()` to support subcommands while preserving backward compatibility.

- [ ] **Step 1: Write failing tests for state file helpers**

Create `tests/test_generate_podcast.py`:

```python
"""Tests for generate-podcast.py state file and subcommand logic."""

import json
import sys
from datetime import datetime, timezone
from pathlib import Path
from unittest.mock import patch

import pytest

# Add scripts dir to path so we can import the module
sys.path.insert(0, str(Path(__file__).resolve().parent.parent / "scripts"))

# We need to import functions after they exist — these tests will fail until implemented
from importlib import import_module


@pytest.fixture
def podcast_module():
    """Import generate-podcast as a module (hyphen in filename)."""
    import importlib.util
    spec = importlib.util.spec_from_file_location(
        "generate_podcast",
        Path(__file__).resolve().parent.parent / "scripts" / "generate-podcast.py",
    )
    mod = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mod)
    return mod


@pytest.fixture
def state_file(tmp_path, podcast_module):
    """Override STATE_FILE to use a temp directory."""
    sf = tmp_path / ".podcast-state.json"
    podcast_module.STATE_FILE = sf
    return sf


class TestWriteState:
    def test_writes_valid_json(self, podcast_module, state_file):
        podcast_module.write_state("nb123", "task456", "2026-04-22", "video")
        data = json.loads(state_file.read_text())
        assert data["notebook_id"] == "nb123"
        assert data["task_id"] == "task456"
        assert data["date"] == "2026-04-22"
        assert data["media_type"] == "video"
        assert "started_at" in data

    def test_started_at_is_iso_format(self, podcast_module, state_file):
        podcast_module.write_state("nb1", "t1", "2026-04-22", "audio")
        data = json.loads(state_file.read_text())
        # Should not raise
        datetime.fromisoformat(data["started_at"])


class TestReadState:
    def test_reads_existing_state(self, podcast_module, state_file):
        state_file.write_text(json.dumps({
            "notebook_id": "nb1",
            "task_id": "t1",
            "date": "2026-04-22",
            "media_type": "video",
            "started_at": "2026-04-22T08:00:00+00:00",
        }))
        data = podcast_module.read_state()
        assert data["notebook_id"] == "nb1"

    def test_returns_none_when_no_file(self, podcast_module, state_file):
        assert podcast_module.read_state() is None


class TestDeleteState:
    def test_removes_file(self, podcast_module, state_file):
        state_file.write_text("{}")
        podcast_module.delete_state()
        assert not state_file.exists()

    def test_no_error_when_missing(self, podcast_module, state_file):
        podcast_module.delete_state()  # Should not raise
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
source .venv/bin/activate && python -m pytest tests/test_generate_podcast.py -v
```

Expected: failures because `write_state`, `read_state`, `delete_state`, and `STATE_FILE` don't exist yet.

- [ ] **Step 3: Implement state file helpers and new arg parsing**

In `scripts/generate-podcast.py`, add the following imports and constants at the top (after line 9):

```python
import json
```

Add `STATE_FILE` constant after `PODCASTS_DIR` (after line 12):

```python
STATE_FILE = Path(__file__).resolve().parent.parent / ".podcast-state.json"
```

Add three helper functions after `strip_front_matter` (after line 38):

```python
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
```

Also add `from datetime import date, datetime, timezone` (update the existing import on line 8).

Replace `parse_args()` (lines 15-30) with a version that supports subcommands.

Note: `--date` and `--media-type` are added to the top-level parser only (for legacy mode). When a subcommand is used, argparse routes to the subparser which also defines these args. This avoids duplicate argument conflicts.

```python
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
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
source .venv/bin/activate && python -m pytest tests/test_generate_podcast.py -v
```

Expected: all tests pass.

- [ ] **Step 5: Commit**

```bash
git add scripts/generate-podcast.py tests/test_generate_podcast.py
git commit -m "feat: add state file helpers and subcommand arg parsing"
```

---

### Task 3: Implement `start` subcommand

**Files:**
- Modify: `scripts/generate-podcast.py`
- Modify: `tests/test_generate_podcast.py`

- [ ] **Step 1: Write failing tests for `start` subcommand**

Add to `tests/test_generate_podcast.py`:

```python
class TestStartCommand:
    def test_start_creates_state_file(self, podcast_module, state_file, tmp_path):
        """start should create notebook, kick off generation, write state file."""
        # We can't call NotebookLM in tests, so we test the start_generation function
        # with mocked client
        import asyncio
        from unittest.mock import AsyncMock, MagicMock

        mock_client = AsyncMock()
        mock_nb = MagicMock()
        mock_nb.id = "nb-test-123"
        mock_client.notebooks.create.return_value = mock_nb
        mock_client.sources.add_text.return_value = None

        mock_status = MagicMock()
        mock_status.task_id = "task-test-456"
        mock_client.artifacts.generate_video.return_value = mock_status

        # Create a fake report
        reports_dir = tmp_path / "reports"
        reports_dir.mkdir()
        report = reports_dir / "2026-04-22.md"
        report.write_text("---\ntitle: Test\n---\n\nReport content here.")

        podcast_module.REPORTS_DIR = reports_dir

        asyncio.run(podcast_module.start_generation(mock_client, "2026-04-22", "video"))

        assert state_file.exists()
        data = json.loads(state_file.read_text())
        assert data["notebook_id"] == "nb-test-123"
        assert data["task_id"] == "task-test-456"
        assert data["date"] == "2026-04-22"
        assert data["media_type"] == "video"

    def test_start_skips_if_state_exists_same_date(self, podcast_module, state_file, capsys):
        """start should skip if state file exists for same date/media-type."""
        state_file.write_text(json.dumps({
            "notebook_id": "nb1", "task_id": "t1",
            "date": "2026-04-22", "media_type": "video",
            "started_at": "2026-04-22T08:00:00+00:00",
        }))
        import asyncio
        result = asyncio.run(podcast_module.start_generation(None, "2026-04-22", "video"))
        assert result is None  # Skipped, no new generation started
        captured = capsys.readouterr()
        assert "already in progress" in captured.out.lower()

    def test_start_errors_if_state_exists_different_date(self, podcast_module, state_file):
        """start should error if state file exists for a different date."""
        state_file.write_text(json.dumps({
            "notebook_id": "nb1", "task_id": "t1",
            "date": "2026-04-21", "media_type": "video",
            "started_at": "2026-04-21T08:00:00+00:00",
        }))
        import asyncio
        with pytest.raises(SystemExit) as exc_info:
            asyncio.run(podcast_module.start_generation(None, "2026-04-22", "video"))
        assert exc_info.value.code == 1
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
source .venv/bin/activate && python -m pytest tests/test_generate_podcast.py::TestStartCommand -v
```

Expected: failures because `start_generation` doesn't exist.

- [ ] **Step 3: Implement `start_generation`**

Add to `scripts/generate-podcast.py` after the state file helpers:

```python
async def start_generation(client, report_date: str, media_type: str) -> str | None:
    """Create notebook, add source, kick off generation, write state file.

    Returns task_id on success, None if skipped (already in progress).
    Exits with code 1 if state file exists for a different date.
    """
    existing = read_state()
    if existing:
        if existing["date"] == report_date and existing["media_type"] == media_type:
            print(f"Generation already in progress for {report_date}")
            return None
        else:
            print(
                f"Error: State file exists for a different date ({existing['date']}). "
                f"Run 'download' or delete .podcast-state.json first.",
                file=sys.stderr,
            )
            sys.exit(1)

    report_path = REPORTS_DIR / f"{report_date}.md"
    if not report_path.exists():
        print(f"Error: Report not found at {report_path}", file=sys.stderr)
        sys.exit(1)

    raw_content = report_path.read_text(encoding="utf-8")
    content = strip_front_matter(raw_content)

    if not content.strip():
        print(f"Error: Report {report_path} is empty after stripping front matter", file=sys.stderr)
        sys.exit(1)

    notebook_title = f"AI Daily Report — {report_date}"
    nb = await client.notebooks.create(notebook_title)
    print(f"Created notebook: {notebook_title} ({nb.id})")

    await client.sources.add_text(nb.id, notebook_title, content)
    print("Added report content as source")

    if media_type == "video":
        from notebooklm import VideoFormat, VideoStyle
        status = await client.artifacts.generate_video(
            nb.id,
            video_format=VideoFormat.EXPLAINER,
            video_style=VideoStyle.AUTO_SELECT,
        )
    else:
        from notebooklm import AudioFormat, AudioLength
        status = await client.artifacts.generate_audio(
            nb.id,
            audio_format=AudioFormat.DEEP_DIVE,
            audio_length=AudioLength.SHORT,
        )

    print(f"Started {media_type} generation (task: {status.task_id})")
    write_state(nb.id, status.task_id, report_date, media_type)
    return status.task_id
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
source .venv/bin/activate && python -m pytest tests/test_generate_podcast.py::TestStartCommand -v
```

Expected: all pass.

- [ ] **Step 5: Commit**

```bash
git add scripts/generate-podcast.py tests/test_generate_podcast.py
git commit -m "feat: implement start subcommand for async podcast generation"
```

---

### Task 4: Implement `poll` subcommand

**Files:**
- Modify: `scripts/generate-podcast.py`
- Modify: `tests/test_generate_podcast.py`

- [ ] **Step 1: Write failing tests for `poll` subcommand**

Add to `tests/test_generate_podcast.py`:

```python
class TestPollCommand:
    def test_poll_returns_complete(self, podcast_module, state_file):
        """poll should exit 0 and print 'complete' when done."""
        state_file.write_text(json.dumps({
            "notebook_id": "nb1", "task_id": "t1",
            "date": "2026-04-22", "media_type": "video",
            "started_at": "2026-04-22T08:00:00+00:00",
        }))
        import asyncio
        from unittest.mock import AsyncMock, MagicMock

        mock_client = AsyncMock()
        mock_status = MagicMock()
        mock_status.is_complete = True
        mock_status.is_failed = False
        mock_status.status = "completed"
        mock_client.artifacts.poll_status.return_value = mock_status

        exit_code = asyncio.run(podcast_module.poll_generation(mock_client))
        assert exit_code == 0

    def test_poll_returns_in_progress(self, podcast_module, state_file):
        """poll should exit 1 and print status when still working."""
        state_file.write_text(json.dumps({
            "notebook_id": "nb1", "task_id": "t1",
            "date": "2026-04-22", "media_type": "video",
            "started_at": "2026-04-22T08:00:00+00:00",
        }))
        import asyncio
        from unittest.mock import AsyncMock, MagicMock

        mock_client = AsyncMock()
        mock_status = MagicMock()
        mock_status.is_complete = False
        mock_status.is_failed = False
        mock_status.status = "in_progress"
        mock_client.artifacts.poll_status.return_value = mock_status

        exit_code = asyncio.run(podcast_module.poll_generation(mock_client))
        assert exit_code == 1

    def test_poll_returns_failed(self, podcast_module, state_file):
        """poll should exit 2 when generation failed."""
        state_file.write_text(json.dumps({
            "notebook_id": "nb1", "task_id": "t1",
            "date": "2026-04-22", "media_type": "video",
            "started_at": "2026-04-22T08:00:00+00:00",
        }))
        import asyncio
        from unittest.mock import AsyncMock, MagicMock

        mock_client = AsyncMock()
        mock_status = MagicMock()
        mock_status.is_complete = False
        mock_status.is_failed = True
        mock_status.status = "failed"
        mock_status.error = "Rate limit exceeded"
        mock_client.artifacts.poll_status.return_value = mock_status

        exit_code = asyncio.run(podcast_module.poll_generation(mock_client))
        assert exit_code == 2

    def test_poll_no_state_file(self, podcast_module, state_file):
        """poll should exit 1 when no state file exists."""
        import asyncio
        exit_code = asyncio.run(podcast_module.poll_generation(None))
        assert exit_code == 1
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
source .venv/bin/activate && python -m pytest tests/test_generate_podcast.py::TestPollCommand -v
```

Expected: failures because `poll_generation` doesn't exist.

- [ ] **Step 3: Implement `poll_generation`**

Add to `scripts/generate-podcast.py`:

```python
async def poll_generation(client) -> int:
    """Check the status of an in-progress generation.

    Returns exit code: 0=complete, 1=in-progress/no-state, 2=failed.
    """
    state = read_state()
    if not state:
        print("No generation in progress. Run 'start' first.")
        return 1

    status = await client.artifacts.poll_status(state["notebook_id"], state["task_id"])

    if status.is_complete:
        print("complete")
        return 0
    elif status.is_failed:
        print(f"failed: {status.error}", file=sys.stderr)
        return 2
    else:
        print(status.status)
        return 1
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
source .venv/bin/activate && python -m pytest tests/test_generate_podcast.py::TestPollCommand -v
```

Expected: all pass.

- [ ] **Step 5: Commit**

```bash
git add scripts/generate-podcast.py tests/test_generate_podcast.py
git commit -m "feat: implement poll subcommand for async podcast generation"
```

---

### Task 5: Implement `download` subcommand

**Files:**
- Modify: `scripts/generate-podcast.py`
- Modify: `tests/test_generate_podcast.py`

- [ ] **Step 1: Write failing tests for `download` subcommand**

Add to `tests/test_generate_podcast.py`:

```python
class TestDownloadCommand:
    def test_download_saves_file_and_cleans_up(self, podcast_module, state_file, tmp_path):
        """download should download artifact, delete notebook, remove state."""
        state_file.write_text(json.dumps({
            "notebook_id": "nb1", "task_id": "t1",
            "date": "2026-04-22", "media_type": "video",
            "started_at": "2026-04-22T08:00:00+00:00",
        }))
        podcasts_dir = tmp_path / "podcasts"
        podcast_module.PODCASTS_DIR = podcasts_dir

        import asyncio
        from unittest.mock import AsyncMock, MagicMock

        mock_client = AsyncMock()
        # poll_status returns complete
        mock_status = MagicMock()
        mock_status.is_complete = True
        mock_client.artifacts.poll_status.return_value = mock_status
        # download_video creates the file as a side effect
        async def fake_download(nb_id, path):
            Path(path).parent.mkdir(parents=True, exist_ok=True)
            Path(path).write_text("fake video")
        mock_client.artifacts.download_video.side_effect = fake_download

        output = asyncio.run(podcast_module.download_artifact(mock_client))
        assert output == podcasts_dir / "2026-04-22.mp4"
        assert (podcasts_dir / "2026-04-22.mp4").exists()
        assert not state_file.exists()
        mock_client.notebooks.delete.assert_called_once_with("nb1")

    def test_download_errors_if_not_complete(self, podcast_module, state_file):
        """download should error if generation is not yet complete."""
        state_file.write_text(json.dumps({
            "notebook_id": "nb1", "task_id": "t1",
            "date": "2026-04-22", "media_type": "video",
            "started_at": "2026-04-22T08:00:00+00:00",
        }))
        import asyncio
        from unittest.mock import AsyncMock, MagicMock

        mock_client = AsyncMock()
        mock_status = MagicMock()
        mock_status.is_complete = False
        mock_client.artifacts.poll_status.return_value = mock_status

        with pytest.raises(SystemExit) as exc_info:
            asyncio.run(podcast_module.download_artifact(mock_client))
        assert exc_info.value.code == 1

    def test_download_no_state_file(self, podcast_module, state_file):
        """download should error when no state file exists."""
        import asyncio
        with pytest.raises(SystemExit) as exc_info:
            asyncio.run(podcast_module.download_artifact(None))
        assert exc_info.value.code == 1

    def test_download_warns_on_cleanup_failure(self, podcast_module, state_file, tmp_path, capsys):
        """download should still succeed if notebook deletion fails."""
        state_file.write_text(json.dumps({
            "notebook_id": "nb1", "task_id": "t1",
            "date": "2026-04-22", "media_type": "audio",
            "started_at": "2026-04-22T08:00:00+00:00",
        }))
        podcasts_dir = tmp_path / "podcasts"
        podcast_module.PODCASTS_DIR = podcasts_dir

        import asyncio
        from unittest.mock import AsyncMock, MagicMock

        mock_client = AsyncMock()
        mock_status = MagicMock()
        mock_status.is_complete = True
        mock_client.artifacts.poll_status.return_value = mock_status

        async def fake_download(nb_id, path):
            Path(path).parent.mkdir(parents=True, exist_ok=True)
            Path(path).write_text("fake audio")
        mock_client.artifacts.download_audio.side_effect = fake_download
        mock_client.notebooks.delete.side_effect = Exception("API error")

        output = asyncio.run(podcast_module.download_artifact(mock_client))
        assert output == podcasts_dir / "2026-04-22.mp3"
        captured = capsys.readouterr()
        assert "warning" in captured.out.lower() or "warning" in captured.err.lower()
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
source .venv/bin/activate && python -m pytest tests/test_generate_podcast.py::TestDownloadCommand -v
```

Expected: failures because `download_artifact` doesn't exist.

- [ ] **Step 3: Implement `download_artifact`**

Add to `scripts/generate-podcast.py`:

```python
async def download_artifact(client) -> Path:
    """Download completed artifact, clean up notebook, remove state file.

    Returns path to the downloaded file.
    Exits with code 1 if no state file or generation not complete.
    """
    state = read_state()
    if not state:
        print("Error: No generation in progress. Run 'start' first.", file=sys.stderr)
        sys.exit(1)

    status = await client.artifacts.poll_status(state["notebook_id"], state["task_id"])
    if not status.is_complete:
        print("Error: Generation not complete. Run 'poll' to check status.", file=sys.stderr)
        sys.exit(1)

    PODCASTS_DIR.mkdir(exist_ok=True)
    ext = ".mp4" if state["media_type"] == "video" else ".mp3"
    output_path = PODCASTS_DIR / f"{state['date']}{ext}"

    if state["media_type"] == "video":
        await client.artifacts.download_video(state["notebook_id"], str(output_path))
    else:
        await client.artifacts.download_audio(state["notebook_id"], str(output_path))

    print(f"Downloaded {state['media_type']} to {output_path}")

    # Clean up notebook (warn on failure, don't fail the download)
    try:
        await client.notebooks.delete(state["notebook_id"])
        print(f"Cleaned up notebook {state['notebook_id']}")
    except Exception as e:
        print(f"Warning: Failed to delete notebook {state['notebook_id']}: {e}", file=sys.stderr)

    delete_state()
    return output_path
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
source .venv/bin/activate && python -m pytest tests/test_generate_podcast.py::TestDownloadCommand -v
```

Expected: all pass.

- [ ] **Step 5: Commit**

```bash
git add scripts/generate-podcast.py tests/test_generate_podcast.py
git commit -m "feat: implement download subcommand for async podcast generation"
```

---

### Task 6: Wire up `main()` to route subcommands

**Files:**
- Modify: `scripts/generate-podcast.py:115-130`

This task updates `main()` to dispatch to the correct function based on the subcommand, while preserving the original monolithic behavior when no subcommand is given.

- [ ] **Step 1: Write a failing test for subcommand routing**

Add to `tests/test_generate_podcast.py`:

```python
class TestMainRouting:
    def test_no_subcommand_calls_legacy_generate(self, podcast_module, monkeypatch):
        """No subcommand should call the original generate_podcast function."""
        called = {}
        async def fake_generate(date, media_type):
            called["date"] = date
            called["media_type"] = media_type
            return Path("/fake/output.mp4")

        monkeypatch.setattr(podcast_module, "generate_podcast", fake_generate)
        monkeypatch.setattr("sys.argv", ["generate-podcast.py", "--date", "2026-04-22", "--media-type", "video"])
        podcast_module.main()
        assert called["date"] == "2026-04-22"
        assert called["media_type"] == "video"
```

- [ ] **Step 2: Run test to verify it fails**

```bash
source .venv/bin/activate && python -m pytest tests/test_generate_podcast.py::TestMainRouting -v
```

Expected: fail (current main doesn't handle subcommand routing yet — though it may pass if parse_args still works in legacy mode; the real validation comes in step 4).

- [ ] **Step 3: Implement the new `main()` function**

Replace the existing `main()` function (lines 115-130) in `scripts/generate-podcast.py` with:

```python
def main() -> None:
    args = parse_args()

    if args.command is None:
        # Legacy mode: monolithic generate
        try:
            datetime.strptime(args.date, "%Y-%m-%d")
        except ValueError:
            print(f"Error: Invalid date format '{args.date}'. Use YYYY-MM-DD.", file=sys.stderr)
            sys.exit(1)

        output = asyncio.run(generate_podcast(args.date, args.media_type))
        print(f"\n{args.media_type.title()} generated: {output}")

    elif args.command == "start":
        try:
            datetime.strptime(args.date, "%Y-%m-%d")
        except ValueError:
            print(f"Error: Invalid date format '{args.date}'. Use YYYY-MM-DD.", file=sys.stderr)
            sys.exit(1)

        async def run_start():
            from notebooklm import NotebookLMClient
            async with await NotebookLMClient.from_storage() as client:
                await start_generation(client, args.date, args.media_type)

        asyncio.run(run_start())

    elif args.command == "poll":
        async def run_poll():
            from notebooklm import NotebookLMClient
            async with await NotebookLMClient.from_storage() as client:
                return await poll_generation(client)

        exit_code = asyncio.run(run_poll())
        sys.exit(exit_code)

    elif args.command == "download":
        async def run_download():
            from notebooklm import NotebookLMClient
            async with await NotebookLMClient.from_storage() as client:
                return await download_artifact(client)

        output = asyncio.run(run_download())
        print(f"\nDownload complete: {output}")
```

- [ ] **Step 4: Run all tests**

```bash
source .venv/bin/activate && python -m pytest tests/test_generate_podcast.py -v
```

Expected: all tests pass.

- [ ] **Step 5: Commit**

```bash
git add scripts/generate-podcast.py tests/test_generate_podcast.py
git commit -m "feat: wire up main() to route start/poll/download subcommands"
```

---

### Task 7: Update the daily-pipeline skill

**Files:**
- Modify: `.claude/skills/daily-pipeline.md:53-69`

- [ ] **Step 1: Replace Step 3 in the skill**

Replace the current Step 3 section (lines 53-69) in `.claude/skills/daily-pipeline.md` with:

```markdown
## Step 3: Generate Video

**Check:** Does `podcasts/DATE.mp4` already exist? If yes, skip this step and note "Video already exists for DATE".

If not, generate the video in three phases:

### Step 3a: Start Generation

Run:
```bash
python scripts/generate-podcast.py start --date DATE --media-type video
```

If this command fails, stop the pipeline and show the error.

Tell the user: "Video generation started. Polling for completion (this typically takes 10-30 minutes)."

### Step 3b: Poll for Completion

Run this command in a loop, waiting 30 seconds between each attempt:
```bash
python scripts/generate-podcast.py poll
```

- **Exit code 0** (prints `complete`): generation is done — proceed to Step 3c.
- **Exit code 1** (prints `pending` or `in_progress`): still working — wait 30 seconds and poll again.
- **Exit code 2** (prints error): generation failed — stop the pipeline and show the error.

Maximum 40 poll attempts (~20 minutes). If exceeded, stop the pipeline with: "Video generation timed out after 20 minutes of polling. The state file `.podcast-state.json` has been preserved for manual recovery."

Between polls, tell the user the current status (e.g., "Poll 5/40: in_progress").

### Step 3c: Download

Run:
```bash
python scripts/generate-podcast.py download
```

If this command fails, stop the pipeline and show the error.

Verify that `podcasts/DATE.mp4` was created.
```

- [ ] **Step 2: Verify the skill file is well-formed**

Read back `.claude/skills/daily-pipeline.md` and confirm:
- Steps 1, 2, 4, 5, 6 are unchanged
- Step 3 now has 3a, 3b, 3c sub-steps
- No references to the old blocking command remain

- [ ] **Step 3: Commit**

```bash
git add .claude/skills/daily-pipeline.md
git commit -m "feat: update daily-pipeline skill to use async polling for video generation"
```

---

### Task 8: Run full test suite and verify

**Files:** None (verification only)

- [ ] **Step 1: Run all tests**

```bash
source .venv/bin/activate && python -m pytest tests/test_generate_podcast.py -v
```

Expected: all tests pass.

- [ ] **Step 2: Verify legacy mode still works syntactically**

```bash
source .venv/bin/activate && python scripts/generate-podcast.py --help
```

Confirm output shows both subcommands (start, poll, download) and legacy --date/--media-type options.

- [ ] **Step 3: Verify subcommand help**

```bash
source .venv/bin/activate && python scripts/generate-podcast.py start --help
source .venv/bin/activate && python scripts/generate-podcast.py poll --help
source .venv/bin/activate && python scripts/generate-podcast.py download --help
```

Each should show its specific help text.

- [ ] **Step 4: Verify `.podcast-state.json` is gitignored**

```bash
echo "test" > .podcast-state.json && git status --porcelain .podcast-state.json
```

Expected: no output (file is ignored). Clean up:

```bash
rm .podcast-state.json
```
