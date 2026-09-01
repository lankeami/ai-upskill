import { chromium, Browser, Page } from '@playwright/test';
import { existsSync, statSync } from 'fs';

export interface ScreenshotOptions {
  url: string;
  outputPath: string;
  width: number;
  height: number;
  timeout?: number;
  selector?: string; // Optional: capture only this element
}

export interface ScreenshotResult {
  outputPath: string;
  width: number;
  height: number;
  fileSize: number;
}

export interface SourceCapture {
  url: string;
  outputPath: string;
}

export interface SourceCaptureResult {
  url: string;
  outputPath: string;
  ok: boolean;
  error?: string;
}

/**
 * Capture screenshots of many source pages, reusing a single browser.
 * Failures are recorded per-URL rather than aborting the batch —
 * news sites time out, block bots, or 404, and one bad page
 * must not sink the whole video.
 */
export async function captureSourceScreenshots(
  captures: SourceCapture[],
  opts: { width: number; height: number; timeout?: number }
): Promise<SourceCaptureResult[]> {
  const { width, height, timeout = 20000 } = opts;
  const results: SourceCaptureResult[] = [];

  const browser = await chromium.launch({ headless: true });
  try {
    for (const { url, outputPath } of captures) {
      const page = await browser.newPage({ viewport: { width, height } });
      try {
        page.setDefaultTimeout(timeout);
        page.setDefaultNavigationTimeout(timeout);
        // 'domcontentloaded' not 'networkidle': ad-heavy news sites never go idle
        await page.goto(url, { waitUntil: 'domcontentloaded' });
        await page.waitForTimeout(1500); // let fonts/images settle
        await page.screenshot({ path: outputPath, fullPage: false });
        results.push({ url, outputPath, ok: true });
      } catch (e) {
        results.push({
          url,
          outputPath,
          ok: false,
          error: e instanceof Error ? e.message.split('\n')[0] : String(e),
        });
      } finally {
        await page.close();
      }
    }
  } finally {
    await browser.close();
  }

  return results;
}

/**
 * Capture a screenshot of a report page using Playwright.
 * Sets viewport to specified dimensions and waits for page load.
 */
export async function captureReportScreenshot(
  options: ScreenshotOptions
): Promise<ScreenshotResult> {
  const {
    url,
    outputPath,
    width,
    height,
    timeout = 30000,
    selector,
  } = options;

  let browser: Browser | null = null;
  let page: Page | null = null;

  try {
    browser = await chromium.launch({
      headless: true,
    });

    page = await browser.newPage({
      viewport: { width, height },
    });

    page.setDefaultTimeout(timeout);
    page.setDefaultNavigationTimeout(timeout);

    // Navigate to the report page
    // For file:// URLs, use 'domcontentloaded' instead of 'networkidle'
    const waitUntil = url.startsWith('file://') ? 'domcontentloaded' : 'networkidle';
    await page.goto(url, { waitUntil });

    // Wait for content to load
    await page.waitForLoadState('domcontentloaded');

    // Optionally wait for specific element
    if (selector) {
      await page.waitForSelector(selector, { timeout });
    }

    // Take screenshot
    if (selector) {
      const element = await page.locator(selector).first();
      await element.screenshot({ path: outputPath });
    } else {
      await page.screenshot({ path: outputPath, fullPage: false });
    }

    // Verify file exists and get size
    if (!existsSync(outputPath)) {
      throw new Error(`Screenshot was not saved to ${outputPath}`);
    }

    const stats = statSync(outputPath);
    const fileSize = stats.size;

    return {
      outputPath,
      width,
      height,
      fileSize,
    };
  } finally {
    if (page) {
      await page.close();
    }
    if (browser) {
      await browser.close();
    }
  }
}
