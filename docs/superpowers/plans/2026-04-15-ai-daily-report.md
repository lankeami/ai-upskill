# AI Daily Report Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Go CLI tool that generates daily Markdown reports of AI news by aggregating RSS feeds, Reddit, and Hacker News, categorized by company/product.

**Architecture:** Single Go binary using Cobra for CLI. Fetcher package pulls from sources in parallel via errgroup. Processor package does rule-based keyword filtering, company categorization, and deduplication. Renderer package writes Markdown files to `reports/`.

**Tech Stack:** Go 1.26, Cobra (CLI), Viper (config), gofeed (RSS), errgroup (parallel fetching), agnivade/levenshtein (dedup)

---

## File Structure

```
ai-upskill/
├── cmd/
│   └── ai-report/
│       └── main.go                  # CLI entry point, Cobra root + generate + sources commands
├── internal/
│   ├── config/
│   │   └── config.go                # Config loading via Viper
│   ├── fetcher/
│   │   ├── types.go                 # RawItem struct, Source interface
│   │   ├── reddit.go                # Reddit JSON API fetcher
│   │   ├── reddit_test.go
│   │   ├── hackernews.go            # HN API fetcher
│   │   ├── hackernews_test.go
│   │   ├── rss.go                   # Generic RSS fetcher
│   │   ├── rss_test.go
│   │   ├── fetcher.go               # Orchestrator: runs all sources in parallel
│   │   └── fetcher_test.go
│   ├── processor/
│   │   ├── classifier.go            # AI relevance filter + company categorization
│   │   ├── classifier_test.go
│   │   ├── dedup.go                 # URL + fuzzy title deduplication
│   │   └── dedup_test.go
│   └── renderer/
│       ├── markdown.go              # Markdown report generation
│       └── markdown_test.go
├── reports/                          # Generated reports (gitkeep)
├── config.yaml                       # Default configuration
├── .github/
│   └── workflows/
│       └── daily-report.yml          # GitHub Actions cron
├── go.mod
└── go.sum
```

---

### Task 1: Project Initialization

**Files:**
- Create: `go.mod`
- Create: `cmd/ai-report/main.go`
- Create: `config.yaml`
- Create: `reports/.gitkeep`

- [ ] **Step 1: Initialize Go module**

Run:
```bash
cd /Users/jaychinthrajah/workspaces/_personal_/ai-upskill
go mod init github.com/jaychinthrajah/ai-upskill
```

Expected: `go.mod` created with module path.

- [ ] **Step 2: Install dependencies**

Run:
```bash
go get github.com/spf13/cobra@latest
go get github.com/spf13/viper@latest
go get github.com/mmcdole/gofeed@latest
go get golang.org/x/sync@latest
go get github.com/agnivade/levenshtein@latest
```

Expected: Dependencies added to `go.mod` and `go.sum` created.

- [ ] **Step 3: Create config.yaml**

Create `config.yaml`:

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
  Meta: ["meta ai", "llama"]
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

- [ ] **Step 4: Create minimal main.go with Cobra root command**

Create `cmd/ai-report/main.go`:

```go
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ai-report",
	Short: "Generate daily AI news reports",
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate the daily AI news report",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("generate not yet implemented")
		return nil
	},
}

var sourcesCmd = &cobra.Command{
	Use:   "sources",
	Short: "List configured news sources",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("sources not yet implemented")
		return nil
	},
}

func init() {
	generateCmd.Flags().String("date", "", "Generate report for a specific date (YYYY-MM-DD)")
	generateCmd.Flags().Bool("dry-run", false, "Show report without writing to file")
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(sourcesCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

- [ ] **Step 5: Create reports directory with gitkeep**

Run:
```bash
mkdir -p reports && touch reports/.gitkeep
```

- [ ] **Step 6: Verify it compiles and runs**

Run:
```bash
go build -o ai-report ./cmd/ai-report && ./ai-report --help
```

Expected: Help output showing `generate` and `sources` subcommands.

- [ ] **Step 7: Commit**

```bash
git init
git add go.mod go.sum cmd/ config.yaml reports/.gitkeep
git commit -m "feat: initialize project with Cobra CLI skeleton and config"
```

---

### Task 2: Config Loading

**Files:**
- Create: `internal/config/config.go`

- [ ] **Step 1: Write config.go**

Create `internal/config/config.go`:

```go
package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type RSSSource struct {
	Name string `mapstructure:"name"`
	URL  string `mapstructure:"url"`
}

type RedditConfig struct {
	Subreddits []string `mapstructure:"subreddits"`
}

type HackerNewsConfig struct {
	MinScore int `mapstructure:"min_score"`
}

type SourcesConfig struct {
	Reddit     RedditConfig     `mapstructure:"reddit"`
	HackerNews HackerNewsConfig `mapstructure:"hackernews"`
	RSS        []RSSSource      `mapstructure:"rss"`
}

type DedupConfig struct {
	LevenshteinThreshold float64 `mapstructure:"levenshtein_threshold"`
}

type OutputConfig struct {
	Dir    string `mapstructure:"dir"`
	Format string `mapstructure:"format"`
}

