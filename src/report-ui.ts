export interface PlayerOptions {
  date: string;
  audioUrl: string | null;
  videoUrl: string | null;
  videoWidth?: number;
  videoHeight?: number;
}

/**
 * Build HTML for audio and video players to embed in a report page.
 * Returns semantic HTML with accessibility features.
 */
export function buildReportPlayerHTML(options: PlayerOptions): string {
  const { date, audioUrl, videoUrl, videoWidth = 1920, videoHeight = 1080 } = options;

  let html = '<section class="media-players" aria-label="Report media">\n';

  // Video player (if available)
  if (videoUrl) {
    html += `  <div class="video-player">
    <h3>Watch Report</h3>
    <video width="${videoWidth}" height="${videoHeight}" controls poster="">
      <source src="${escapeHTML(videoUrl)}" type="video/mp4">
      Your browser does not support the video tag.
      <a href="${escapeHTML(videoUrl)}">Download video</a>
    </video>
  </div>\n`;
  }

  // Audio player (if available)
  if (audioUrl) {
    html += `  <div class="audio-player">
    <h3>Listen to Podcast</h3>
    <audio controls>
      <source src="${escapeHTML(audioUrl)}" type="audio/mpeg">
      Your browser does not support the audio element.
      <a href="${escapeHTML(audioUrl)}">Download audio</a>
    </audio>
  </div>\n`;
  }

  html += '</section>\n';

  return html;
}

/**
 * Escape HTML special characters to prevent XSS.
 */
function escapeHTML(str: string): string {
  const map: { [key: string]: string } = {
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#039;',
  };
  return str.replace(/[&<>"']/g, (char) => map[char]);
}

/**
 * Build complete player CSS for styling.
 */
export function buildPlayerCSS(): string {
  return `
.media-players {
  display: flex;
  flex-direction: column;
  gap: 2rem;
  margin: 2rem 0;
  padding: 1rem;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  background: #f9f9f9;
}

.video-player,
.audio-player {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.video-player h3,
.audio-player h3 {
  margin: 0;
  font-size: 1.1rem;
  color: #333;
}

video,
audio {
  width: 100%;
  max-width: 100%;
  border-radius: 4px;
}

@media (min-width: 768px) {
  .media-players {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
    gap: 2rem;
  }
}
`;
}

/**
 * Inject player HTML into Jekyll front matter.
 * Updates a markdown report file with player URLs.
 */
export function injectPlayersIntoReport(
  reportContent: string,
  options: { audioUrl?: string; videoUrl?: string }
): string {
  // Extract front matter
  const fmMatch = reportContent.match(/^---\n([\s\S]*?)\n---\n([\s\S]*)$/);
  if (!fmMatch) {
    return reportContent;
  }

  const [, frontMatter, body] = fmMatch;
  let updatedFM = frontMatter;

  if (options.audioUrl) {
    updatedFM = updatedFM.replace(/audio_url:.*$/m, `audio_url: "${options.audioUrl}"`);
    if (!updatedFM.includes('audio_url:')) {
      updatedFM += `\naudio_url: "${options.audioUrl}"`;
    }
  }

  if (options.videoUrl) {
    updatedFM = updatedFM.replace(/video_url:.*$/m, `video_url: "${options.videoUrl}"`);
    if (!updatedFM.includes('video_url:')) {
      updatedFM += `\nvideo_url: "${options.videoUrl}"`;
    }
  }

  return `---\n${updatedFM}\n---\n${body}`;
}
