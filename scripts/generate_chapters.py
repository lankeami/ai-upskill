#!/usr/bin/env python3
"""
Generate timestamped chapters from a report markdown and NotebookLM transcript.
Matches article headings to transcript text, calculates timestamps with fuzzy matching.

Usage:
  python generate_chapters.py \
    --report reports/2026-07-28.md \
    --transcript podcast_transcript_2026-07-28.vtt \
    --output reports/2026-07-28.chapters.json
"""

import json
import re
import argparse
from pathlib import Path
from typing import List, Dict, Tuple
from difflib import SequenceMatcher
import sys


def parse_vtt_transcript(vtt_path: str) -> List[Tuple[int, str]]:
    """Parse VTT transcript and return list of (start_time_seconds, text) tuples."""
    with open(vtt_path, 'r') as f:
        content = f.read()

    blocks = content.split('\n\n')[1:]  # Skip WEBVTT header
    entries = []

    for block in blocks:
        lines = block.strip().split('\n')
        if len(lines) < 2:
            continue

        timestamp = lines[0]  # "00:00:15.500 --> 00:00:18.200"
        text = ' '.join(lines[1:])

        # Extract start time
        start_match = re.match(r'(\d{2}):(\d{2}):(\d{2})', timestamp)
        if start_match:
            h, m, s = map(int, start_match.groups())
            total_seconds = h * 3600 + m * 60 + s
            entries.append((total_seconds, text))

    return sorted(entries, key=lambda x: x[0])


def parse_report_sections(markdown_path: str) -> List[Dict]:
    """Extract article sections from report markdown."""
    with open(markdown_path, 'r') as f:
        content = f.read()

    sections = []
    # Match markdown headings (## or ### only, skip top-level #)
    heading_pattern = r'^(#{2,3})\s+(.+)$'

    for match in re.finditer(heading_pattern, content, re.MULTILINE):
        level = len(match.group(1))
        title = match.group(2).strip()
        sections.append({'title': title, 'level': level})

    return sections


def extract_keywords(text: str) -> List[str]:
    """Extract meaningful keywords from text for matching."""
    # Remove URLs and markdown formatting
    clean = re.sub(r'\[([^\]]+)\]\([^\)]+\)', r'\1', text)
    clean = re.sub(r'[*_~`]', '', clean)

    # Split and filter short words
    words = [w.lower() for w in clean.split() if len(w) > 3]

    # Remove common stop words that add noise
    stop_words = {
        'that', 'this', 'with', 'from', 'have', 'were', 'will', 'been',
        'more', 'also', 'just', 'new', 'best', 'like', 'some', 'many',
        'other', 'such', 'even', 'year', 'work', 'says', 'said', 'much'
    }

    return [w for w in words if w not in stop_words]


def fuzzy_match(section_keywords: List[str], transcript_entries: List[Tuple[int, str]]) -> Tuple[int, float]:
    """Find best matching time in transcript using fuzzy keyword matching.

    Returns: (best_timestamp, confidence_score)
    """
    if not section_keywords:
        return None, 0.0

    best_time = None
    best_score = 0.0

    for timestamp, text in transcript_entries:
        transcript_keywords = extract_keywords(text)

        # Count keyword matches and calculate match ratio
        matches = sum(1 for kw in section_keywords if kw in transcript_keywords)
        match_ratio = matches / len(section_keywords) if section_keywords else 0.0

        # Bonus for exact phrase matches (fuzzy)
        text_lower = text.lower()
        phrase_score = 0.0

        # Try matching 2-3 word phrases
        for i in range(len(section_keywords) - 1):
            phrase = ' '.join(section_keywords[i:i+2])
            if phrase in text_lower:
                phrase_score += 0.5

        total_score = match_ratio + (phrase_score * 0.1)

        if total_score > best_score:
            best_score = total_score
            best_time = timestamp

    return best_time, best_score


