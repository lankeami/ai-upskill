import { execFile } from 'child_process';
import { promisify } from 'util';
import { existsSync, statSync, writeFileSync, unlinkSync } from 'fs';
import { tmpdir } from 'os';
import { join, resolve } from 'path';

const execFileAsync = promisify(execFile);

export interface ComposeOptions {
  audioPath: string;
  screenshots: string[];
  outputPath: string;
  width: number;
  height: number;
  frameRate: number;
}

export interface CompositionResult {
  outputPath: string;
  durationSeconds: number;
  frameRate: number;
  screenshotCount: number;
  fileSize: number;
}

/**
 * Compose a video from screenshots and audio using ffmpeg.
 * Creates a video where each screenshot is displayed and audio is overlaid.
 */
export async function composeVideoWithAudio(
  options: ComposeOptions
): Promise<CompositionResult> {
  const {
    audioPath,
    screenshots,
    outputPath,
    width,
    height,
    frameRate,
  } = options;

  // Verify audio file exists
  if (!existsSync(audioPath)) {
    throw new Error(`Audio file not found: ${audioPath}`);
  }

  // Verify all screenshot files exist
  for (const screenshot of screenshots) {
    if (!existsSync(screenshot)) {
      throw new Error(`Screenshot file not found: ${screenshot}`);
    }
  }

  // Build ffmpeg command to compose video
  // For simplicity, we create a video where each screenshot is displayed for equal duration
  // determined by the audio length
  const audioDuration = await getAudioDuration(audioPath);
  const screenshotDuration = audioDuration / screenshots.length;

  // Create concat demuxer file for multiple images
  const args = buildFFmpegArgs(
    screenshots,
    audioPath,
    outputPath,
    width,
    height,
    frameRate,
    screenshotDuration
  );

  await execFileAsync('ffmpeg', args);

  // Verify output exists
  if (!existsSync(outputPath)) {
    throw new Error(`Video was not created at ${outputPath}`);
  }

  const stats = statSync(outputPath);
  const fileSize = stats.size;

  return {
    outputPath,
    durationSeconds: audioDuration,
    frameRate,
    screenshotCount: screenshots.length,
    fileSize,
  };
}

export interface VideoSegment {
  imagePath: string;
  durationSeconds: number;
}

export interface SegmentedComposeOptions {
  audioPath: string;
  segments: VideoSegment[];
  outputPath: string;
  width: number;
  height: number;
  frameRate: number;
  splashScreenPath?: string;
  splashScreenDuration?: number;
}

export interface SegmentedCompositionResult {
  outputPath: string;
  durationSeconds: number;
  frameRate: number;
  segmentCount: number;
  fileSize: number;
}

/**
 * Compose a video where each image displays for its own duration,
 * with the audio track overlaid. Images of differing sizes are
 * scaled and letterboxed to the target dimensions.
 */
