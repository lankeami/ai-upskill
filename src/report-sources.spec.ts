import { describe, it, expect } from 'vitest';
import { extractReportSources } from './report-sources';

const SAMPLE_REPORT = `---
layout: report
title: "AI Daily Report — 2026-07-28"
date: 2026-07-28
---

# AI Daily Report — 2026-07-28

## anthropic
- **Anthropic responds to open-weight debate** — [TechCrunch AI](https://techcrunch.com/2026/07/27/anthropic-responds/)
  Anthropic founder made his views clear.
- **Discovering Cryptographic Weaknesses with Claude** — [Hacker News](https://www.anthropic.com/research/crypto-weaknesses)
  Researchers find weaknesses.

## google
- **Google AI search becoming default** — [TechCrunch AI](https://techcrunch.com/2026/07/27/google-ai-search/)
  AI Overviews now appear in 43% of searches.

## Other/Independent
- **Cyera acquires Oasis Security** — [TechCrunch AI](https://techcrunch.com/2026/07/28/cyera-oasis/)
  Third acquisition this year.
`;

describe('extractReportSources', () => {
  it('extracts all article sources in document order', () => {
    const sources = extractReportSources(SAMPLE_REPORT);

    expect(sources).toHaveLength(4);
    expect(sources[0]).toEqual({
      company: 'anthropic',
      title: 'Anthropic responds to open-weight debate',
      url: 'https://techcrunch.com/2026/07/27/anthropic-responds/',
    });
    expect(sources[1].url).toBe('https://www.anthropic.com/research/crypto-weaknesses');
    expect(sources[2].company).toBe('google');
    expect(sources[3].company).toBe('Other/Independent');
  });

  it('returns empty array for report with no articles', () => {
    const sources = extractReportSources('# Empty Report\n\nNothing here.');
    expect(sources).toEqual([]);
  });

  it('ignores non-article links in body text', () => {
    const report = `## openai
- **Real article** — [Hacker News](https://example.com/article)
  Summary with an inline [link](https://example.com/inline) that should not count.
`;
    const sources = extractReportSources(report);
    expect(sources).toHaveLength(1);
    expect(sources[0].url).toBe('https://example.com/article');
  });
});
