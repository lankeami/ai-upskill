package fetcher

import (
	"encoding/json"
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
						"title":       "GPT-5 is here",
						"url":         "https://example.com/gpt5",
						"selftext":    "OpenAI just released GPT-5",
						"created_utc": recentTimestamp,
						"permalink":   "/r/artificial/comments/abc123/gpt5_is_here/",
					},
				},
				{
					"data": map[string]interface{}{
						"title":       "Old post",
						"url":         "https://example.com/old",
						"selftext":    "This is old",
						"created_utc": oldTimestamp,
						"permalink":   "/r/artificial/comments/def456/old_post/",
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