type Config struct {
	Sources   SourcesConfig        `mapstructure:"sources"`
	Keywords  map[string][]string  `mapstructure:"keywords"`
	Companies map[string][]string  `mapstructure:"companies"`
	Dedup     DedupConfig          `mapstructure:"dedup"`
	Output    OutputConfig         `mapstructure:"output"`
}

func Load(path string) (*Config, error) {
	viper.SetConfigFile(path)
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshalling config: %w", err)
	}

	return &cfg, nil
}
```

- [ ] **Step 2: Verify it compiles**

Run:
```bash
go build ./internal/config/
```

Expected: No errors.

- [ ] **Step 3: Commit**

```bash
git add internal/config/
git commit -m "feat: add config loading with Viper"
```

---

### Task 3: Fetcher Types and RSS Fetcher

**Files:**
- Create: `internal/fetcher/types.go`
- Create: `internal/fetcher/rss.go`
- Create: `internal/fetcher/rss_test.go`

- [ ] **Step 1: Write types.go with shared RawItem struct**

Create `internal/fetcher/types.go`:

```go
package fetcher

import "time"

// RawItem represents a single news item from any source.
type RawItem struct {
	Title       string
	URL         string
	Description string
	Source      string    // e.g., "Reddit", "Hacker News", "TechCrunch AI"
	PublishedAt time.Time
}
```

- [ ] **Step 2: Write failing test for RSS fetcher**

Create `internal/fetcher/rss_test.go`:

```go
package fetcher

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const testRSSFeed = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Test Feed</title>
    <item>
      <title>New AI Model Released</title>
      <link>https://example.com/ai-model</link>
      <description>A new AI model was released today.</description>
      <pubDate>%s</pubDate>
    </item>
    <item>
      <title>Old News</title>
      <link>https://example.com/old</link>
      <description>This is old news.</description>
      <pubDate>Mon, 01 Jan 2024 00:00:00 GMT</pubDate>
    </item>
  </channel>
</rss>`

func TestFetchRSS(t *testing.T) {
	now := time.Now().UTC()
	feed := fmt.Sprintf(testRSSFeed, now.Format(time.RFC1123Z))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte(feed))
	}))
	defer server.Close()

	since := now.Add(-24 * time.Hour)
	items, err := FetchRSS(server.URL, "Test Feed", since)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	if items[0].Title != "New AI Model Released" {
		t.Errorf("expected title 'New AI Model Released', got %q", items[0].Title)
	}
	if items[0].Source != "Test Feed" {
		t.Errorf("expected source 'Test Feed', got %q", items[0].Source)
	}
	if items[0].URL != "https://example.com/ai-model" {
		t.Errorf("expected URL 'https://example.com/ai-model', got %q", items[0].URL)
	}
}
```

Note: add `"fmt"` to the imports in the test file.

- [ ] **Step 3: Run test to verify it fails**

Run:
```bash
go test ./internal/fetcher/ -run TestFetchRSS -v
```

Expected: FAIL — `FetchRSS` not defined.

- [ ] **Step 4: Implement FetchRSS**

Create `internal/fetcher/rss.go`:

```go
package fetcher

import (
	"fmt"
	"time"

	"github.com/mmcdole/gofeed"
)

func FetchRSS(url, sourceName string, since time.Time) ([]RawItem, error) {
	parser := gofeed.NewParser()
	feed, err := parser.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("parsing RSS feed %s: %w", sourceName, err)
	}

	var items []RawItem
	for _, entry := range feed.Items {
		published := time.Time{}
		if entry.PublishedParsed != nil {
			published = *entry.PublishedParsed
		} else if entry.UpdatedParsed != nil {
			published = *entry.UpdatedParsed
		}

		if !published.IsZero() && published.Before(since) {
			continue
		}

		items = append(items, RawItem{
			Title:       entry.Title,
			URL:         entry.Link,
			Description: entry.Description,
			Source:       sourceName,
			PublishedAt: published,
		})
	}

	return items, nil
}
```

- [ ] **Step 5: Run test to verify it passes**

Run:
```bash
go test ./internal/fetcher/ -run TestFetchRSS -v
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/fetcher/types.go internal/fetcher/rss.go internal/fetcher/rss_test.go
git commit -m "feat: add RSS fetcher with time filtering"
```

---

### Task 4: Reddit Fetcher

**Files:**
- Create: `internal/fetcher/reddit.go`
- Create: `internal/fetcher/reddit_test.go`

- [ ] **Step 1: Write failing test for Reddit fetcher**

Create `internal/fetcher/reddit_test.go`:

