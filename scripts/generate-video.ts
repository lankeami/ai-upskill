#!/usr/bin/env node
/**
 * Generate a video for a daily report: screenshot each source article
 * page with Playwright, then compose them with the podcast audio so each
 * page is on screen while (approximately) being discussed.
 *
 * Timing: if reports/DATE.chapters.json exists, company sections use the
 * chapter timestamps; otherwise audio time is split equally per article.
 *
 * Usage: npx tsx scripts/generate-video.ts --date 2026-07-28 \
 *          --audio podcasts/2026-07-28.mp3 --output videos/2026-07-28.mp4
 */

import { program } from 'commander';
import { existsSync, readFileSync, mkdirSync } from 'fs';
import { join } from 'path';
import { extractReportSources } from '../src/report-sources';
import { captureSourceScreenshots } from '../src/screenshot';
import {
  composeSegmentedVideo,
  getAudioDuration,
  VideoSegment,
} from '../src/video-composition';

program
  .option('--date <date>', 'Report date (YYYY-MM-DD)')
  .option('--audio <path>', 'Path to audio file')
  .option('--output <path>', 'Output video file path')
  .option('--report <path>', 'Report markdown path (defaults to reports/DATE.md)')
  .option('--width <number>', 'Video width', '1920')
  .option('--height <number>', 'Video height', '1080')
  .option('--frame-rate <number>', 'Frame rate (fps)', '30');

program.parse();
const options = program.opts();

async function main() {
  const date = options.date || new Date().toISOString().split('T')[0];
  const audioPath = options.audio || `podcasts/${date}.mp3`;
  const outputPath = options.output || `videos/${date}.mp4`;
  const reportPath = options.report || `reports/${date}.md`;
  const width = parseInt(options.width, 10);
  const height = parseInt(options.height, 10);
  const frameRate = parseInt(options.frameRate, 10);

  if (!existsSync(audioPath)) {
    console.error(`Error: Audio file not found: ${audioPath}`);
    process.exit(1);
  }
  if (!existsSync(reportPath)) {
    console.error(`Error: Report not found: ${reportPath}`);
    process.exit(1);
  }

  console.log(`Generating video for ${date}...`);

  // 1. Extract source articles from the report
  const sources = extractReportSources(readFileSync(reportPath, 'utf-8'));
  if (sources.length === 0) {
    console.error('Error: No article sources found in report');
    process.exit(1);
  }
  console.log(`Found ${sources.length} source articles`);

  // 2. Screenshot every source page (one shared browser; failures skipped)
  const shotDir = `/tmp/report-sources-${date}`;
  mkdirSync(shotDir, { recursive: true });

  console.log('Capturing source page screenshots...');
  const captures = sources.map((s, i) => ({
    url: s.url,
    outputPath: join(shotDir, `${String(i).padStart(3, '0')}.png`),
  }));
  const results = await captureSourceScreenshots(captures, { width, height });

  const succeeded = results.filter((r) => r.ok);
  const failed = results.filter((r) => !r.ok);
  console.log(`✓ Captured ${succeeded.length}/${sources.length} pages`);
  failed.forEach((f) => console.warn(`  ⚠ skipped ${f.url}: ${f.error}`));

  if (succeeded.length === 0) {
    console.error('Error: No screenshots captured, cannot build video');
    process.exit(1);
  }

  // 3. Assign each captured page an equal slice of the audio.
  //    (chapters.json timing can refine this once transcripts are automated)
  const audioDuration = await getAudioDuration(audioPath);
  const perSegment = audioDuration / succeeded.length;
  const segments: VideoSegment[] = succeeded.map((r) => ({
    imagePath: r.outputPath,
    durationSeconds: perSegment,
  }));
  console.log(
    `Composing ${segments.length} segments × ${perSegment.toFixed(1)}s over ${audioDuration.toFixed(0)}s of audio...`
  );

  // 4. Compose
  const video = await composeSegmentedVideo({
    audioPath,
    segments,
    outputPath,
    width,
    height,
    frameRate,
  });

  console.log(`✓ Video created: ${video.outputPath}`);
  console.log(`  Duration: ${video.durationSeconds.toFixed(0)}s`);
  console.log(`  Segments: ${video.segmentCount}`);
  console.log(`  File size: ${(video.fileSize / 1024 / 1024).toFixed(2)} MB`);
}

main();
