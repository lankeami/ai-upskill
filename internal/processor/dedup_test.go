package processor

import (
	"testing"
	"time"

	"github.com/jaychinthrajah/ai-upskill/internal/fetcher"
)

func TestDeduplicate(t *testing.T) {
	now := time.Now()

	items := []fetcher.RawItem{
		{Title: "GPT-5 Released by OpenAI", URL: "https://example.com/gpt5", Source: "Reddit", PublishedAt: now},
		{Title: "GPT-5 Released by OpenAI", URL: "https://example.com/gpt5", Source: "HN", PublishedAt: now},
		{Title: "GPT-5 released by OpenAI today", URL: "https://other.com/gpt5", Source: "TechCrunch", PublishedAt: now},
		{Title: "Completely different story", URL: "https://example.com/other", Source: "Reddit", PublishedAt: now},
	}

	result := Deduplicate(items, 0.85)

	if len(result) != 2 {
		t.Fatalf("expected 2 deduplicated items, got %d", len(result))
	}
}

func TestDeduplicatePreservesSources(t *testing.T) {
	now := time.Now()

	items := []fetcher.RawItem{
		{Title: "GPT-5 Released", URL: "https://example.com/gpt5", Source: "Reddit", PublishedAt: now},
		{Title: "GPT-5 Released", URL: "https://example.com/gpt5", Source: "Hacker News", PublishedAt: now},
	}

	result := Deduplicate(items, 0.85)

	if len(result) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result))
	}

	if len(result[0].Sources) != 2 {
		t.Errorf("expected 2 sources, got %d", len(result[0].Sources))
	}
}