```go
package fetcher

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetchReddit(t *testing.T) {
	now := time.Now().UTC()
	recentTimestamp := float64(now.Unix())
	oldTimestamp := float64(now.Add(-48 * time.Hour).Unix())

	response := map[string]interface{}{
		"data": map[string]interface{}{
			"children": []map[string]interface{}{
				{
					"data": map[string]interface{}{
						"title":      "GPT-5 is here",
						"url":        "https://example.com/gpt5",
						"selftext":   "OpenAI just released GPT-5",
						"created_utc": recentTimestamp,
						"permalink":  "/r/artificial/comments/abc123/gpt5_is_here/",
					},
				},
				{
					"data": map[string]interface{}{
						"title":      "Old post",
						"url":        "https://example.com/old",
						"selftext":   "This is old",
						"created_utc": oldTimestamp,
						"permalink":  "/r/artificial/comments/def456/old_post/",
					},
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	since := now.Add(-24 * time.Hour)
	items, err := FetchReddit("artificial", since, server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	if items[0].Title != "GPT-5 is here" {
		t.Errorf("expected title 'GPT-5 is here', got %q", items[0].Title)
	}
	if items[0].Source != "Reddit r/artificial" {
		t.Errorf("expected source 'Reddit r/artificial', got %q", items[0].Source)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:
```bash
go test ./internal/fetcher/ -run TestFetchReddit -v
```

Expected: FAIL — `FetchReddit` not defined.

- [ ] **Step 3: Implement FetchReddit**

Create `internal/fetcher/reddit.go`:

```go
package fetcher

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type redditResponse struct {
	Data struct {
		Children []struct {
			Data struct {
				Title      string  `json:"title"`
				URL        string  `json:"url"`
				Selftext   string  `json:"selftext"`
				CreatedUTC float64 `json:"created_utc"`
				Permalink  string  `json:"permalink"`
			} `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

// FetchReddit fetches recent posts from a subreddit.
// baseURL is optional — pass "" to use the real Reddit API.
func FetchReddit(subreddit string, since time.Time, baseURL string) ([]RawItem, error) {
	url := fmt.Sprintf("https://www.reddit.com/r/%s/new.json?limit=50", subreddit)
	if baseURL != "" {
		url = fmt.Sprintf("%s/r/%s/new.json?limit=50", baseURL, subreddit)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request for r/%s: %w", subreddit, err)
	}
	req.Header.Set("User-Agent", "ai-report/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching r/%s: %w", subreddit, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("r/%s returned status %d", subreddit, resp.StatusCode)
	}

	var result redditResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding r/%s response: %w", subreddit, err)
	}

	var items []RawItem
	for _, child := range result.Data.Children {
		published := time.Unix(int64(child.Data.CreatedUTC), 0).UTC()
		if published.Before(since) {
			continue
		}

		postURL := child.Data.URL
		if child.Data.URL == "" || child.Data.URL == fmt.Sprintf("https://www.reddit.com%s", child.Data.Permalink) {
			postURL = fmt.Sprintf("https://www.reddit.com%s", child.Data.Permalink)
		}

		items = append(items, RawItem{
			Title:       child.Data.Title,
			URL:         postURL,
			Description: child.Data.Selftext,
			Source:      fmt.Sprintf("Reddit r/%s", subreddit),
			PublishedAt: published,
		})
	}

	return items, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run:
```bash
go test ./internal/fetcher/ -run TestFetchReddit -v
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/fetcher/reddit.go internal/fetcher/reddit_test.go
git commit -m "feat: add Reddit JSON API fetcher"
```

---

### Task 5: Hacker News Fetcher

**Files:**
- Create: `internal/fetcher/hackernews.go`
- Create: `internal/fetcher/hackernews_test.go`

- [ ] **Step 1: Write failing test for HN fetcher**

Create `internal/fetcher/hackernews_test.go`:

```go
package fetcher

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetchHackerNews(t *testing.T) {
	now := time.Now().UTC()

	mux := http.NewServeMux()

	// Top stories endpoint
	mux.HandleFunc("/v0/topstories.json", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]int{1, 2})
	})

	// Story 1: recent, high score
	mux.HandleFunc("/v0/item/1.json", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":    1,
			"title": "Claude 4 Released",
			"url":   "https://example.com/claude4",
			"score": 150,
			"time":  now.Unix(),
			"type":  "story",
		})
	})

	// Story 2: old, should be filtered
	mux.HandleFunc("/v0/item/2.json", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":    2,
			"title": "Old Story",
			"url":   "https://example.com/old",
			"score": 200,
			"time":  now.Add(-48 * time.Hour).Unix(),
			"type":  "story",
		})
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	since := now.Add(-24 * time.Hour)
	items, err := FetchHackerNews(10, since, server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	if items[0].Title != "Claude 4 Released" {
		t.Errorf("expected title 'Claude 4 Released', got %q", items[0].Title)
	}
	if items[0].Source != "Hacker News" {
		t.Errorf("expected source 'Hacker News', got %q", items[0].Source)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:
```bash
go test ./internal/fetcher/ -run TestFetchHackerNews -v
```

Expected: FAIL — `FetchHackerNews` not defined.

- [ ] **Step 3: Implement FetchHackerNews**

Create `internal/fetcher/hackernews.go`:

```go
package fetcher

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type hnItem struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
	Score int    `json:"score"`
	Time  int64  `json:"time"`
	Type  string `json:"type"`
}

// FetchHackerNews fetches top stories from HN that meet the score threshold.
// baseURL is optional — pass "" to use the real HN API.
func FetchHackerNews(minScore int, since time.Time, baseURL string) ([]RawItem, error) {
	apiBase := "https://hacker-news.firebaseio.com"
	if baseURL != "" {
		apiBase = baseURL
	}

	// Fetch top story IDs
	resp, err := http.Get(fmt.Sprintf("%s/v0/topstories.json", apiBase))
	if err != nil {
		return nil, fmt.Errorf("fetching HN top stories: %w", err)
	}
	defer resp.Body.Close()

	var storyIDs []int
	if err := json.NewDecoder(resp.Body).Decode(&storyIDs); err != nil {
		return nil, fmt.Errorf("decoding HN top stories: %w", err)
	}

	// Limit to top 100 to avoid excessive requests
	if len(storyIDs) > 100 {
		storyIDs = storyIDs[:100]
	}

	// Fetch stories in parallel
	var mu sync.Mutex
	var items []RawItem
	var wg sync.WaitGroup

	// Use a semaphore to limit concurrent requests
	sem := make(chan struct{}, 10)

	for _, id := range storyIDs {
		wg.Add(1)
		go func(storyID int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			storyResp, err := http.Get(fmt.Sprintf("%s/v0/item/%d.json", apiBase, storyID))
			if err != nil {
				return
			}
			defer storyResp.Body.Close()

			var story hnItem
			if err := json.NewDecoder(storyResp.Body).Decode(&story); err != nil {
				return
			}

			if story.Type != "story" {
				return
			}

			published := time.Unix(story.Time, 0).UTC()
			if published.Before(since) {
				return
			}

			if story.Score < minScore {
				return
			}

			url := story.URL
			if url == "" {
				url = fmt.Sprintf("https://news.ycombinator.com/item?id=%d", story.ID)
			}

			mu.Lock()
			items = append(items, RawItem{
				Title:       story.Title,
				URL:         url,
				Description: "",
				Source:      "Hacker News",
				PublishedAt: published,
			})
			mu.Unlock()
		}(id)
	}

	wg.Wait()
	return items, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run:
```bash
go test ./internal/fetcher/ -run TestFetchHackerNews -v
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/fetcher/hackernews.go internal/fetcher/hackernews_test.go
git commit -m "feat: add Hacker News API fetcher"
```

---

### Task 6: Fetcher Orchestrator

**Files:**
- Create: `internal/fetcher/fetcher.go`
- Create: `internal/fetcher/fetcher_test.go`

- [ ] **Step 1: Write failing test for FetchAll**

Create `internal/fetcher/fetcher_test.go`:

```go
package fetcher

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jaychinthrajah/ai-upskill/internal/config"
)

func TestFetchAll(t *testing.T) {
	now := time.Now().UTC()

	// Mock Reddit server
	redditMux := http.NewServeMux()
	redditMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"children": []map[string]interface{}{
					{
						"data": map[string]interface{}{
							"title":       "Reddit AI Post",
							"url":         "https://example.com/reddit-ai",
							"selftext":    "AI stuff",
							"created_utc": float64(now.Unix()),
							"permalink":   "/r/artificial/comments/abc/test/",
						},
					},
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	})
	redditServer := httptest.NewServer(redditMux)
	defer redditServer.Close()

	// Mock HN server
	hnMux := http.NewServeMux()
	hnMux.HandleFunc("/v0/topstories.json", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]int{1})
	})
	hnMux.HandleFunc("/v0/item/1.json", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":    1,
			"title": "HN AI Post",
			"url":   "https://example.com/hn-ai",
			"score": 50,
			"time":  now.Unix(),
			"type":  "story",
		})
	})
	hnServer := httptest.NewServer(hnMux)
	defer hnServer.Close()

	// Mock RSS server
	rssFeed := fmt.Sprintf(`<?xml version="1.0"?>
<rss version="2.0">
  <channel>
    <item>
      <title>RSS AI Post</title>
      <link>https://example.com/rss-ai</link>
      <description>AI from RSS</description>
      <pubDate>%s</pubDate>
    </item>
  </channel>
</rss>`, now.Format(time.RFC1123Z))
	rssServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(rssFeed))
	}))
	defer rssServer.Close()

	cfg := &config.SourcesConfig{
		Reddit:     config.RedditConfig{Subreddits: []string{"artificial"}},
		HackerNews: config.HackerNewsConfig{MinScore: 10},
		RSS:        []config.RSSSource{{Name: "Test RSS", URL: rssServer.URL}},
	}

	items, errs := FetchAll(cfg, now.Add(-24*time.Hour), redditServer.URL, hnServer.URL)

	if len(errs) > 0 {
		t.Logf("fetch errors: %v", errs)
	}

	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:
