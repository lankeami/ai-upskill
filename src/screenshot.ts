import { chromium, Browser, Page } from '@playwright/test';
import { existsSync } from 'fs';

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

    const { statSync } = require('fs');
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
