import { describe, it, expect } from 'vitest';
import { buildReportPlayerHTML } from './report-ui';

describe('Report page UI - audio and video players', () => {
  it('renders both audio and video players', () => {
    const html = buildReportPlayerHTML({
      date: '2026-07-28',
      audioUrl: 'https://github.com/lankeami/ai-upskill/releases/download/podcast-2026-07-28/2026-07-28.mp3',
      videoUrl: 'https://github.com/lankeami/ai-upskill/releases/download/podcast-2026-07-28/2026-07-28.mp4',
    });

    expect(html).toContain('<audio');
    expect(html).toContain('<video');
    expect(html).toContain('controls');
  });

  it('includes correct audio source URL', () => {
    const audioUrl = 'https://example.com/audio.mp3';
    const html = buildReportPlayerHTML({
      date: '2026-07-28',
      audioUrl,
      videoUrl: 'https://example.com/video.mp4',
    });

    expect(html).toContain(audioUrl);
    expect(html).toContain('type="audio/mpeg"');
    expect(html).toContain('<audio');
  });

  it('includes correct video source URL', () => {
    const videoUrl = 'https://example.com/video.mp4';
    const html = buildReportPlayerHTML({
      date: '2026-07-28',
      audioUrl: 'https://example.com/audio.mp3',
      videoUrl,
    });

    expect(html).toContain(videoUrl);
    expect(html).toContain('type="video/mp4"');
    expect(html).toContain('<video');
  });

  it('includes video player with correct dimensions', () => {
    const html = buildReportPlayerHTML({
      date: '2026-07-28',
      audioUrl: 'https://example.com/audio.mp3',
      videoUrl: 'https://example.com/video.mp4',
      videoWidth: 1920,
      videoHeight: 1080,
    });

    expect(html).toContain('width="1920"');
    expect(html).toContain('height="1080"');
  });

  it('player HTML is valid and accessible', () => {
    const html = buildReportPlayerHTML({
      date: '2026-07-28',
      audioUrl: 'https://example.com/audio.mp3',
      videoUrl: 'https://example.com/video.mp4',
    });

    // Check for accessibility attributes
    expect(html).toContain('controls');
    expect(html).toContain('poster'); // Video poster image

    // Check that players have labels/descriptions
    expect(html).toMatch(/(?:Audio|Podcast|Listen)/i);
    expect(html).toMatch(/(?:Video|Watch)/i);
  });

  it('renders gracefully when video URL is not provided', () => {
    const html = buildReportPlayerHTML({
      date: '2026-07-28',
      audioUrl: 'https://example.com/audio.mp3',
      videoUrl: null,
    });

    expect(html).toContain('<audio');
    expect(html).toContain('controls');
    // Should not have broken video element
    expect(html).not.toContain('<video');
  });

  it('renders gracefully when audio URL is not provided', () => {
    const html = buildReportPlayerHTML({
      date: '2026-07-28',
      audioUrl: null,
      videoUrl: 'https://example.com/video.mp4',
    });

    expect(html).toContain('<video');
    expect(html).toContain('controls');
    // Should not have broken audio element
    expect(html).not.toContain('<audio');
  });

  it('wraps players in semantic HTML structure', () => {
    const html = buildReportPlayerHTML({
      date: '2026-07-28',
      audioUrl: 'https://example.com/audio.mp3',
      videoUrl: 'https://example.com/video.mp4',
    });

    // Check for semantic HTML
    expect(html).toMatch(/<section|<article|<div class="media"/i);
  });
});
