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
