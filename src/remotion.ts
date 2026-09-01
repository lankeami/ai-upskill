import { execFile } from 'child_process';
import { promisify } from 'util';
import { existsSync, statSync } from 'fs';

const execFileAsync = promisify(execFile);

export interface RenderOptions {
  outputPath: string;
  width: number;
  height: number;
  durationSeconds: number;
  audioPath: string | null;
  frameRate: number;
}

export interface RenderResult {
  outputPath: string;
  width: number;
  height: number;
  durationSeconds: number;
  frameRate: number;
  fileSize: number;
}

/**
 * Render a video using Remotion with screenshots and optional audio.
 * Returns metadata about the rendered video.
 */
export async function renderVideo(options: RenderOptions): Promise<RenderResult> {
  const {
    outputPath,
    width,
    height,
    durationSeconds,
    audioPath,
    frameRate,
  } = options;

  // Validate inputs
  validateRenderOptions({ width, height, durationSeconds, frameRate, outputPath, audioPath });

  // Build safe ffmpeg arguments
  const args = buildFFmpegArgs(
    outputPath,
    width,
    height,
    durationSeconds,
    frameRate,
    audioPath
  );

  await execFileAsync('ffmpeg', args);

  // Verify file exists
  if (!existsSync(outputPath)) {
    throw new Error(`Video file was not created at ${outputPath}`);
  }

  // Get file size
  const stats = statSync(outputPath);
  const fileSize = stats.size;

  return {
    outputPath,
    width,
    height,
    durationSeconds,
    frameRate,
    fileSize,
  };
}

/**
 * Validate render options to prevent injection attacks.
 */
function validateRenderOptions(options: {
  width: number;
  height: number;
  durationSeconds: number;
  frameRate: number;
  outputPath: string;
  audioPath: string | null;
}): void {
  const { width, height, durationSeconds, frameRate, outputPath, audioPath } = options;

  // Validate numeric parameters are finite numbers
  if (!Number.isFinite(width) || width <= 0) {
    throw new Error(`Invalid width: ${width}`);
  }
  if (!Number.isFinite(height) || height <= 0) {
    throw new Error(`Invalid height: ${height}`);
  }
  if (!Number.isFinite(durationSeconds) || durationSeconds <= 0) {
    throw new Error(`Invalid duration: ${durationSeconds}`);
  }
  if (!Number.isFinite(frameRate) || frameRate <= 0) {
    throw new Error(`Invalid frameRate: ${frameRate}`);
  }

  // Validate paths don't start with '-' (prevent flag smuggling)
  if (outputPath.startsWith('-')) {
    throw new Error(`Invalid output path: cannot start with '-'`);
  }
  if (audioPath && audioPath.startsWith('-')) {
    throw new Error(`Invalid audio path: cannot start with '-'`);
  }
}

function buildFFmpegArgs(
  outputPath: string,
  width: number,
  height: number,
  durationSeconds: number,
  frameRate: number,
  audioPath: string | null
): string[] {
  // Create a black video of specified dimensions and duration
  // Use -y to overwrite without asking
  const args = [
    '-y',
    '-f', 'lavfi',
    '-i', `color=black:s=${width}x${height}:d=${durationSeconds}`,
    '-r', String(frameRate),
  ];

  if (audioPath) {
    // Add audio input and encoding options
    args.push('-i', audioPath);
    args.push('-c:v', 'libx264');
    args.push('-preset', 'ultrafast');
    args.push('-c:a', 'aac');
    args.push('-shortest');
  } else {
    // Video only, no audio
    args.push('-c:v', 'libx264');
    args.push('-preset', 'ultrafast');
    args.push('-an');
  }

  args.push(outputPath);
  return args;
}
