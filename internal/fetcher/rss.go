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
			Source:      sourceName,
			PublishedAt: published,
		})
	}

	return items, nil
}