def find_section_start_time(section_title: str, transcript_entries: List[Tuple[int, str]]) -> Tuple[int, float]:
    """Find start time of section in transcript with confidence score."""
    keywords = extract_keywords(section_title)

    if not keywords:
        return None, 0.0

    best_time, confidence = fuzzy_match(keywords, transcript_entries)
    return best_time, confidence


def seconds_to_timestamp(seconds: int) -> str:
    """Convert seconds to HH:MM:SS format."""
    hours = seconds // 3600
    minutes = (seconds % 3600) // 60
    secs = seconds % 60
    return f"{hours:02d}:{minutes:02d}:{secs:02d}"


def generate_chapters(report_path: str, transcript_path: str) -> Tuple[List[Dict], List[Dict]]:
    """Generate chapters with timestamps and confidence scores.

    Returns: (chapters_list, unmatched_sections)
    """
    try:
        transcript_entries = parse_vtt_transcript(transcript_path)
    except FileNotFoundError:
        print(f"Error: Transcript file not found: {transcript_path}", file=sys.stderr)
        return [], []

    try:
        sections = parse_report_sections(report_path)
    except FileNotFoundError:
        print(f"Error: Report file not found: {report_path}", file=sys.stderr)
        return [], []

    if not transcript_entries:
        print("Error: No transcript entries found", file=sys.stderr)
        return [], []

    chapters = []
    unmatched = []

    for section in sections:
        start_time, confidence = find_section_start_time(section['title'], transcript_entries)

        # Only include if confidence is reasonable (at least 2+ keyword matches)
        if start_time is not None and confidence > 0.5:
            chapters.append({
                'title': section['title'],
                'start_seconds': start_time,
                'start_timestamp': seconds_to_timestamp(start_time),
                'level': section['level'],
                'confidence': round(confidence, 2)
            })
        else:
            unmatched.append({
                'title': section['title'],
                'level': section['level'],
                'confidence': round(confidence, 2),
                'reason': 'Low confidence match' if confidence else 'No keywords found'
            })

    # Sort by timestamp
    chapters.sort(key=lambda x: x['start_seconds'])

    # Calculate end times
    max_time = max(t for t, _ in transcript_entries) if transcript_entries else 0
    for i, chapter in enumerate(chapters):
        if i + 1 < len(chapters):
            chapter['end_seconds'] = chapters[i + 1]['start_seconds']
        else:
            chapter['end_seconds'] = max_time

        chapter['end_timestamp'] = seconds_to_timestamp(chapter['end_seconds'])

    return chapters, unmatched


def main():
    parser = argparse.ArgumentParser(
        description='Generate timestamped chapters from report and transcript'
    )
    parser.add_argument('--report', required=True, help='Path to report markdown')
    parser.add_argument('--transcript', required=True, help='Path to VTT transcript')
    parser.add_argument('--output', required=True, help='Output JSON file')
    parser.add_argument('--verbose', action='store_true', help='Show matching details')

    args = parser.parse_args()

    chapters, unmatched = generate_chapters(args.report, args.transcript)

    if args.verbose and unmatched:
        print(f"Note: {len(unmatched)} sections could not be matched with high confidence:", file=sys.stderr)
        for um in unmatched:
            print(f"  - {um['title']} ({um['reason']}, confidence: {um['confidence']})", file=sys.stderr)

    output = {
        'report': Path(args.report).name,
        'transcript': Path(args.transcript).name,
        'total_sections': len(chapters) + len(unmatched),
        'matched_sections': len(chapters),
        'chapters': chapters
    }

    with open(args.output, 'w') as f:
        json.dump(output, f, indent=2)

    print(f"Generated {len(chapters)} chapters → {args.output}")
    for ch in chapters:
        confidence_str = f" (confidence: {ch['confidence']})" if ch['confidence'] < 0.8 else ""
        print(f"  {ch['start_timestamp']} {ch['title']}{confidence_str}")

    if unmatched and not args.verbose:
        print(f"({len(unmatched)} sections unmatched — use --verbose for details)")


if __name__ == '__main__':
    main()
