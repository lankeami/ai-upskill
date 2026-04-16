# GitHub Pages Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Turn the ai-upskill repo into a Jekyll-based GitHub Pages site serving an index page and styled daily report pages.

**Architecture:** Jekyll at repo root, deployed from `main` branch. The Go renderer is updated to emit Jekyll front matter. New `_layouts/` and `_config.yml` provide site structure. A CI test script ensures Jekyll never processes unintended files.

**Tech Stack:** Jekyll (GitHub Pages native), Liquid templates, HTML/CSS, Go (renderer update), Bash (test script)

**Spec:** `docs/superpowers/specs/2026-04-15-github-pages-design.md`

---

### Task 1: Create Jekyll configuration

**Files:**
- Create: `_config.yml`

- [ ] **Step 1: Create `_config.yml`**

```yaml
title: AI Daily Report
description: Daily AI news aggregated from Reddit, Hacker News, and tech RSS feeds.
baseurl: ""
url: ""

exclude:
  - cmd/
  - internal/
  - docs/
  - go.mod
  - go.sum
  - config.yaml
  - "*.go"
  - ai-report
  - scripts/
  - Makefile
  - README.md
  - CLAUDE.md
  - .claude/
  - ai-report/
  - Gemfile.lock
```

- [ ] **Step 2: Commit**

```bash
git add _config.yml
git commit -m "feat(jekyll): add Jekyll configuration with exclude list"
```

---

### Task 2: Create default layout

**Files:**
- Create: `_layouts/default.html`

