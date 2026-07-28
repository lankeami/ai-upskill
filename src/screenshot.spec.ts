import { describe, it, expect, beforeAll, afterAll } from 'vitest';
import { existsSync, unlinkSync, writeFileSync } from 'fs';
import { captureReportScreenshot } from './screenshot';
import { fileURLToPath } from 'url';
import { dirname, join } from 'path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const TEST_OUTPUT_DIR = '/tmp/test-screenshots';
const TEST_HTML_FILE = '/tmp/test-report.html';
const EXPECTED_WIDTH = 1920;
const EXPECTED_HEIGHT = 1080;

// Create a simple HTML file for testing
const createTestHTML = () => {
  const html = `<!DOCTYPE html>
<html>
<head>
  <title>Test Report</title>
</head>
<body>
  <h1>Test Report - 2026-07-28</h1>
  <div class="content">
    <h2>Section 1</h2>
    <p>This is test content for the report.</p>
  </div>
</body>
</html>`;
  writeFileSync(TEST_HTML_FILE, html);
};

describe('Playwright screenshot capture', () => {
  beforeAll(() => {
    if (!existsSync(TEST_OUTPUT_DIR)) {
      require('fs').mkdirSync(TEST_OUTPUT_DIR, { recursive: true });
    }
    createTestHTML();
  });

  afterAll(() => {
    // Cleanup test files
    if (existsSync(TEST_HTML_FILE)) {
      unlinkSync(TEST_HTML_FILE);
    }
    const screenshotDir = TEST_OUTPUT_DIR;
    if (existsSync(screenshotDir)) {
      const files = require('fs').readdirSync(screenshotDir);
      files.forEach((file: string) => {
        unlinkSync(`${screenshotDir}/${file}`);
      });
    }
  });

  it('captures report page screenshot with correct dimensions', async () => {
    // Use local file:// URL
    const reportUrl = `file://${TEST_HTML_FILE}`;
    const outputPath = `${TEST_OUTPUT_DIR}/report-screenshot.png`;

    const result = await captureReportScreenshot({
      url: reportUrl,
      outputPath,
      width: EXPECTED_WIDTH,
      height: EXPECTED_HEIGHT,
      timeout: 30000,
    });

    // Verify screenshot file exists
    expect(existsSync(outputPath)).toBe(true);

    // Verify result metadata
    expect(result).toBeDefined();
    expect(result.outputPath).toBe(outputPath);
    expect(result.width).toBe(EXPECTED_WIDTH);
    expect(result.height).toBe(EXPECTED_HEIGHT);
    expect(result.fileSize).toBeGreaterThan(0);
  });

  it('captures headline area of report page', async () => {
    const reportUrl = `file://${TEST_HTML_FILE}`;
    const outputPath = `${TEST_OUTPUT_DIR}/headline-screenshot.png`;

    const result = await captureReportScreenshot({
      url: reportUrl,
      outputPath,
      width: 1280,
      height: 400, // Just headline area
      timeout: 30000,
      selector: 'h1', // Capture just the headline
    });

    expect(existsSync(outputPath)).toBe(true);
    expect(result.width).toBe(1280);
    expect(result.height).toBe(400);
    expect(result.fileSize).toBeGreaterThan(0);
  });

  it('returns consistent dimensions across multiple captures', async () => {
    const reportUrl = `file://${TEST_HTML_FILE}`;
    const outputPath1 = `${TEST_OUTPUT_DIR}/screenshot1.png`;
    const outputPath2 = `${TEST_OUTPUT_DIR}/screenshot2.png`;

    const result1 = await captureReportScreenshot({
      url: reportUrl,
      outputPath: outputPath1,
      width: 1920,
      height: 1080,
      timeout: 30000,
    });

    const result2 = await captureReportScreenshot({
      url: reportUrl,
      outputPath: outputPath2,
      width: 1920,
      height: 1080,
      timeout: 30000,
    });

    expect(result1.width).toBe(result2.width);
    expect(result1.height).toBe(result2.height);
    expect(existsSync(outputPath1)).toBe(true);
    expect(existsSync(outputPath2)).toBe(true);
  });
});
