# AI Daily Report — Design Spec

## Overview

A Go CLI tool (`ai-report`) that generates a daily Markdown report of AI news by aggregating content from RSS feeds, Reddit, and Hacker News. Items are categorized by company/product using rule-based keyword matching.

## Architecture

Single Go binary with this flow:

```
CLI invocation → Fetch sources (parallel HTTP) → Raw items → Rule-based filter/categorize → Deduplicate → Markdown renderer → reports/YYYY-MM-DD.md
```

### Project Structure

```
ai-upskill/
├── cmd/
│   └── ai-report/
│       └── main.go              # CLI entry point (Cobra)
├── internal/
│   ├── fetcher/
│   │   ├── fetcher.go           # Source fetching orchestrator (parallel via errgroup)
│   │   ├── reddit.go            # Reddit JSON API (append .json to subreddit URLs)
│   │   ├── hackernews.go        # HN Firebase API
│   │   └── rss.go               # Generic RSS feed parser
│   ├── processor/
│   │   └── classifier.go        # Rule-based categorization, filtering, dedup
│   └── renderer/
│       └── markdown.go          # Markdown file generation
├── reports/                      # Generated daily reports (committed to repo)
├── config.yaml                   # Source URLs, keywords, company mappings
├── .github/
│   └── workflows/
│       └── daily-report.yml      # GitHub Actions cron schedule
├── go.mod
└── go.sum
```

## Data Sources (v1)

| Source | Method | What It Captures |
|--------|--------|-----------------|
| Reddit (`r/artificial`, `r/MachineLearning`, `r/LocalLLaMA`, `r/ChatGPT`) | JSON API (no auth — append `.json` to subreddit URL) | Community discussions, product launches, open source releases |
| Hacker News | Firebase API (`https://hacker-news.firebaseio.com`) | Tech-focused AI news, research papers, launches |
| TechCrunch AI | RSS feed | Funding, acquisitions, product announcements |
| The Verge AI | RSS feed | Consumer AI product news |
| Ars Technica AI | RSS feed | In-depth AI analysis and news |

Only items from the last 24 hours are included.

### Future Data Sources (Post-v1)

Direct platform scraping to be explored after v1 is stable:

- **Twitter/X**: Official API ($100/mo basic tier) or community proxies
- **Instagram**: Unofficial APIs or headless browser scraping
- **TikTok**: Unofficial APIs or headless browser scraping

These were deferred because the aggregator approach captures the same news through secondary sources without the cost and fragility of direct platform scraping.

## Processing & Categorization (Rule-Based)

### 1. AI Relevance Filter

Case-insensitive keyword matching against a configurable list in `config.yaml`. Items whose title or description don't match any keyword are discarded.

Default keywords: `AI`, `artificial intelligence`, `LLM`, `GPT`, `machine learning`, `neural network`, `foundation model`, `deep learning`, `transformer`, `diffusion model`, `generative AI`, `NLP`, `computer vision`, `reinforcement learning`.

### 2. Company Categorization

Case-insensitive keyword/regex matching assigns each item to a company:

| Company | Keywords |
|---------|----------|
| OpenAI | `openai`, `gpt`, `chatgpt`, `dall-e`, `sora`, `o1`, `o3` |
| Google | `google`, `gemini`, `deepmind`, `bard` |
| Anthropic | `anthropic`, `claude` |
| Meta | `meta ai`, `llama`, `meta's` |
| Microsoft | `microsoft`, `copilot`, `azure ai`, `phi-` |
| Mistral | `mistral`, `mixtral` |
| Apple | `apple intelligence`, `apple ai`, `apple ml` |
| Stability AI | `stability ai`, `stable diffusion`, `stablelm` |
| xAI | `xai`, `grok` |

Items matching no company → "Other/Independent". All mappings are configurable in `config.yaml`.

### 3. Deduplication

- URL-based dedup (exact match)
- Fuzzy title matching (Levenshtein distance threshold) to merge the same story from different sources

When items are merged, all source URLs are preserved.

### 4. Sorting

Within each company section, items sorted by number of sources covering the story (most-covered first).

### Future Enhancement

LLM-powered categorization and summarization (via Claude API) to replace rule-based processing when budget allows. The processor is a separate package so swapping implementations is straightforward.

## Report Output Format

Generated file: `reports/YYYY-MM-DD.md`

```markdown
# AI Daily Report — 2026-04-15

## OpenAI
- **GPT-5 Turbo announced with 2x context window** — [TechCrunch](https://...) | [Reddit](https://...)
- **New ChatGPT plugin marketplace launches** — [The Verge](https://...)

## Google
- **Gemini 3.0 benchmarks leaked** — [Hacker News](https://...) | [Ars Technica](https://...)

## Anthropic
- **Claude adds tool use in production** — [Reddit](https://...)

## Other/Independent
- **Open-source model "Falcon 3" released by TII** — [Hacker News](https://...)

---
*Generated at 2026-04-15T08:00:00Z | Sources: Reddit, Hacker News, TechCrunch, The Verge, Ars Technica*
```

- Empty company sections are omitted
- Each item shows title and source links (multiple if covered by several sources)
- Footer with generation timestamp and sources used

## CLI Interface

Built with Cobra.

```bash
# Generate today's report
ai-report generate

# Generate report for a specific date
ai-report generate --date 2026-04-14

# List available sources
ai-report sources

# Dry run — fetch and show what would be in the report without writing file
ai-report generate --dry-run
```

## Scheduling (GitHub Actions)

`.github/workflows/daily-report.yml`:

- Cron trigger: daily at 08:00 UTC
- Steps: checkout repo → run `ai-report generate` → commit report → push
- Manual trigger via `workflow_dispatch`

## Configuration

`config.yaml`:

```yaml
sources:
  reddit:
    subreddits:
      - artificial
      - MachineLearning
      - LocalLLaMA
      - ChatGPT
  hackernews:
    min_score: 10
  rss:
    - name: TechCrunch AI
      url: https://techcrunch.com/category/artificial-intelligence/feed/
    - name: The Verge AI
      url: https://www.theverge.com/rss/ai-artificial-intelligence/index.xml
    - name: Ars Technica AI
      url: https://feeds.arstechnica.com/arstechnica/technology-lab

keywords:
  ai_relevance:
    - AI
    - artificial intelligence
    - LLM
    - GPT
    - machine learning
    - neural network
    - foundation model
    - deep learning
    - transformer
    - diffusion model
    - generative AI
    - NLP
    - computer vision
    - reinforcement learning

companies:
  OpenAI: ["openai", "gpt", "chatgpt", "dall-e", "sora", "o1", "o3"]
  Google: ["google", "gemini", "deepmind", "bard"]
  Anthropic: ["anthropic", "claude"]
  Meta: ["meta ai", "llama", "meta's"]
  Microsoft: ["microsoft", "copilot", "azure ai", "phi-"]
  Mistral: ["mistral", "mixtral"]
  Apple: ["apple intelligence", "apple ai", "apple ml"]
  Stability AI: ["stability ai", "stable diffusion", "stablelm"]
  xAI: ["xai", "grok"]

output:
  dir: reports
  format: markdown

dedup:
  levenshtein_threshold: 0.85
```

## Dependencies

- **cobra** — CLI framework
- **viper** — config file parsing (YAML)
- **gofeed** — RSS/Atom feed parser
- **errgroup** — parallel fetching (stdlib `golang.org/x/sync/errgroup`)
- **agnivade/levenshtein** — fuzzy title matching for dedup