- [ ] **Step 1: Create `_layouts/default.html`**

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{ page.title | default: site.title }}</title>
  <style>
    * { margin: 0; padding: 0; box-sizing: border-box; }
    body {
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif;
      line-height: 1.6;
      color: #24292f;
      max-width: 800px;
      margin: 0 auto;
      padding: 2rem 1rem;
    }
    a { color: #0969da; text-decoration: none; }
    a:hover { text-decoration: underline; }
    h1 { margin-bottom: 0.5rem; font-size: 1.75rem; }
    h2 { margin-top: 1.5rem; margin-bottom: 0.5rem; font-size: 1.25rem; }
    ul { padding-left: 1.5rem; }
    li { margin-bottom: 0.25rem; }
    .site-header { margin-bottom: 2rem; padding-bottom: 1rem; border-bottom: 1px solid #d1d9e0; }
    .site-header a { color: #24292f; font-weight: 600; }
    .site-footer { margin-top: 3rem; padding-top: 1rem; border-top: 1px solid #d1d9e0; font-size: 0.85rem; color: #656d76; }
    hr { border: none; border-top: 1px solid #d1d9e0; margin: 1.5rem 0; }
  </style>
</head>
<body>
  <header class="site-header">
    <a href="{{ '/' | relative_url }}">{{ site.title }}</a>
  </header>
  <main>
    {{ content }}
  </main>
  <footer class="site-footer">
    <p>Generated daily from Reddit, Hacker News, and RSS feeds.</p>
  </footer>
</body>
</html>
```

- [ ] **Step 2: Commit**

```bash
git add _layouts/default.html
git commit -m "feat(jekyll): add default HTML layout"
```

---

### Task 3: Create report layout with collapsible sections, TOC, and navigation

**Files:**
- Create: `_layouts/report.html`

This is the most complex layout. It extends `default.html` and uses Liquid + JavaScript to:
1. Render a human-readable date header
2. Generate a table of contents from `companies` front matter
3. Wrap each `## Company` section in a collapsible `<details>` element
4. Add previous/next report navigation

- [ ] **Step 1: Create `_layouts/report.html`**

```html
---
layout: default
---

<style>
  .report-nav { display: flex; justify-content: space-between; margin-bottom: 1.5rem; font-size: 0.9rem; }
  .report-nav a { color: #0969da; }
  .report-nav .disabled { color: #656d76; }
  .report-date { font-size: 1.1rem; color: #656d76; margin-bottom: 1rem; }
  .report-toc { margin-bottom: 1.5rem; padding: 1rem; background: #f6f8fa; border-radius: 6px; }
  .report-toc h3 { font-size: 0.9rem; margin-bottom: 0.5rem; color: #656d76; text-transform: uppercase; letter-spacing: 0.05em; }
  .report-toc ul { list-style: none; padding: 0; display: flex; flex-wrap: wrap; gap: 0.5rem; }
  .report-toc li a { font-size: 0.85rem; padding: 0.2rem 0.5rem; background: #fff; border: 1px solid #d1d9e0; border-radius: 4px; display: inline-block; }
  .report-toc li a:hover { background: #0969da; color: #fff; text-decoration: none; }
  details { margin-bottom: 1rem; }
  details summary { cursor: pointer; font-size: 1.25rem; font-weight: 600; padding: 0.5rem 0; }
  details summary:hover { color: #0969da; }
  details[open] summary { margin-bottom: 0.5rem; }
</style>

{% comment %} Build sorted list of report pages for prev/next nav {% endcomment %}
{% assign report_pages = site.pages | where_exp: "p", "p.path contains 'reports/'" | where_exp: "p", "p.date" | sort: "date" %}

{% assign prev_report = nil %}
{% assign next_report = nil %}
{% for p in report_pages %}
  {% if p.url == page.url %}
    {% if forloop.index0 > 0 %}
      {% assign prev_idx = forloop.index0 | minus: 1 %}
      {% assign prev_report = report_pages[prev_idx] %}
    {% endif %}
    {% assign next_idx = forloop.index0 | plus: 1 %}
    {% if next_idx < report_pages.size %}
      {% assign next_report = report_pages[next_idx] %}
    {% endif %}
  {% endif %}
{% endfor %}

<nav class="report-nav">
  {% if prev_report %}
    <a href="{{ prev_report.url | relative_url }}">&larr; {{ prev_report.date | date: "%B %d, %Y" }}</a>
  {% else %}
    <span class="disabled">&larr; No earlier report</span>
  {% endif %}
  <a href="{{ '/' | relative_url }}">All Reports</a>
  {% if next_report %}
    <a href="{{ next_report.url | relative_url }}">{{ next_report.date | date: "%B %d, %Y" }} &rarr;</a>
  {% else %}
    <span class="disabled">No later report &rarr;</span>
  {% endif %}
</nav>

<h1>{{ page.title }}</h1>
<p class="report-date">{{ page.date | date: "%A, %B %d, %Y" }}</p>

{% if page.companies %}
<nav class="report-toc">
  <h3>Companies</h3>
  <ul>
    {% for company in page.companies %}
      <li><a href="#{{ company | slugify }}">{{ company }}</a></li>
    {% endfor %}
  </ul>
</nav>
{% endif %}

<div id="report-content">
{{ content }}
</div>

<nav class="report-nav" style="margin-top: 2rem;">
  {% if prev_report %}
    <a href="{{ prev_report.url | relative_url }}">&larr; {{ prev_report.date | date: "%B %d, %Y" }}</a>
  {% else %}
    <span class="disabled">&larr; No earlier report</span>
  {% endif %}
  <a href="{{ '/' | relative_url }}">All Reports</a>
  {% if next_report %}
    <a href="{{ next_report.url | relative_url }}">{{ next_report.date | date: "%B %d, %Y" }} &rarr;</a>
  {% else %}
    <span class="disabled">No later report &rarr;</span>
  {% endif %}
</nav>

<script>
// Wrap each h2 + following content in a collapsible <details> element
document.addEventListener('DOMContentLoaded', function() {
  var content = document.getElementById('report-content');
  var children = Array.from(content.children);
  var details = null;

  children.forEach(function(el) {
    if (el.tagName === 'H2') {
      // Close previous details if open
      if (details) {
        content.insertBefore(details, el);
      }
      details = document.createElement('details');
      details.setAttribute('open', '');
      var summary = document.createElement('summary');
      summary.id = el.id || el.textContent.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/(^-|-$)/g, '');
      summary.textContent = el.textContent;
      details.appendChild(summary);
      el.remove();
    } else if (details) {
      details.appendChild(el);
    }
  });
  // Append last details
  if (details) {
    content.appendChild(details);
  }
});
</script>
```

- [ ] **Step 2: Commit**

```bash
git add _layouts/report.html
git commit -m "feat(jekyll): add report layout with TOC, collapsible sections, and nav"
```

---

### Task 4: Create index page

**Files:**
- Create: `index.md`

- [ ] **Step 1: Create `index.md`**

```markdown
---
layout: default
title: AI Daily Report
---

# AI Daily Report

A daily aggregation of AI news from Reddit, Hacker News, and tech RSS feeds, organized by company. Reports are generated automatically each day.

## Reports

{% assign report_pages = site.pages | where_exp: "p", "p.path contains 'reports/'" | where_exp: "p", "p.date" | sort: "date" | reverse %}

{% for report in report_pages %}
{% assign company_count = report.companies | size %}
{% if company_count > 3 %}
  {% assign shown = report.companies | slice: 0, 3 | join: ", " %}
  {% assign remaining = company_count | minus: 3 %}
- [{{ report.date | date: "%B %d, %Y" }}]({{ report.url | relative_url }}) — {{ shown }} + {{ remaining }} more — {{ report.item_count }} items
{% else %}
- [{{ report.date | date: "%B %d, %Y" }}]({{ report.url | relative_url }}) — {{ report.companies | join: ", " }} — {{ report.item_count }} items
{% endif %}
{% endfor %}
```

- [ ] **Step 2: Commit**

```bash
git add index.md
git commit -m "feat(jekyll): add index page with report listing and blurbs"
```

---

### Task 5: Update Go renderer to emit Jekyll front matter

**Files:**
- Modify: `internal/renderer/markdown.go:17-48`
- Modify: `internal/renderer/markdown_test.go`

The `RenderMarkdown` function must prepend Jekyll front matter to each report with `layout`, `title`, `date`, `companies` list, and `item_count`.

- [ ] **Step 1: Write failing test for front matter output**

Add this test to `internal/renderer/markdown_test.go`:

```go
func TestRenderMarkdownIncludesFrontMatter(t *testing.T) {
	date := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)

	classified := map[string][]processor.DeduplicatedItem{
		"OpenAI": {
			{
				Title:   "GPT-5 Released",
				Sources: []processor.SourceRef{{Name: "Reddit", URL: "https://reddit.com/gpt5"}},
			},
		},
		"Google": {
			{
				Title:   "Gemini Update",
				Sources: []processor.SourceRef{{Name: "HN", URL: "https://hn.com/gemini"}},
			},
			{
				Title:   "Bard News",
				Sources: []processor.SourceRef{{Name: "RSS", URL: "https://rss.com/bard"}},
			},
		},
	}

	result := RenderMarkdown(classified, date, []string{"Reddit"})

	if !strings.HasPrefix(result, "---\n") {
		t.Fatal("report must start with front matter delimiter")
	}
	if !strings.Contains(result, "layout: report") {
		t.Error("missing layout field in front matter")
	}
	if !strings.Contains(result, "date: 2026-04-15") {
		t.Error("missing date field in front matter")
	}
	if !strings.Contains(result, "item_count: 3") {
		t.Error("missing or incorrect item_count in front matter")
	}
	if !strings.Contains(result, "\"OpenAI\"") {
		t.Error("missing OpenAI in companies list")
	}
	if !strings.Contains(result, "\"Google\"") {
		t.Error("missing Google in companies list")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/jaychinthrajah/workspaces/_personal_/ai-upskill && go test ./internal/renderer/ -run TestRenderMarkdownIncludesFrontMatter -v`
Expected: FAIL — the current renderer does not emit front matter.

- [ ] **Step 3: Update `RenderMarkdown` to emit front matter**

Replace the beginning of the `RenderMarkdown` function in `internal/renderer/markdown.go`:

```go
func RenderMarkdown(classified map[string][]processor.DeduplicatedItem, date time.Time, sources []string) string {
	var b strings.Builder

	// Collect companies and total item count for front matter
	companies := make([]string, 0)
	itemCount := 0
	rendered := make(map[string]bool)

	for _, company := range companyOrder {
		items, ok := classified[company]
		if !ok || len(items) == 0 {
			continue
		}
		companies = append(companies, company)
		itemCount += len(items)
	}
	for company, items := range classified {
		if !rendered[company] {
			found := false
			for _, c := range companies {
				if c == company {
					found = true
					break
				}
			}
			if !found {
				companies = append(companies, company)
				itemCount += len(items)
			}
		}
	}

	// Write Jekyll front matter
	b.WriteString("---\n")
	b.WriteString("layout: report\n")
	b.WriteString(fmt.Sprintf("title: \"AI Daily Report — %s\"\n", date.Format("2006-01-02")))
	b.WriteString(fmt.Sprintf("date: %s\n", date.Format("2006-01-02")))
	companiesJSON := make([]string, len(companies))
	for i, c := range companies {
		companiesJSON[i] = fmt.Sprintf("\"%s\"", c)
	}
	b.WriteString(fmt.Sprintf("companies: [%s]\n", strings.Join(companiesJSON, ", ")))
	b.WriteString(fmt.Sprintf("item_count: %d\n", itemCount))
	b.WriteString("---\n\n")

	b.WriteString(fmt.Sprintf("# AI Daily Report — %s\n\n", date.Format("2006-01-02")))

	rendered = make(map[string]bool)
	for _, company := range companyOrder {
		items, ok := classified[company]
		if !ok || len(items) == 0 {
			continue
		}
		rendered[company] = true
		renderSection(&b, company, items)
	}

	remaining := make([]string, 0)
	for company := range classified {
		if !rendered[company] {
			remaining = append(remaining, company)
		}
	}
	sort.Strings(remaining)
	for _, company := range remaining {
		renderSection(&b, company, classified[company])
	}

	b.WriteString("---\n")
	b.WriteString(fmt.Sprintf("*Generated at %s | Sources: %s*\n",
		time.Now().UTC().Format(time.RFC3339),
		strings.Join(sources, ", ")))

	return b.String()
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/jaychinthrajah/workspaces/_personal_/ai-upskill && go test ./internal/renderer/ -v`
Expected: All tests PASS (including the existing ones — they use `strings.Contains` so front matter prefix won't break them).

- [ ] **Step 5: Commit**

```bash
git add internal/renderer/markdown.go internal/renderer/markdown_test.go
git commit -m "feat(renderer): emit Jekyll front matter in Markdown reports"
```

---

### Task 6: Add front matter to existing report

**Files:**
- Modify: `reports/2026-04-15.md` (prepend front matter)

The existing report was generated before the Go renderer emitted front matter. We need to manually add front matter to it.

- [ ] **Step 1: Count companies and items in the existing report**

Run: `cd /Users/jaychinthrajah/workspaces/_personal_/ai-upskill && echo "Companies:" && grep '^## ' reports/2026-04-15.md && echo "Item count:" && grep '^- \*\*' reports/2026-04-15.md | wc -l`

Use the output to construct the front matter.

- [ ] **Step 2: Prepend front matter to `reports/2026-04-15.md`**

Add this to the very top of the file (before the `# AI Daily Report` heading):

```yaml
---
layout: report
title: "AI Daily Report — 2026-04-15"
date: 2026-04-15
companies: ["Other/Independent", "anthropic", "google", "meta", "microsoft", "mistral", "openai", "xai"]
item_count: <actual count from step 1>
---

```

Note: Replace `<actual count from step 1>` with the real number. The companies list should match the `## ` headings found in step 1, preserving their exact casing.

- [ ] **Step 3: Commit**

```bash
git add reports/2026-04-15.md
git commit -m "feat(reports): add Jekyll front matter to existing report"
```

---

### Task 7: Create Jekyll build safety test

**Files:**
- Create: `scripts/test-jekyll-output.sh`

- [ ] **Step 1: Create `scripts/test-jekyll-output.sh`**

```bash
#!/usr/bin/env bash
set -euo pipefail

echo "Building Jekyll site..."
bundle exec jekyll build --destination _site 2>&1

echo ""
echo "Checking _site output..."

# Collect all HTML files in _site
html_files=$(find _site -name '*.html' -type f | sort)

echo "HTML files found:"
echo "$html_files"
echo ""

# Check each HTML file is either index.html or under reports/
fail=0
while IFS= read -r file; do
  # Strip _site/ prefix
  rel="${file#_site/}"
  case "$rel" in
    index.html) ;;
    reports/*.html) ;;
    *)
      echo "UNEXPECTED FILE: $rel"
      fail=1
      ;;
  esac
done <<< "$html_files"

if [ "$fail" -eq 1 ]; then
  echo ""
  echo "FAIL: Jekyll built unexpected files. Update _config.yml exclude list."
  exit 1
fi

echo ""
echo "PASS: Jekyll output contains only index.html and reports/*.html"
```

- [ ] **Step 2: Make the script executable**

Run: `chmod +x scripts/test-jekyll-output.sh`

- [ ] **Step 3: Commit**

```bash
git add scripts/test-jekyll-output.sh
git commit -m "feat(ci): add Jekyll build output safety test"
```

---

### Task 8: Add Gemfile for Jekyll

**Files:**
- Create: `Gemfile`

GitHub Pages needs a Gemfile to ensure the right Jekyll version is used.

- [ ] **Step 1: Create `Gemfile`**

```ruby
source "https://rubygems.org"

gem "github-pages", group: :jekyll_plugins
```

- [ ] **Step 2: Add `Gemfile` and `Gemfile.lock` to `_config.yml` exclude list if not already present**

Verify `Gemfile.lock` is in the exclude list in `_config.yml`. (It was included in Task 1.)

- [ ] **Step 3: Commit**

```bash
git add Gemfile
git commit -m "feat(jekyll): add Gemfile for GitHub Pages"
```

---

### Task 9: Update CLAUDE.md with Jekyll exclusion rule

**Files:**
- Modify: `.claude/CLAUDE.md`

- [ ] **Step 1: Add Jekyll exclusion rule to CLAUDE.md**

Append this section:

```markdown
## Jekyll / GitHub Pages

**Jekyll Exclusion Rule:** Jekyll must only serve `index.md` and `reports/*.md`. All other directories and top-level Markdown files must be listed in `_config.yml`'s `exclude` list. When adding new top-level directories or Markdown files to the repo, add them to the exclude list and verify with `scripts/test-jekyll-output.sh`.
```

- [ ] **Step 2: Commit**

```bash
git add .claude/CLAUDE.md
git commit -m "docs: add Jekyll exclusion rule to CLAUDE.md"
```

---

### Task 10: Update GitHub Actions workflow

**Files:**
- Modify: `.github/workflows/daily-report.yml`

The workflow no longer needs a post-processing step since the Go renderer now emits front matter. But we should add the Jekyll safety test.

- [ ] **Step 1: Update `.github/workflows/daily-report.yml`**

Replace the entire file content with:

```yaml
name: Daily AI Report

on:
  schedule:
    - cron: '0 8 * * *'  # Daily at 08:00 UTC
  workflow_dispatch:       # Manual trigger

permissions:
  contents: write

jobs:
  generate-report:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.26'

      - name: Build
        run: go build -o ai-report ./cmd/ai-report

      - name: Generate report
        run: ./ai-report generate

      - name: Set up Ruby
        uses: ruby/setup-ruby@v1
        with:
          ruby-version: '3.2'
          bundler-cache: true

      - name: Verify Jekyll output
        run: ./scripts/test-jekyll-output.sh

      - name: Commit and push report
        run: |
          git config user.name "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"
          git add reports/
          if git diff --cached --quiet; then
            echo "No new report to commit"
          else
            git commit -m "chore: daily AI report for $(date -u +%Y-%m-%d)"
            git push
          fi
```

- [ ] **Step 2: Commit**

```bash
git add .github/workflows/daily-report.yml
git commit -m "feat(ci): add Jekyll safety test to daily report workflow"
```

---

### Task 11: Local verification

- [ ] **Step 1: Install Jekyll locally and run the build test**

Run:
```bash
cd /Users/jaychinthrajah/workspaces/_personal_/ai-upskill
bundle install
./scripts/test-jekyll-output.sh
```

Expected: "PASS: Jekyll output contains only index.html and reports/*.html"

- [ ] **Step 2: Preview the site locally**

Run: `bundle exec jekyll serve`

Then open `http://localhost:4000` in a browser. Verify:
1. Index page shows with the project description and report list
2. The report blurb shows companies and item count
3. Clicking a report opens the styled report page
4. Report has: date header, TOC, collapsible sections, prev/next nav (disabled for single report)
5. No other pages are served (e.g., no docs/ content)

- [ ] **Step 3: Run Go tests to confirm nothing is broken**

Run: `cd /Users/jaychinthrajah/workspaces/_personal_/ai-upskill && go test ./...`
Expected: All tests PASS.
