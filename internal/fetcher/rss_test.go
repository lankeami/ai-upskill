package fetcher

import (
	"fmt"
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
