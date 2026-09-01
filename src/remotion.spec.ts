import { describe, it, expect, beforeAll, afterAll } from 'vitest';
import { existsSync, unlinkSync } from 'fs';
import { dirname } from 'path';
import { renderVideo } from './remotion';

const TEST_OUTPUT_FILE = '/tmp/test-video.mp4';
const EXPECTED_DIMENSIONS = { width: 1920, height: 1080 };
const EXPECTED_DURATION_SECONDS = 5;

describe('Remotion video rendering', () => {
  afterAll(() => {
    if (existsSync(TEST_OUTPUT_FILE)) {
      unlinkSync(TEST_OUTPUT_FILE);
    }
  });

  it('creates a video file with correct dimensions and duration', async () => {
    const result = await renderVideo({
      outputPath: TEST_OUTPUT_FILE,
      width: EXPECTED_DIMENSIONS.width,
      height: EXPECTED_DIMENSIONS.height,
      durationSeconds: EXPECTED_DURATION_SECONDS,
      audioPath: null, // Test without audio first
      frameRate: 30,
    });

    // Verify video file exists
    expect(existsSync(TEST_OUTPUT_FILE)).toBe(true);

    // Verify result metadata
    expect(result).toBeDefined();
    expect(result.outputPath).toBe(TEST_OUTPUT_FILE);
    expect(result.width).toBe(EXPECTED_DIMENSIONS.width);
    expect(result.height).toBe(EXPECTED_DIMENSIONS.height);
    expect(result.durationSeconds).toBeCloseTo(EXPECTED_DURATION_SECONDS, 0);
    expect(result.fileSize).toBeGreaterThan(0);
  });

  it('matches frame rate and duration when rendering', async () => {
    const frameRate = 24;
    const durationSeconds = 3;

    const result = await renderVideo({
      outputPath: TEST_OUTPUT_FILE,
      width: 1280,
      height: 720,
      durationSeconds,
      audioPath: null,
      frameRate,
    });

    expect(result.frameRate).toBe(frameRate);
    expect(result.durationSeconds).toBeCloseTo(durationSeconds, 0);
  });
});
