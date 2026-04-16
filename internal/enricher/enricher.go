package enricher

import (
	"strings"
	"sync"

	"github.com/jaychinthrajah/ai-upskill/internal/fetcher"
)

// needsEnrichment returns true if the item's description is empty or contains
// HTML tags (heuristic: presence of both "<" and ">").
func needsEnrichment(desc string) bool {
	if desc == "" {
		return true
	}
	return strings.Contains(desc, "<") && strings.Contains(desc, ">")
}

// isRedditSelfPost returns true for Reddit self-post URLs that don't have
// useful external content to scrape.
func isRedditSelfPost(url string) bool {
	return strings.Contains(url, "reddit.com/r/")
}

// Enrich fetches meta descriptions for items that need enrichment, using up to
// concurrency goroutines in parallel. Items with existing non-HTML descriptions
// and Reddit self-post URLs are left untouched. Returns the updated slice.
func Enrich(items []fetcher.RawItem, concurrency int) []fetcher.RawItem {
	if concurrency <= 0 {
		concurrency = 1
	}

	result := make([]fetcher.RawItem, len(items))
	copy(result, items)

	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i := range result {
		if !needsEnrichment(result[i].Description) {
			continue
		}
		if isRedditSelfPost(result[i].URL) {
			continue
		}

		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			meta := FetchMeta(result[idx].URL)
			if meta.Description != "" {
				mu.Lock()
				result[idx].Description = meta.Description
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()
	return result
}