```bash
go test ./internal/fetcher/ -run TestFetchAll -v
```

Expected: FAIL — `FetchAll` not defined.

- [ ] **Step 3: Implement FetchAll**

Create `internal/fetcher/fetcher.go`:

```go
package fetcher

import (
	"sync"
	"time"

	"github.com/jaychinthrajah/ai-upskill/internal/config"
)

// FetchAll fetches from all configured sources in parallel.
// redditBaseURL and hnBaseURL are optional overrides for testing — pass "" for production.
func FetchAll(cfg *config.SourcesConfig, since time.Time, redditBaseURL, hnBaseURL string) ([]RawItem, []error) {
	var mu sync.Mutex
	var allItems []RawItem
	var allErrors []error
	var wg sync.WaitGroup

	// Reddit
	for _, sub := range cfg.Reddit.Subreddits {
		wg.Add(1)
		go func(subreddit string) {
			defer wg.Done()
			items, err := FetchReddit(subreddit, since, redditBaseURL)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				allErrors = append(allErrors, err)
				return
			}
			allItems = append(allItems, items...)
		}(sub)
	}

	// Hacker News
	wg.Add(1)
	go func() {
		defer wg.Done()
		items, err := FetchHackerNews(cfg.HackerNews.MinScore, since, hnBaseURL)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			allErrors = append(allErrors, err)
			return
		}
		allItems = append(allItems, items...)
	}()

	// RSS feeds
	for _, feed := range cfg.RSS {
		wg.Add(1)
		go func(f config.RSSSource) {
			defer wg.Done()
			items, err := FetchRSS(f.URL, f.Name, since)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				allErrors = append(allErrors, err)
				return
			}
			allItems = append(allItems, items...)
		}(feed)
	}

	wg.Wait()
	return allItems, allErrors
}
```

