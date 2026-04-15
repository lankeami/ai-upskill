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
