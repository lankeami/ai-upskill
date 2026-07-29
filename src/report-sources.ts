export interface ReportSource {
  company: string;
  title: string;
  url: string;
}

/**
 * Extract article sources from a daily report markdown, in document order.
 * Articles are top-level bullets of the form:
 *   - **Title** — [Source Name](https://url)
 */
export function extractReportSources(markdown: string): ReportSource[] {
  const sources: ReportSource[] = [];
  let currentCompany = '';

  const headingRe = /^##\s+(.+)$/;
  const articleRe = /^-\s+\*\*(.+?)\*\*\s+—\s+\[[^\]]+\]\((https?:\/\/[^)]+)\)/;

  for (const line of markdown.split('\n')) {
    const heading = line.match(headingRe);
    if (heading) {
      currentCompany = heading[1].trim();
      continue;
    }

    const article = line.match(articleRe);
    if (article) {
      sources.push({
        company: currentCompany,
        title: article[1].trim(),
        url: article[2],
      });
    }
  }

  return sources;
}
