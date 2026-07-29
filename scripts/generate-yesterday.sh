#!/bin/bash
set -e

# Generate podcast and video for yesterday without updating repository
# Usage: ./scripts/generate-yesterday.sh

# Calculate yesterday's date (python3 is portable across macOS/Linux)
YESTERDAY=$(python3 -c "from datetime import datetime, timedelta, timezone; print((datetime.now(timezone.utc) - timedelta(days=1)).strftime('%Y-%m-%d'))")
echo "📅 Generating report, podcast, and video for: $YESTERDAY"
echo ""

# Check prerequisites
echo "🔍 Checking prerequisites..."

if ! command -v go &> /dev/null; then
  echo "❌ Go not found. Please install Go."
  exit 1
fi

if ! command -v python3 &> /dev/null; then
  echo "❌ Python 3 not found. Please install Python 3."
  exit 1
fi

if ! command -v node &> /dev/null; then
  echo "❌ Node.js not found. Please install Node.js."
  exit 1
fi

if [ -z "$NOTEBOOKLM_AUTH_JSON" ]; then
  echo "❌ NOTEBOOKLM_AUTH_JSON environment variable not set."
  echo "   Set it with: export NOTEBOOKLM_AUTH_JSON='<your-json-credentials>'"
  exit 1
fi

echo "✓ All prerequisites found"
echo ""

# Check if report already exists
if [ ! -f "reports/$YESTERDAY.md" ]; then
  echo "📝 Generating report for $YESTERDAY..."

  if [ ! -f "ai-report" ]; then
    echo "   Building Go CLI..."
    go build -o ai-report ./cmd/ai-report
  fi

  ./ai-report generate --date "$YESTERDAY"
  echo "✓ Report generated: reports/$YESTERDAY.md"
else
  echo "✓ Report already exists: reports/$YESTERDAY.md"
fi
echo ""

# Check if audio already exists
if [ ! -f "podcasts/$YESTERDAY.mp3" ]; then
  echo "🎙️  Generating podcast for $YESTERDAY..."
  echo "   Starting audio generation (this takes 5-10 minutes)..."

  python3 scripts/generate-podcast.py start --date "$YESTERDAY" --media-type audio

  echo "   Polling for completion..."
  POLL_COUNT=0
  MAX_POLLS=40
  while [ $POLL_COUNT -lt $MAX_POLLS ]; do
    POLL_COUNT=$((POLL_COUNT + 1))
    RESULT=$(python3 scripts/generate-podcast.py poll 2>&1 || true)

    if echo "$RESULT" | grep -q "complete"; then
      echo "   ✓ Audio generation complete"
      break
    fi

    echo "   Poll $POLL_COUNT/$MAX_POLLS: in progress..."
    sleep 30
  done

  if [ $POLL_COUNT -ge $MAX_POLLS ]; then
    echo "⚠️  Audio generation timed out after 20 minutes"
    echo "   The state is saved in .podcast-state.json for recovery"
    exit 1
  fi

  echo "   Downloading audio..."
  python3 scripts/generate-podcast.py download
  echo "✓ Podcast generated: podcasts/$YESTERDAY.mp3"
else
  echo "✓ Podcast already exists: podcasts/$YESTERDAY.mp3"
fi
echo ""

# Generate chapters (optional, requires transcript)
if [ -f "podcasts/$YESTERDAY.vtt" ]; then
  if [ ! -f "reports/$YESTERDAY.chapters.json" ]; then
    echo "📖 Generating chapters from transcript..."
    python3 scripts/generate_chapters.py \
      --report "reports/$YESTERDAY.md" \
      --transcript "podcasts/$YESTERDAY.vtt" \
      --output "reports/$YESTERDAY.chapters.json"
    echo "✓ Chapters generated: reports/$YESTERDAY.chapters.json"
  else
    echo "✓ Chapters already exist: reports/$YESTERDAY.chapters.json"
  fi
else
  echo "⚠️  Transcript not found (podcasts/$YESTERDAY.vtt)"
  echo "   Download from NotebookLM manually to generate chapters"
fi
echo ""

# Generate video
if [ ! -f "videos/$YESTERDAY.mp4" ]; then
  echo "🎬 Generating video for $YESTERDAY..."

  mkdir -p videos

  npx tsx scripts/generate-video.ts \
    --date "$YESTERDAY" \
    --audio "podcasts/$YESTERDAY.mp3" \
    --output "videos/$YESTERDAY.mp4"

  if [ -f "videos/$YESTERDAY.mp4" ]; then
    VIDEO_SIZE=$(ls -lh "videos/$YESTERDAY.mp4" | awk '{print $5}')
    echo "✓ Video generated: videos/$YESTERDAY.mp4 ($VIDEO_SIZE)"
  else
    echo "❌ Video generation failed"
    exit 1
  fi
else
  echo "✓ Video already exists: videos/$YESTERDAY.mp4"
fi
echo ""

# Summary
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ Generation complete for $YESTERDAY"
echo ""
echo "Generated files:"
echo "  📄 Report:     reports/$YESTERDAY.md"
echo "  🎙️  Podcast:    podcasts/$YESTERDAY.mp3"
echo "  📖 Chapters:   reports/$YESTERDAY.chapters.json (if transcript available)"
echo "  🎬 Video:      videos/$YESTERDAY.mp4"
echo ""
echo "💡 Next steps:"
echo "  - Review the video: videos/$YESTERDAY.mp4"
echo "  - To publish as a release, run:"
echo "    gh release create podcast-$YESTERDAY \\\"
echo "      podcasts/$YESTERDAY.mp3 \\\"
echo "      videos/$YESTERDAY.mp4 \\\"
echo "      --title \"Podcast — $YESTERDAY\" \\\"
echo "      --notes \"Daily AI Report for $YESTERDAY\""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