- [ ] **Step 4: Run test to verify it passes**

Run:
```bash
go test ./internal/fetcher/ -run TestFetchAll -v
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/fetcher/fetcher.go internal/fetcher/fetcher_test.go
git commit -m "feat: add parallel fetcher orchestrator"
```

---

### Task 7: Classifier (AI Filter + Company Categorization)

**Files:**
- Create: `internal/processor/classifier.go`
- Create: `internal/processor/classifier_test.go`

- [ ] **Step 1: Write failing tests**

Create `internal/processor/classifier_test.go`:

```go
package processor

import (
	"testing"
	"time"

	"github.com/jaychinthrajah/ai-upskill/internal/fetcher"
)

func TestIsAIRelevant(t *testing.T) {
	keywords := []string{"AI", "LLM", "GPT", "machine learning"}

	tests := []struct {
		title    string
		desc     string
		expected bool
	}{
		{"New GPT-5 model released", "", true},
		{"Machine learning breakthrough", "", true},
		{"Best pizza in NYC", "", false},
		{"No keywords here", "But the description mentions AI research", true},
		{"OpenAI announces new LLM", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			item := fetcher.RawItem{Title: tt.title, Description: tt.desc}
			got := IsAIRelevant(item, keywords)
			if got != tt.expected {
				t.Errorf("IsAIRelevant(%q, %q) = %v, want %v", tt.title, tt.desc, got, tt.expected)
			}
		})
	}
}

func TestCategorizeByCompany(t *testing.T) {
	companies := map[string][]string{
		"OpenAI":    {"openai", "gpt", "chatgpt"},
		"Google":    {"google", "gemini", "deepmind"},
		"Anthropic": {"anthropic", "claude"},
	}

	tests := []struct {
		title    string
		expected string
	}{
		{"OpenAI releases GPT-5", "OpenAI"},
		{"Google's Gemini gets upgrade", "Google"},
		{"Claude improves at coding", "Anthropic"},
		{"New open-source model from TII", "Other/Independent"},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			item := fetcher.RawItem{Title: tt.title}
			got := CategorizeByCompany(item, companies)
			if got != tt.expected {
				t.Errorf("CategorizeByCompany(%q) = %q, want %q", tt.title, got, tt.expected)
			}
		})
	}
}

func TestClassify(t *testing.T) {
	keywords := []string{"AI", "LLM", "GPT"}
	companies := map[string][]string{
		"OpenAI": {"openai", "gpt"},
		"Google": {"google", "gemini"},
	}

	items := []fetcher.RawItem{
		{Title: "GPT-5 released by OpenAI", URL: "https://example.com/gpt5", Source: "Reddit", PublishedAt: time.Now()},
		{Title: "Best pizza in NYC", URL: "https://example.com/pizza", Source: "Reddit", PublishedAt: time.Now()},
		{Title: "Google Gemini update", URL: "https://example.com/gemini", Source: "HN", PublishedAt: time.Now()},
	}

	result := Classify(items, keywords, companies)

	if len(result) != 2 {
		t.Fatalf("expected 2 companies, got %d", len(result))
	}

	openaiItems, ok := result["OpenAI"]
	if !ok || len(openaiItems) != 1 {
		t.Errorf("expected 1 OpenAI item, got %d", len(openaiItems))
	}

	googleItems, ok := result["Google"]
	if !ok || len(googleItems) != 1 {
		t.Errorf("expected 1 Google item, got %d", len(googleItems))
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run:
```bash
go test ./internal/processor/ -v
```

Expected: FAIL — functions not defined.

- [ ] **Step 3: Implement classifier.go**

Create `internal/processor/classifier.go`:

```go
package processor

import (
	"strings"

	"github.com/jaychinthrajah/ai-upskill/internal/fetcher"
)

// ClassifiedItem wraps a RawItem with its assigned company.
type ClassifiedItem struct {
	fetcher.RawItem
	Company string
}

