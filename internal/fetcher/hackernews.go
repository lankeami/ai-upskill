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

func FetchHackerNews(minScore int, since time.Time, baseURL string) ([]RawItem, error) {
	apiBase := "https://hacker-news.firebaseio.com"
	if baseURL != "" {
		apiBase = baseURL
	}

	resp, err := http.Get(fmt.Sprintf("%s/v0/topstories.json", apiBase))
	if err != nil {
		return nil, fmt.Errorf("fetching HN top stories: %w", err)
	}
	defer resp.Body.Close()

	var storyIDs []int
	if err := json.NewDecoder(resp.Body).Decode(&storyIDs); err != nil {
		return nil, fmt.Errorf("decoding HN top stories: %w", err)
	}

	if len(storyIDs) > 100 {
		storyIDs = storyIDs[:100]
	}

	var mu sync.Mutex
	var items []RawItem
	var wg sync.WaitGroup

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
