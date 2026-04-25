"""Tests for upload-youtube.py helper functions."""

import json
import sys
from pathlib import Path
from unittest.mock import MagicMock, patch

import pytest

import importlib.util


def load_module():
    spec = importlib.util.spec_from_file_location(
        "upload_youtube",
        Path(__file__).resolve().parent.parent / "scripts" / "upload-youtube.py",
    )
    mod = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mod)
    return mod


@pytest.fixture
def mod():
    return load_module()


class TestBuildTitle:
    def test_formats_date_correctly(self, mod):
        assert mod.build_title("2026-04-22") == "Toiletpaper Press \u2014 AI Daily 2026-04-22"

    def test_different_date(self, mod):
        assert mod.build_title("2026-01-01") == "Toiletpaper Press \u2014 AI Daily 2026-01-01"


class TestExtractDescription:
    def test_extracts_first_five_bullets(self, mod, tmp_path):
        report = tmp_path / "2026-04-22.md"
        report.write_text(
            "---\ntitle: Test\ndate: 2026-04-22\n---\n\n"
            "## Section One\n"
            "- **Item 1** \u2014 Source 1\n"
            "  Summary one.\n"
            "- **Item 2** \u2014 Source 2\n"
            "  Summary two.\n"
            "- **Item 3** \u2014 Source 3\n"
            "  Summary three.\n"
            "- **Item 4** \u2014 Source 4\n"
            "  Summary four.\n"
            "- **Item 5** \u2014 Source 5\n"
            "  Summary five.\n"
            "- **Item 6** \u2014 Source 6\n"
            "  Summary six.\n"
        )
        desc = mod.extract_description("2026-04-22", reports_dir=tmp_path)
        lines = desc.split("\n")
        bullet_lines = [l for l in lines if l.startswith("- ")]
        assert len(bullet_lines) == 5

    def test_appends_report_link(self, mod, tmp_path):
        report = tmp_path / "2026-04-22.md"
        report.write_text(
            "---\ntitle: Test\n---\n\n- **Item 1** \u2014 Source\n  Summary.\n"
        )
        desc = mod.extract_description("2026-04-22", reports_dir=tmp_path)
        assert "https://lankeami.github.io/ai-upskill/reports/2026-04-22" in desc

    def test_handles_fewer_than_five_bullets(self, mod, tmp_path):
        report = tmp_path / "2026-04-22.md"
        report.write_text(
            "---\ntitle: Test\n---\n\n- **Item 1** \u2014 Source\n  Summary.\n"
            "- **Item 2** \u2014 Source\n  Summary.\n"
        )
        desc = mod.extract_description("2026-04-22", reports_dir=tmp_path)
        bullet_lines = [l for l in desc.split("\n") if l.startswith("- ")]
        assert len(bullet_lines) == 2

    def test_strips_front_matter(self, mod, tmp_path):
        report = tmp_path / "2026-04-22.md"
        report.write_text(
            "---\ntitle: Test\npodcast_url: http://example.com\n---\n\n"
            "- **Item 1** \u2014 Source\n  Summary.\n"
        )
        desc = mod.extract_description("2026-04-22", reports_dir=tmp_path)
        assert "podcast_url" not in desc
        assert "title:" not in desc


class TestLoadCredentials:
    def test_loads_token_from_file(self, mod, tmp_path, monkeypatch):
        """load_credentials should return Credentials when token file exists."""
        token_data = {
            "token": "tok",
            "refresh_token": "ref",
            "token_uri": "https://oauth2.googleapis.com/token",
            "client_id": "cid",
            "client_secret": "csec",
            "scopes": ["https://www.googleapis.com/auth/youtube.upload"],
        }
        token_file = tmp_path / ".youtube-token.json"
        token_file.write_text(json.dumps(token_data))

        monkeypatch.setattr(mod, "TOKEN_FILE", token_file)
        monkeypatch.setenv("YOUTUBE_CLIENT_ID", "cid")
        monkeypatch.setenv("YOUTUBE_CLIENT_SECRET", "csec")

        mock_creds = MagicMock()
        mock_creds.valid = True
        mock_creds.expired = False
        with patch("google.oauth2.credentials.Credentials.from_authorized_user_file", return_value=mock_creds):
            creds = mod.load_credentials()
            assert creds is not None

    def test_exits_when_token_missing(self, mod, tmp_path, monkeypatch):
        """load_credentials should exit 1 when token file is missing."""
        monkeypatch.setattr(mod, "TOKEN_FILE", tmp_path / ".youtube-token.json")
        monkeypatch.setenv("YOUTUBE_CLIENT_ID", "cid")
        monkeypatch.setenv("YOUTUBE_CLIENT_SECRET", "csec")

        with pytest.raises(SystemExit) as exc_info:
            mod.load_credentials()
        assert exc_info.value.code == 1

    def test_exits_when_client_id_missing(self, mod, tmp_path, monkeypatch):
        """load_credentials should exit 1 when YOUTUBE_CLIENT_ID is not set."""
        monkeypatch.delenv("YOUTUBE_CLIENT_ID", raising=False)
        monkeypatch.delenv("YOUTUBE_CLIENT_SECRET", raising=False)
        monkeypatch.setattr(mod, "TOKEN_FILE", tmp_path / ".youtube-token.json")

        with pytest.raises(SystemExit) as exc_info:
            mod.load_credentials()
        assert exc_info.value.code == 1


class TestFindExistingVideo:
    def test_returns_url_when_title_matches(self, mod):
        """find_existing_video should return URL if a video with matching title exists."""
        mock_youtube = MagicMock()
        mock_youtube.search.return_value.list.return_value.execute.return_value = {
            "items": [
                {
                    "id": {"videoId": "abc123"},
                    "snippet": {"title": "Toiletpaper Press \u2014 AI Daily 2026-04-22"},
                }
            ]
        }

        result = mod.find_existing_video(mock_youtube, "Toiletpaper Press \u2014 AI Daily 2026-04-22")
        assert result == "https://youtube.com/watch?v=abc123"

    def test_returns_none_when_no_match(self, mod):
        """find_existing_video should return None when no matching video exists."""
        mock_youtube = MagicMock()
        mock_youtube.search.return_value.list.return_value.execute.return_value = {
            "items": [
                {
                    "id": {"videoId": "xyz999"},
                    "snippet": {"title": "Some other video"},
                }
            ]
        }

        result = mod.find_existing_video(mock_youtube, "Toiletpaper Press \u2014 AI Daily 2026-04-22")
        assert result is None

    def test_returns_none_when_empty_results(self, mod):
        """find_existing_video should return None when search returns no items."""
        mock_youtube = MagicMock()
        mock_youtube.search.return_value.list.return_value.execute.return_value = {"items": []}

        result = mod.find_existing_video(mock_youtube, "Toiletpaper Press \u2014 AI Daily 2026-04-22")
        assert result is None
