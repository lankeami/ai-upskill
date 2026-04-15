package fetcher

import (
	"sync"
	"time"

	"github.com/jaychinthrajah/ai-upskill/internal/config"
)

func FetchAll(cfg *config.SourcesConfig, since time.Time, redditBaseURL, hnBaseURL string) ([]RawItem, []error) {
	var mu sync.Mutex
	var allItems []RawItem
	var allErrors []error
	var wg sync.WaitGroup

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
