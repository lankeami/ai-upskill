import { describe, it, expect, beforeAll, afterAll } from 'vitest';
import { existsSync, unlinkSync } from 'fs';
import { execFile } from 'child_process';
import { promisify } from 'util';
import { composeSegmentedVideo } from './video-composition';

const execFileAsync = promisify(execFile);

const TEST_AUDIO = '/tmp/segments-test-audio.mp3';
const TEST_IMG_A = '/tmp/segments-img-a.png';
const TEST_IMG_B = '/tmp/segments-img-b.png';
const TEST_OUTPUT = '/tmp/segments-test-video.mp4';

describe('composeSegmentedVideo', () => {
  beforeAll(async () => {
    // 4-second silent audio
    await execFileAsync('ffmpeg', [
      '-y', '-f', 'lavfi', '-i', 'anullsrc=r=44100:cl=mono',
      '-t', '4', '-acodec', 'libmp3lame', TEST_AUDIO,
    ]);
    // Two solid-color frames of different sizes (tests scaling/padding)
    await execFileAsync('ffmpeg', [
      '-y', '-f', 'lavfi', '-i', 'color=red:s=1280x720:d=1',
      '-frames:v', '1', TEST_IMG_A,
    ]);
    await execFileAsync('ffmpeg', [
      '-y', '-f', 'lavfi', '-i', 'color=green:s=800x1200:d=1',
      '-frames:v', '1', TEST_IMG_B,
    ]);
  });

  afterAll(() => {
    [TEST_AUDIO, TEST_IMG_A, TEST_IMG_B, TEST_OUTPUT].forEach((f) => {
      if (existsSync(f)) unlinkSync(f);
    });
  });

  it('renders each segment for its specified duration with audio', async () => {
    const result = await composeSegmentedVideo({
      audioPath: TEST_AUDIO,
      segments: [
        { imagePath: TEST_IMG_A, durationSeconds: 1.5 },
        { imagePath: TEST_IMG_B, durationSeconds: 2.5 },
      ],
      outputPath: TEST_OUTPUT,
      width: 1920,
      height: 1080,
      frameRate: 30,
    });

    expect(existsSync(TEST_OUTPUT)).toBe(true);
    expect(result.segmentCount).toBe(2);
    expect(result.fileSize).toBeGreaterThan(0);
    // Total duration ≈ audio duration (4s)
    expect(result.durationSeconds).toBeCloseTo(4, 0);

    // Probe the file: video stream must be 1920x1080 with an audio track
    const { stdout } = await execFileAsync('ffprobe', [
      '-v', 'error',
      '-show_entries', 'stream=codec_type,width,height',
      '-of', 'default=noprint_wrappers=1',
      TEST_OUTPUT,
    ]);
    expect(stdout).toContain('codec_type=video');
    expect(stdout).toContain('codec_type=audio');
    expect(stdout).toContain('width=1920');
    expect(stdout).toContain('height=1080');
  });

  it('rejects empty segment list', async () => {
    await expect(
      composeSegmentedVideo({
        audioPath: TEST_AUDIO,
        segments: [],
        outputPath: TEST_OUTPUT,
        width: 1920,
        height: 1080,
        frameRate: 30,
      })
    ).rejects.toThrow(/at least one segment/i);
  });
});
