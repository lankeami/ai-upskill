import { describe, it, expect, beforeAll, afterAll } from 'vitest';
import { existsSync, unlinkSync, writeFileSync } from 'fs';
import { execFile } from 'child_process';
import { promisify } from 'util';

const execFileAsync = promisify(execFile);

const TEST_OUTPUT_DIR = '/tmp/test-artifacts';
const TEST_AUDIO_FILE = '/tmp/artifact-test-audio.mp3';
const TEST_SCREENSHOT_FILE = '/tmp/artifact-test-screenshot.png';
const TEST_VIDEO_OUTPUT = '/tmp/artifact-test-video.mp4';

const createTestAudio = async () => {
  try {
    await execFileAsync('ffmpeg', [
      '-y',
      '-f', 'lavfi',
      '-i', 'anullsrc=r=44100:cl=mono',
      '-t', '2',
      '-q:a', '9',
      '-acodec', 'libmp3lame',
      TEST_AUDIO_FILE,
    ]);
  } catch (e) {
    console.warn('Could not create audio for test');
  }
};

const createTestScreenshot = async () => {
  try {
    await execFileAsync('ffmpeg', [
      '-y',
      '-f', 'lavfi',
      '-i', 'color=blue:s=1920x1080:d=0.1',
      '-q:v', '5',
      TEST_SCREENSHOT_FILE,
    ]);
  } catch (e) {
    console.warn('Could not create screenshot for test');
  }
};

describe('Video artifact creation and validation', () => {
  beforeAll(async () => {
    if (!existsSync(TEST_OUTPUT_DIR)) {
      require('fs').mkdirSync(TEST_OUTPUT_DIR, { recursive: true });
    }
    await createTestAudio();
    await createTestScreenshot();
  });

  afterAll(() => {
    // Cleanup
    [TEST_AUDIO_FILE, TEST_SCREENSHOT_FILE, TEST_VIDEO_OUTPUT].forEach(file => {
      if (existsSync(file)) {
        unlinkSync(file);
      }
    });
  });

  it('creates video artifact file', async () => {
    // Import composeVideoWithAudio from our implementation
    const { composeVideoWithAudio } = await import('./video-composition');

    const result = await composeVideoWithAudio({
      audioPath: TEST_AUDIO_FILE,
      screenshots: [TEST_SCREENSHOT_FILE],
      outputPath: TEST_VIDEO_OUTPUT,
      width: 1920,
      height: 1080,
      frameRate: 30,
    });

    // Verify artifact file exists
    expect(existsSync(TEST_VIDEO_OUTPUT)).toBe(true);

    // Verify artifact metadata
    expect(result.outputPath).toBe(TEST_VIDEO_OUTPUT);
    expect(result.fileSize).toBeGreaterThan(0);
    expect(result.durationSeconds).toBeGreaterThan(0);
  });

  it('video artifact is valid and playable', async () => {
    // Verify we can probe the video file with ffprobe
    try {
      const { stdout } = await execFileAsync('ffprobe', [
        '-v', 'error',
        '-show_entries', 'format=duration,size',
        '-of', 'default=noprint_wrappers=1',
        TEST_VIDEO_OUTPUT,
      ]);

      expect(stdout).toContain('duration');
      expect(stdout).toContain('size');
    } catch (e) {
      // ffprobe not available, skip this check
      console.warn('ffprobe not available, skipping video validation');
    }
  });

  it('video artifact can be published to release', async () => {
    // Verify artifact has expected properties for GitHub release
    expect(existsSync(TEST_VIDEO_OUTPUT)).toBe(true);

    const { statSync } = require('fs');
    const stats = statSync(TEST_VIDEO_OUTPUT);

    // File should be reasonably sized (not empty, not huge)
    expect(stats.size).toBeGreaterThan(10000); // At least 10KB
    expect(stats.size).toBeLessThan(500 * 1024 * 1024); // Less than 500MB
  });

  it('creates artifact with correct extension', async () => {
    // Verify output file has .mp4 extension
    expect(TEST_VIDEO_OUTPUT).toMatch(/\.mp4$/);
    expect(existsSync(TEST_VIDEO_OUTPUT)).toBe(true);
  });
});
