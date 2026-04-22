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


class TestStartCommand:
    def test_start_creates_state_file(self, podcast_module, state_file, tmp_path):
        """start should create notebook, kick off generation, write state file."""
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
        assert result is None
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
        mock_status = MagicMock()
        mock_status.is_complete = True
        mock_client.artifacts.poll_status.return_value = mock_status

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
