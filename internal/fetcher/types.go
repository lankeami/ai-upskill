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
