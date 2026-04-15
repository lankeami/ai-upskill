package fetcher

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type redditResponse struct {
	Data struct {
		Children []struct {
			Data struct {
				Title      string  `json:"title"`
				URL        string  `json:"url"`
				Selftext   string  `json:"selftext"`
				CreatedUTC float64 `json:"created_utc"`
				Permalink  string  `json:"permalink"`
			} `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

func FetchReddit(subreddit string, since time.Time, baseURL string) ([]RawItem, error) {
	url := fmt.Sprintf("https://www.reddit.com/r/%s/new.json?limit=50", subreddit)
	if baseURL != "" {
		url = fmt.Sprintf("%s/r/%s/new.json?limit=50", baseURL, subreddit)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request for r/%s: %w", subreddit, err)
	}
	req.Header.Set("User-Agent", "ai-report/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching r/%s: %w", subreddit, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("r/%s returned status %d", subreddit, resp.StatusCode)
	}

	var result redditResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding r/%s response: %w", subreddit, err)
	}

	var items []RawItem
	for _, child := range result.Data.Children {
		published := time.Unix(int64(child.Data.CreatedUTC), 0).UTC()
		if published.Before(since) {
			continue
		}

		postURL := child.Data.URL
		if child.Data.URL == "" || child.Data.URL == fmt.Sprintf("https://www.reddit.com%s", child.Data.Permalink) {
			postURL = fmt.Sprintf("https://www.reddit.com%s", child.Data.Permalink)
		}

		items = append(items, RawItem{
			Title:       child.Data.Title,
			URL:         postURL,
			Description: child.Data.Selftext,
			Source:      fmt.Sprintf("Reddit r/%s", subreddit),
			PublishedAt: published,
		})
	}

	return items, nil
}