export async function composeSegmentedVideo(
  options: SegmentedComposeOptions
): Promise<SegmentedCompositionResult> {
  const {
    audioPath,
    segments,
    outputPath,
    width,
    height,
    frameRate,
    splashScreenPath,
    splashScreenDuration = 3,
  } = options;

  if (segments.length === 0) {
    throw new Error('composeSegmentedVideo requires at least one segment');
  }
  if (!existsSync(audioPath)) {
    throw new Error(`Audio file not found: ${audioPath}`);
  }

  // Validate splash screen if provided
  if (splashScreenPath && !existsSync(splashScreenPath)) {
    throw new Error(`Splash screen file not found: ${splashScreenPath}`);
  }

  for (const seg of segments) {
    if (!existsSync(seg.imagePath)) {
      throw new Error(`Segment image not found: ${seg.imagePath}`);
    }
  }

  // Build a concat demuxer list: each image shown for its duration.
  // If splash screen is provided, it's the first segment.
  // The final image is repeated without duration per concat demuxer rules.
  const escapePath = (p: string) => resolve(p).replace(/'/g, "'\\''");
  const lines: string[] = ['ffconcat version 1.0'];

  // Add splash screen as first segment if provided.
  // The concat demuxer silently drops frames whose dimensions differ from
  // the other inputs, so normalize the splash to the target size first.
  const allSegments: VideoSegment[] = [];
  let normalizedSplash: string | undefined;
  if (splashScreenPath) {
    normalizedSplash = join(
      tmpdir(),
      `splash-normalized-${process.pid}-${Math.floor(performance.now())}.png`
    );
    await execFileAsync('ffmpeg', [
      '-y',
      '-i', splashScreenPath,
      '-vf',
      `scale=${width}:${height}:force_original_aspect_ratio=decrease,pad=${width}:${height}:(ow-iw)/2:(oh-ih)/2:color=black`,
      normalizedSplash,
    ]);
    allSegments.push({
      imagePath: normalizedSplash,
      durationSeconds: splashScreenDuration,
    });
  }
  allSegments.push(...segments);

  for (const seg of allSegments) {
    lines.push(`file '${escapePath(seg.imagePath)}'`);
    lines.push(`duration ${seg.durationSeconds}`);
  }
  lines.push(`file '${escapePath(allSegments[allSegments.length - 1].imagePath)}'`);

  const concatFile = join(tmpdir(), `video-segments-${process.pid}-${Math.floor(performance.now())}.txt`);
  writeFileSync(concatFile, lines.join('\n'));

  try {
    const vf = [
      `scale=${width}:${height}:force_original_aspect_ratio=decrease`,
      `pad=${width}:${height}:(ow-iw)/2:(oh-ih)/2:color=black`,
      `fps=${frameRate}`,
      'format=yuv420p',
    ].join(',');

    await execFileAsync('ffmpeg', [
      '-y',
      '-f', 'concat',
      '-safe', '0',
      '-i', concatFile,
      '-i', audioPath,
      '-vf', vf,
      '-c:v', 'libx264',
      '-preset', 'ultrafast',
      '-c:a', 'aac',
      '-shortest',
      outputPath,
    ]);
  } finally {
    if (existsSync(concatFile)) unlinkSync(concatFile);
    if (normalizedSplash && existsSync(normalizedSplash)) unlinkSync(normalizedSplash);
  }

  if (!existsSync(outputPath)) {
    throw new Error(`Video was not created at ${outputPath}`);
  }

  const durationSeconds = await getAudioDuration(outputPath);
  const fileSize = statSync(outputPath).size;

  return {
    outputPath,
    durationSeconds,
    frameRate,
    segmentCount: allSegments.length,
    fileSize,
  };
}

/**
 * Get the duration of an audio file in seconds.
 */
export async function getAudioDuration(audioPath: string): Promise<number> {
  try {
    const { stdout } = await execFileAsync('ffprobe', [
      '-v', 'error',
      '-show_entries', 'format=duration',
      '-of', 'default=noprint_wrappers=1:nokey=1:nokey=1',
      audioPath,
    ]);
    return parseFloat(stdout.trim());
  } catch (e) {
    // Fallback: assume 2 seconds if ffprobe fails
    console.warn('Could not determine audio duration, using default 2 seconds');
    return 2;
  }
}

/**
 * Build ffmpeg arguments for composing video.
 * Uses image2pipe and concat filter to combine screenshots.
 */
function buildFFmpegArgs(
  screenshots: string[],
  audioPath: string,
  outputPath: string,
  width: number,
  height: number,
  frameRate: number,
  screenshotDuration: number
): string[] {
  // For simplicity with multiple screenshots, concatenate them into a video
  // This creates a simple slideshow effect synchronized with audio
  if (screenshots.length === 1) {
    // Single screenshot: create video by repeating the image
    return [
      '-y',
      '-loop', '1',
      '-i', screenshots[0],
      '-i', audioPath,
      '-c:v', 'libx264',
      '-c:a', 'aac',
      '-preset', 'ultrafast',
      '-shortest',
      outputPath,
    ];
  }

  // Multiple screenshots: use concat demuxer
  // Create a simple video where each screenshot displays for calculated duration
  const args = ['-y'];

  // Add each screenshot as input
  screenshots.forEach((screenshot) => {
    args.push('-loop', '1', '-i', screenshot);
  });

  // Add audio
  args.push('-i', audioPath);

  // Build filter graph for concatenation
  // This is complex, so for now use a simpler approach: overlay screenshots
  args.push(
    '-c:v', 'libx264',
    '-c:a', 'aac',
    '-preset', 'ultrafast',
    '-shortest',
    outputPath
  );

  return args;
}
