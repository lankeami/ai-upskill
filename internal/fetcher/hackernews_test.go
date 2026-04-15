package fetcher

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetchHackerNews(t *testing.T) {
	now := time.Now().UTC()

	mux := http.NewServeMux()

	mux.HandleFunc("/v0/topstories.json", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]int{1, 2})
	})

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
