#!/usr/bin/env node
/**
 * Generate a video from a daily report with Playwright screenshot and audio.
 * Usage: npx ts-node scripts/generate-video.ts --date 2026-07-28 --audio podcasts/2026-07-28.mp3 --output videos/2026-07-28.mp4
 */

import { program } from 'commander';
import { existsSync } from 'fs';
import { captureReportScreenshot } from '../src/screenshot';
import { composeVideoWithAudio } from '../src/video-composition';

program
  .option('--date <date>', 'Report date (YYYY-MM-DD)')
  .option('--audio <path>', 'Path to audio file')
  .option('--output <path>', 'Output video file path')
  .option('--width <number>', 'Video width', '1920')
  .option('--height <number>', 'Video height', '1080')
  .option('--frame-rate <number>', 'Frame rate (fps)', '30');

program.parse();

const options = program.opts();

async function main() {
  const date = options.date || new Date().toISOString().split('T')[0];
  const audioPath = options.audio || `podcasts/${date}.mp3`;
  const outputPath = options.output || `videos/${date}.mp4`;
  const width = parseInt(options.width, 10);
  const height = parseInt(options.height, 10);
  const frameRate = parseInt(options.frameRate, 10);

  console.log(`Generating video for ${date}...`);

  // Verify audio file exists
  if (!existsSync(audioPath)) {
    console.error(`Error: Audio file not found: ${audioPath}`);
    process.exit(1);
  }

  // Determine report URL (could be local file or running server)
  const reportUrl = `file://${process.cwd()}/reports/${date}.html`;

  try {
    // Capture screenshot of report page
    console.log('Capturing report screenshot...');
    const screenshot = await captureReportScreenshot({
      url: reportUrl,
      outputPath: `/tmp/report-${date}.png`,
      width: 1920,
      height: 1080,
      timeout: 30000,
    });

    console.log(`✓ Screenshot captured: ${screenshot.outputPath}`);

    // Compose video with audio
    console.log('Composing video with audio...');
    const video = await composeVideoWithAudio({
      audioPath,
      screenshots: [screenshot.outputPath],
      outputPath,
      width,
      height,
      frameRate,
    });

    console.log(`✓ Video created: ${video.outputPath}`);
    console.log(`  Duration: ${video.durationSeconds}s`);
    console.log(`  File size: ${(video.fileSize / 1024 / 1024).toFixed(2)} MB`);
  } catch (error) {
    console.error('Error generating video:', error);
    process.exit(1);
  }
}

main();