// IsAIRelevant checks if an item matches any AI-related keyword (case-insensitive).
func IsAIRelevant(item fetcher.RawItem, keywords []string) bool {
	text := strings.ToLower(item.Title + " " + item.Description)
	for _, kw := range keywords {
		if strings.Contains(text, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}

// CategorizeByCompany returns the company name for an item based on keyword matching.
// Returns "Other/Independent" if no company matches.
func CategorizeByCompany(item fetcher.RawItem, companies map[string][]string) string {
	text := strings.ToLower(item.Title + " " + item.Description)
	for company, keywords := range companies {
		for _, kw := range keywords {
			if strings.Contains(text, strings.ToLower(kw)) {
				return company
			}
		}
	}
	return "Other/Independent"
}

// Classify filters items for AI relevance and groups them by company.
// Returns a map of company name → items.
func Classify(items []fetcher.RawItem, aiKeywords []string, companies map[string][]string) map[string][]fetcher.RawItem {
	result := make(map[string][]fetcher.RawItem)

	for _, item := range items {
		if !IsAIRelevant(item, aiKeywords) {
			continue
		}
		company := CategorizeByCompany(item, companies)
		result[company] = append(result[company], item)
	}

	return result
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run:
```bash
go test ./internal/processor/ -v
```

Expected: PASS (all 3 tests).

- [ ] **Step 5: Commit**

```bash
git add internal/processor/classifier.go internal/processor/classifier_test.go
git commit -m "feat: add rule-based AI classifier and company categorizer"
```

---

### Task 8: Deduplication

**Files:**
- Create: `internal/processor/dedup.go`
- Create: `internal/processor/dedup_test.go`

- [ ] **Step 1: Write failing tests**

Create `internal/processor/dedup_test.go`:

```go
package processor

import (
	"testing"
	"time"

	"github.com/jaychinthrajah/ai-upskill/internal/fetcher"
)

func TestDeduplicate(t *testing.T) {
	now := time.Now()

	items := []fetcher.RawItem{
		{Title: "GPT-5 Released by OpenAI", URL: "https://example.com/gpt5", Source: "Reddit", PublishedAt: now},
		{Title: "GPT-5 Released by OpenAI", URL: "https://example.com/gpt5", Source: "HN", PublishedAt: now},
		{Title: "GPT-5 released by OpenAI today", URL: "https://other.com/gpt5", Source: "TechCrunch", PublishedAt: now},
		{Title: "Completely different story", URL: "https://example.com/other", Source: "Reddit", PublishedAt: now},
	}

	result := Deduplicate(items, 0.85)

	// Should merge the first three (same URL or similar title), keep the last one separate
	if len(result) != 2 {
		t.Fatalf("expected 2 deduplicated items, got %d", len(result))
	}
}

func TestDeduplicatePreservesSources(t *testing.T) {
	now := time.Now()

	items := []fetcher.RawItem{
		{Title: "GPT-5 Released", URL: "https://example.com/gpt5", Source: "Reddit", PublishedAt: now},
		{Title: "GPT-5 Released", URL: "https://example.com/gpt5", Source: "Hacker News", PublishedAt: now},
	}

	result := Deduplicate(items, 0.85)

	if len(result) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result))
	}

	if len(result[0].Sources) != 2 {
		t.Errorf("expected 2 sources, got %d", len(result[0].Sources))
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run:
```bash
go test ./internal/processor/ -run TestDedup -v
```

Expected: FAIL — `Deduplicate` not defined.

- [ ] **Step 3: Implement dedup.go**

Create `internal/processor/dedup.go`:

```go
package processor

import (
	"strings"

	"github.com/agnivade/levenshtein"
	"github.com/jaychinthrajah/ai-upskill/internal/fetcher"
)

// DeduplicatedItem represents a merged item with all its sources.
type DeduplicatedItem struct {
	Title       string
	URL         string
	Description string
	Sources     []SourceRef
}

// SourceRef is a source name + URL pair.
type SourceRef struct {
	Name string
	URL  string
}

// titleSimilarity returns a similarity score between 0 and 1 for two titles.
func titleSimilarity(a, b string) float64 {
	a = strings.ToLower(strings.TrimSpace(a))
	b = strings.ToLower(strings.TrimSpace(b))

	if a == b {
		return 1.0
	}

	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}
	if maxLen == 0 {
		return 1.0
	}

	dist := levenshtein.ComputeDistance(a, b)
	return 1.0 - float64(dist)/float64(maxLen)
}

// Deduplicate merges items that share the same URL or have very similar titles.
func Deduplicate(items []fetcher.RawItem, threshold float64) []DeduplicatedItem {
	var result []DeduplicatedItem

	for _, item := range items {
		merged := false
		for i, existing := range result {
			if item.URL == existing.URL || titleSimilarity(item.Title, existing.Title) >= threshold {
				result[i].Sources = append(result[i].Sources, SourceRef{
					Name: item.Source,
					URL:  item.URL,
				})
				merged = true
				break
			}
		}

		if !merged {
			result = append(result, DeduplicatedItem{
				Title:       item.Title,
				URL:         item.URL,
				Description: item.Description,
				Sources: []SourceRef{
					{Name: item.Source, URL: item.URL},
				},
			})
		}
	}

	return result
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run:
```bash
go test ./internal/processor/ -run TestDedup -v
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/processor/dedup.go internal/processor/dedup_test.go
git commit -m "feat: add URL and fuzzy title deduplication"
```

---

### Task 9: Markdown Renderer

**Files:**
- Create: `internal/renderer/markdown.go`
- Create: `internal/renderer/markdown_test.go`

- [ ] **Step 1: Write failing test**

Create `internal/renderer/markdown_test.go`:

```go
package renderer

import (
	"strings"
	"testing"
	"time"

	"github.com/jaychinthrajah/ai-upskill/internal/processor"
)

func TestRenderMarkdown(t *testing.T) {
	date := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)

	classified := map[string][]processor.DeduplicatedItem{
		"OpenAI": {
			{
				Title: "GPT-5 Released",
				Sources: []processor.SourceRef{
					{Name: "Reddit", URL: "https://reddit.com/gpt5"},
					{Name: "TechCrunch", URL: "https://techcrunch.com/gpt5"},
				},
			},
		},
		"Google": {
			{
				Title: "Gemini Update",
				Sources: []processor.SourceRef{
					{Name: "Hacker News", URL: "https://hn.com/gemini"},
				},
			},
		},
	}

	sources := []string{"Reddit", "Hacker News", "TechCrunch"}
	result := RenderMarkdown(classified, date, sources)

	if !strings.Contains(result, "# AI Daily Report — 2026-04-15") {
		t.Error("missing report header")
	}
	if !strings.Contains(result, "## OpenAI") {
		t.Error("missing OpenAI section")
	}
	if !strings.Contains(result, "## Google") {
		t.Error("missing Google section")
	}
	if !strings.Contains(result, "**GPT-5 Released**") {
		t.Error("missing GPT-5 item")
	}
	if !strings.Contains(result, "[Reddit]") {
		t.Error("missing Reddit source link")
	}
	if !strings.Contains(result, "[TechCrunch]") {
		t.Error("missing TechCrunch source link")
	}
}

func TestRenderMarkdownOmitsEmptySections(t *testing.T) {
	date := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)

	classified := map[string][]processor.DeduplicatedItem{
		"OpenAI": {
			{
				Title:   "GPT-5",
				Sources: []processor.SourceRef{{Name: "Reddit", URL: "https://reddit.com/gpt5"}},
			},
		},
	}

	result := RenderMarkdown(classified, date, []string{"Reddit"})

	if strings.Contains(result, "## Google") {
		t.Error("should not contain empty Google section")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run:
```bash
go test ./internal/renderer/ -v
```

Expected: FAIL — `RenderMarkdown` not defined.

- [ ] **Step 3: Implement markdown.go**

Create `internal/renderer/markdown.go`:

```go
package renderer

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jaychinthrajah/ai-upskill/internal/processor"
)

// companyOrder defines the display order for companies.
var companyOrder = []string{
	"OpenAI", "Google", "Anthropic", "Meta", "Microsoft",
	"Mistral", "Apple", "Stability AI", "xAI", "Other/Independent",
}

// RenderMarkdown generates a Markdown report from classified, deduplicated items.
func RenderMarkdown(classified map[string][]processor.DeduplicatedItem, date time.Time, sources []string) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("# AI Daily Report — %s\n\n", date.Format("2006-01-02")))

	// Render known companies in order
	rendered := make(map[string]bool)
	for _, company := range companyOrder {
		items, ok := classified[company]
		if !ok || len(items) == 0 {
			continue
		}
		rendered[company] = true
		renderSection(&b, company, items)
	}

	// Render any remaining companies not in the predefined order
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

	// Footer
	b.WriteString("---\n")
	b.WriteString(fmt.Sprintf("*Generated at %s | Sources: %s*\n",
		time.Now().UTC().Format(time.RFC3339),
		strings.Join(sources, ", ")))

	return b.String()
}

