import { describe, it, expect, beforeAll, afterAll } from 'vitest';
import { existsSync, unlinkSync } from 'fs';
import { execFile } from 'child_process';
import { promisify } from 'util';
import { composeVideoWithAudio, composeSegmentedVideo, VideoSegment } from './video-composition';

const execFileAsync = promisify(execFile);

const TEST_OUTPUT_DIR = '/tmp/test-videos';
const TEST_AUDIO_FILE = '/tmp/test-audio.wav';
const TEST_SCREENSHOT_FILE = '/tmp/test-screenshot.png';

// Create a minimal test audio file (2 seconds of silence)
const createTestAudio = async () => {
  try {
    // Create 2 seconds of silence at 44100Hz
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
    console.warn('Could not create audio with ffmpeg, test may be skipped');
  }
};

// Create a minimal test screenshot (PNG)
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
    console.warn('Could not create screenshot with ffmpeg, test may be skipped');
  }
};

describe('Video composition with audio sync', () => {
  beforeAll(async () => {
    if (!existsSync(TEST_OUTPUT_DIR)) {
      require('fs').mkdirSync(TEST_OUTPUT_DIR, { recursive: true });
    }
    await createTestAudio();
    await createTestScreenshot();
  });

  afterAll(() => {
    // Cleanup
    if (existsSync(TEST_AUDIO_FILE)) {
      unlinkSync(TEST_AUDIO_FILE);
    }
    if (existsSync(TEST_SCREENSHOT_FILE)) {
      unlinkSync(TEST_SCREENSHOT_FILE);
    }
    const dir = TEST_OUTPUT_DIR;
    if (existsSync(dir)) {
      const files = require('fs').readdirSync(dir);
      files.forEach((file: string) => {
        unlinkSync(`${dir}/${file}`);
      });
    }
  });

  it('creates video with matching audio duration', async () => {
    const outputPath = `${TEST_OUTPUT_DIR}/sync-test-video.mp4`;

    const result = await composeVideoWithAudio({
      audioPath: TEST_AUDIO_FILE,
      screenshots: [TEST_SCREENSHOT_FILE],
      outputPath,
      width: 1920,
      height: 1080,
      frameRate: 30,
    });

    expect(existsSync(outputPath)).toBe(true);
    expect(result.outputPath).toBe(outputPath);
    expect(result.durationSeconds).toBeGreaterThan(0);
    expect(result.fileSize).toBeGreaterThan(0);
  });

  it('maintains frame rate throughout video', async () => {
    const outputPath = `${TEST_OUTPUT_DIR}/framerate-test-video.mp4`;
    const frameRate = 24;

    const result = await composeVideoWithAudio({
      audioPath: TEST_AUDIO_FILE,
      screenshots: [TEST_SCREENSHOT_FILE],
      outputPath,
      width: 1280,
      height: 720,
      frameRate,
    });

    expect(result.frameRate).toBe(frameRate);
    expect(existsSync(outputPath)).toBe(true);
  });

  it('handles multiple screenshots in sequence', async () => {
    // For this test, reuse the same screenshot multiple times
    const screenshots = [
      TEST_SCREENSHOT_FILE,
      TEST_SCREENSHOT_FILE,
      TEST_SCREENSHOT_FILE,
    ];
    const outputPath = `${TEST_OUTPUT_DIR}/multi-screenshot-video.mp4`;

    const result = await composeVideoWithAudio({
      audioPath: TEST_AUDIO_FILE,
      screenshots,
      outputPath,
      width: 1920,
      height: 1080,
      frameRate: 30,
    });

    expect(existsSync(outputPath)).toBe(true);
    expect(result.screenshotCount).toBe(screenshots.length);
  });

  it('matches video duration to audio duration', async () => {
    const outputPath = `${TEST_OUTPUT_DIR}/duration-test-video.mp4`;

    const result = await composeVideoWithAudio({
      audioPath: TEST_AUDIO_FILE,
      screenshots: [TEST_SCREENSHOT_FILE],
      outputPath,
      width: 1920,
      height: 1080,
      frameRate: 30,
    });

    // Audio file is 2 seconds, video should match
    expect(result.durationSeconds).toBeCloseTo(2, 0);
  });

  it('displays splash screen as the first frame even when its dimensions differ from content', async () => {
    const outputPath = `${TEST_OUTPUT_DIR}/splash-4k-video.mp4`;
    const splash4k = '/tmp/test-splash-4k.png';
    // Red 4K splash vs blue 1080p content — concat demuxer drops mismatched
    // first frames unless the splash is normalized to target dimensions
    await execFileAsync('ffmpeg', [
      '-y', '-f', 'lavfi', '-i', 'color=red:s=3840x2160:d=0.1', '-frames:v', '1', splash4k,
    ]);

    await composeSegmentedVideo({
      audioPath: TEST_AUDIO_FILE,
      segments: [{ imagePath: TEST_SCREENSHOT_FILE, durationSeconds: 1 }],
      splashScreenPath: splash4k,
      splashScreenDuration: 1,
      outputPath,
      width: 1920,
      height: 1080,
      frameRate: 30,
    });

    // Extract first frame as raw RGB and check the center pixel is red
    const rawFrame = `${TEST_OUTPUT_DIR}/first-frame.rgb`;
    await execFileAsync('ffmpeg', [
      '-y', '-i', outputPath, '-frames:v', '1', '-f', 'rawvideo', '-pix_fmt', 'rgb24', rawFrame,
    ]);
    const buf = require('fs').readFileSync(rawFrame);
    const centerOffset = ((1080 / 2) * 1920 + 1920 / 2) * 3;
    const [r, g, b] = [buf[centerOffset], buf[centerOffset + 1], buf[centerOffset + 2]];
    expect(r).toBeGreaterThan(150); // red channel dominant
    expect(g).toBeLessThan(100);
    expect(b).toBeLessThan(100);

    unlinkSync(splash4k);
  });

  it('includes splash screen as first segment when provided', async () => {
    const outputPath = `${TEST_OUTPUT_DIR}/splash-screen-video.mp4`;
    const splashScreenDuration = 3; // 3 seconds for splash

    const segments: VideoSegment[] = [
      {
        imagePath: TEST_SCREENSHOT_FILE,
        durationSeconds: 2,
      },
    ];

    const result = await composeSegmentedVideo({
      audioPath: TEST_AUDIO_FILE,
      segments,
      splashScreenPath: TEST_SCREENSHOT_FILE,
      splashScreenDuration,
      outputPath,
      width: 1920,
      height: 1080,
      frameRate: 30,
    });

    expect(existsSync(outputPath)).toBe(true);
    expect(result.outputPath).toBe(outputPath);
    // Video should be: splash (3s) + segment (2s) = 5s, but we're mixing with 2s audio
    // so it will be clamped by audio duration or extended. Result should have splash included.
    expect(result.segmentCount).toBe(2); // splash + 1 content segment
  });
});
