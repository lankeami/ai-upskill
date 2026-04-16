# AI Daily Report

Daily AI news aggregated from Reddit, Hacker News, and tech RSS feeds — published automatically to GitHub Pages.

## Architecture

```
Fetch (Reddit, HN, RSS) --> Filter (AI relevance) --> Categorize (by company)
    --> Deduplicate (URL + fuzzy title) --> Render (Markdown) --> GitHub Pages
```

### Tech Stack

- **Language:** Go 1.26
- **CLI Framework:** Cobra + Viper
- **Fetching:** `gofeed` (RSS), `net/http` (Reddit JSON API, HN Firebase API)
- **Deduplication:** URL exact match + Levenshtein distance (0.85 threshold)
- **Publishing:** Jekyll on GitHub Pages (main branch)
- **CI/CD:** GitHub Actions (daily cron at 08:00 UTC)

### Project Structure

```
cmd/ai-report/main.go        # CLI entry point (generate, sources commands)
internal/
  config/config.go            # YAML config loading
  fetcher/                    # Reddit, HN, RSS fetchers (parallel)
  processor/                  # AI relevance filter + company categorizer + dedup
  renderer/markdown.go        # Markdown report generator
_layouts/                     # Jekyll layouts (default, report)
reports/                      # Generated daily reports
scripts/test-jekyll-output.sh # CI safety test
config.yaml                   # Sources, keywords, companies config
_config.yml                   # Jekyll config
.github/workflows/            # CI/CD
```

### Key Design Decisions

| Decision | Choice | Why |
|----------|--------|-----|
| Data sources | Reddit JSON, HN Firebase, RSS | No auth required, broad AI news coverage |
| Categorization | Rule-based keyword matching | Simple, fast, configurable in YAML |
| Deduplication | URL + Levenshtein (0.85) | Catches cross-source same-story coverage |
| Publishing | Jekyll on main branch | Zero-setup GitHub Pages, Markdown-native |
| Jekyll safety | CI build test + exclude list | Prevents Jekyll from processing Go source or docs |

## Updating Companies

Companies and their matching keywords are defined in `config.yaml` under the `companies` section:

```yaml
companies:
  - name: OpenAI
    keywords:
      - openai
      - gpt
      - chatgpt
      - dall-e
      - sora
  - name: Anthropic
    keywords:
      - anthropic
      - claude
  # ...
```

**To add a new company:**

1. Add an entry to the `companies` list in `config.yaml` with a `name` and `keywords` array.
2. Keywords are case-insensitive substring matches against item titles and text.
3. Items that don't match any company land in **Other/Independent**.

**To modify an existing company:**

Edit the `keywords` array for that company in `config.yaml`. No code changes needed.

## Adding or Changing News Sources

Sources are also in `config.yaml`:

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
    feeds:
      - https://techcrunch.com/category/artificial-intelligence/feed/
      - https://www.theverge.com/rss/ai-artificial-intelligence/index.xml
      - https://feeds.arstechnica.com/arstechnica/features
```

Add subreddits or RSS feed URLs directly. Adjust `min_score` to control HN noise.

## Building and Running Locally

### Prerequisites

- Go 1.26+
- Ruby 3.2+ and Bundler (for Jekyll preview)

### Generate a Report

```bash
# Build the CLI
go build -o ai-report ./cmd/ai-report

# Generate today's report
./ai-report generate

# Generate for a specific date
./ai-report generate --date 2026-04-15

# Dry run (print to stdout, don't write file)
./ai-report generate --dry-run

# List configured sources
./ai-report sources
```

Reports are written to `reports/YYYY-MM-DD.md`.

### Preview with Jekyll (GitHub Pages)

```bash
# Install Ruby dependencies
bundle install

# Serve locally
bundle exec jekyll serve

# Open http://localhost:4000
```

### Run Tests

```bash
go test ./...
```

### Jekyll Safety Test

```bash
scripts/test-jekyll-output.sh
```

This verifies Jekyll only builds `index.html` and `reports/*.html` — nothing else. It runs in CI on every report generation.

## GitHub Pages Deployment

The site deploys automatically from the `main` branch:

1. **Daily at 08:00 UTC**, GitHub Actions runs the workflow (`.github/workflows/daily-report.yml`):
   - Builds the Go binary
   - Generates the day's report
   - Runs the Jekyll safety test
   - Commits and pushes if a new report was created
2. **GitHub Pages** picks up the push and rebuilds the Jekyll site.
3. You can also trigger the workflow manually via `workflow_dispatch`.

### Jekyll Exclusion Rule

Jekyll must only serve `index.md` and `reports/*.md`. All other directories and top-level Markdown files are listed in `_config.yml`'s `exclude` list. When adding new top-level directories or Markdown files, add them to the exclude list and verify with `scripts/test-jekyll-output.sh`.
