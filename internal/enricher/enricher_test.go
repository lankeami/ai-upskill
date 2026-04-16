package enricher

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jaychinthrajah/ai-upskill/internal/fetcher"
)

// makeServer returns a test HTTP server that serves a minimal HTML page with
// the given meta description.
func makeServer(t *testing.T, description string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `<html><head><meta name="description" content="%s"></head><body></body></html>`, description)
	}))
}

func TestEnrich_EmptyDescriptionGetsEnriched(t *testing.T) {
	srv := makeServer(t, "A great article about Go")
	defer srv.Close()

	items := []fetcher.RawItem{
		{Title: "Go Article", URL: srv.URL, Description: "", Source: "Test"},
	}

	result := Enrich(items, 1)

	if result[0].Description != "A great article about Go" {
		t.Errorf("expected description to be enriched, got %q", result[0].Description)
	}
}

func TestEnrich_NonHTMLDescriptionIsSkipped(t *testing.T) {
	srv := makeServer(t, "Should not replace this")
	defer srv.Close()

	existing := "Already has a good description"
	items := []fetcher.RawItem{
		{Title: "Article", URL: srv.URL, Description: existing, Source: "Test"},
	}

	result := Enrich(items, 1)

	if result[0].Description != existing {
		t.Errorf("expected description to remain %q, got %q", existing, result[0].Description)
	}
}

func TestEnrich_HTMLDescriptionGetsReplaced(t *testing.T) {
	srv := makeServer(t, "Clean description from meta")
	defer srv.Close()

	items := []fetcher.RawItem{
		{Title: "Article", URL: srv.URL, Description: "<p>Some <b>HTML</b> content</p>", Source: "Test"},
	}

	result := Enrich(items, 1)

	if result[0].Description != "Clean description from meta" {
		t.Errorf("expected HTML description to be replaced, got %q", result[0].Description)
	}
}

func TestEnrich_RedditSelfPostIsSkipped(t *testing.T) {
	srv := makeServer(t, "Should not be fetched")
	defer srv.Close()

	// Use a reddit.com/r/ URL — the server URL is irrelevant here since we
	// embed the reddit path directly.
	redditURL := "https://www.reddit.com/r/golang/comments/abc123/some_post"
	items := []fetcher.RawItem{
		{Title: "Reddit Post", URL: redditURL, Description: "", Source: "Reddit"},
	}

	result := Enrich(items, 1)

	// Description should remain empty because Reddit self-posts are skipped.
	if result[0].Description != "" {
		t.Errorf("expected Reddit self-post to be skipped, got description %q", result[0].Description)
	}
}

func TestEnrich_ConcurrencyLimitIsRespected(t *testing.T) {
	const numItems = 10
	const concurrency = 3

	var active int64
	var maxActive int64

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cur := atomic.AddInt64(&active, 1)
		// Track the high-water mark.
		for {
			old := atomic.LoadInt64(&maxActive)
			if cur <= old {
				break
			}
			if atomic.CompareAndSwapInt64(&maxActive, old, cur) {
				break
			}
		}
		// Small sleep to make concurrent requests overlap.
		time.Sleep(10 * time.Millisecond)
		atomic.AddInt64(&active, -1)
		fmt.Fprint(w, `<html><head><meta name="description" content="desc"></head></html>`)
	}))
	defer srv.Close()

	items := make([]fetcher.RawItem, numItems)
	for i := range items {
		items[i] = fetcher.RawItem{Title: fmt.Sprintf("Item %d", i), URL: srv.URL, Description: "", Source: "Test"}
	}

	result := Enrich(items, concurrency)

	// Verify no panic and all items were enriched.
	for i, item := range result {
		if item.Description == "" {
			t.Errorf("item %d: expected description to be enriched", i)
		}
	}

	if maxActive > int64(concurrency) {
		t.Errorf("concurrency exceeded limit: max active goroutines was %d, limit was %d", maxActive, concurrency)
	}
}

func TestEnrich_OriginalSliceNotMutated(t *testing.T) {
	srv := makeServer(t, "New description")
	defer srv.Close()

	original := []fetcher.RawItem{
		{Title: "Article", URL: srv.URL, Description: "", Source: "Test"},
	}
	originalDesc := original[0].Description

	Enrich(original, 1)

	// The original slice should not be mutated.
	if original[0].Description != originalDesc {
		t.Errorf("original slice was mutated: got %q, want %q", original[0].Description, originalDesc)
	}
}