func renderSection(b *strings.Builder, company string, items []processor.DeduplicatedItem) {
	b.WriteString(fmt.Sprintf("## %s\n", company))

	// Sort by number of sources (most-covered first)
	sort.Slice(items, func(i, j int) bool {
		return len(items[i].Sources) > len(items[j].Sources)
	})

	for _, item := range items {
		sourceLinks := make([]string, 0, len(item.Sources))
		for _, src := range item.Sources {
			sourceLinks = append(sourceLinks, fmt.Sprintf("[%s](%s)", src.Name, src.URL))
		}
		b.WriteString(fmt.Sprintf("- **%s** — %s\n", item.Title, strings.Join(sourceLinks, " | ")))
	}
	b.WriteString("\n")
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run:
```bash
go test ./internal/renderer/ -v
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/renderer/markdown.go internal/renderer/markdown_test.go
git commit -m "feat: add Markdown report renderer"
```

---

### Task 10: Wire Up the Generate Command

**Files:**
- Modify: `cmd/ai-report/main.go`

- [ ] **Step 1: Replace the placeholder generate command with the full implementation**

Replace the contents of `cmd/ai-report/main.go` with:

```go
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jaychinthrajah/ai-upskill/internal/config"
	"github.com/jaychinthrajah/ai-upskill/internal/fetcher"
	"github.com/jaychinthrajah/ai-upskill/internal/processor"
	"github.com/jaychinthrajah/ai-upskill/internal/renderer"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ai-report",
	Short: "Generate daily AI news reports",
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate the daily AI news report",
	RunE:  runGenerate,
}

var sourcesCmd = &cobra.Command{
	Use:   "sources",
	Short: "List configured news sources",
	RunE:  runSources,
}

func init() {
	generateCmd.Flags().String("date", "", "Generate report for a specific date (YYYY-MM-DD)")
	generateCmd.Flags().Bool("dry-run", false, "Show report without writing to file")
	generateCmd.Flags().String("config", "config.yaml", "Path to config file")
	rootCmd.AddCommand(generateCmd)

	sourcesCmd.Flags().String("config", "config.yaml", "Path to config file")
	rootCmd.AddCommand(sourcesCmd)
}

func runGenerate(cmd *cobra.Command, args []string) error {
	configPath, _ := cmd.Flags().GetString("config")
	dateStr, _ := cmd.Flags().GetString("date")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	reportDate := time.Now().UTC()
	if dateStr != "" {
		reportDate, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return fmt.Errorf("invalid date format (use YYYY-MM-DD): %w", err)
		}
	}

	since := reportDate.Truncate(24 * time.Hour).Add(-24 * time.Hour)

	fmt.Println("Fetching from sources...")
	items, fetchErrors := fetcher.FetchAll(&cfg.Sources, since, "", "")

	for _, e := range fetchErrors {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", e)
	}

	fmt.Printf("Fetched %d raw items\n", len(items))

	// Classify
	classified := processor.Classify(items, cfg.Keywords["ai_relevance"], cfg.Companies)

	// Deduplicate within each company
	dedupClassified := make(map[string][]processor.DeduplicatedItem)
	totalItems := 0
	for company, companyItems := range classified {
		deduped := processor.Deduplicate(companyItems, cfg.Dedup.LevenshteinThreshold)
		if len(deduped) > 0 {
			dedupClassified[company] = deduped
			totalItems += len(deduped)
		}
	}

	fmt.Printf("After filtering and dedup: %d items across %d companies\n", totalItems, len(dedupClassified))

	// Collect source names
	sourceNames := collectSourceNames(cfg)

	// Render
	report := renderer.RenderMarkdown(dedupClassified, reportDate, sourceNames)

	if dryRun {
		fmt.Println("\n--- DRY RUN ---")
		fmt.Println(report)
		return nil
	}

	// Write file
	outputDir := cfg.Output.Dir
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("creating output dir: %w", err)
	}

	filename := filepath.Join(outputDir, reportDate.Format("2006-01-02")+".md")
	if err := os.WriteFile(filename, []byte(report), 0644); err != nil {
		return fmt.Errorf("writing report: %w", err)
	}

	fmt.Printf("Report written to %s\n", filename)
	return nil
}

func runSources(cmd *cobra.Command, args []string) error {
	configPath, _ := cmd.Flags().GetString("config")

	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	fmt.Println("Configured sources:")
	fmt.Println()

	fmt.Println("Reddit:")
	for _, sub := range cfg.Sources.Reddit.Subreddits {
		fmt.Printf("  - r/%s\n", sub)
	}

	fmt.Printf("\nHacker News (min score: %d)\n", cfg.Sources.HackerNews.MinScore)

	fmt.Println("\nRSS Feeds:")
	for _, feed := range cfg.Sources.RSS {
		fmt.Printf("  - %s (%s)\n", feed.Name, feed.URL)
	}

	return nil
}

func collectSourceNames(cfg *config.Config) []string {
	names := []string{}
	if len(cfg.Sources.Reddit.Subreddits) > 0 {
		names = append(names, "Reddit")
	}
	names = append(names, "Hacker News")
	for _, feed := range cfg.Sources.RSS {
		names = append(names, feed.Name)
	}
	return names
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

- [ ] **Step 2: Verify it compiles**

Run:
```bash
go build -o ai-report ./cmd/ai-report
```

Expected: No errors, binary created.

- [ ] **Step 3: Test the sources command**

Run:
```bash
./ai-report sources
```

Expected: Lists all configured Reddit subreddits, HN settings, and RSS feeds.

- [ ] **Step 4: Test dry-run**

Run:
```bash
./ai-report generate --dry-run
```

Expected: Fetches from real sources, shows a formatted Markdown report to stdout (may have some warnings about feeds).

- [ ] **Step 5: Commit**

```bash
git add cmd/ai-report/main.go
git commit -m "feat: wire up generate and sources commands with full pipeline"
```

---

### Task 11: GitHub Actions Workflow

**Files:**
- Create: `.github/workflows/daily-report.yml`

- [ ] **Step 1: Create the workflow file**

Create `.github/workflows/daily-report.yml`:

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
mkdir -p .github/workflows
git add .github/workflows/daily-report.yml
git commit -m "feat: add GitHub Actions daily report workflow"
```

---

### Task 12: Run All Tests and Final Verification

- [ ] **Step 1: Run all tests**

Run:
```bash
go test ./... -v
```

Expected: All tests pass.

- [ ] **Step 2: Run the full pipeline end-to-end**

Run:
```bash
go build -o ai-report ./cmd/ai-report && ./ai-report generate --dry-run
```

Expected: Fetches real data, filters, categorizes, deduplicates, and prints a Markdown report.

- [ ] **Step 3: Generate an actual report**

Run:
```bash
./ai-report generate
```

Expected: `reports/YYYY-MM-DD.md` created with today's date.

- [ ] **Step 4: Verify the report looks correct**

Run:
```bash
cat reports/$(date -u +%Y-%m-%d).md
```

Expected: Properly formatted Markdown with company sections and source links.

- [ ] **Step 5: Final commit**

```bash
git add reports/
git commit -m "chore: add first generated daily AI report"
```
