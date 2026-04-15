package processor

import (
	"testing"
	"time"

	"github.com/jaychinthrajah/ai-upskill/internal/fetcher"
)

func TestIsAIRelevant(t *testing.T) {
	keywords := []string{"AI", "LLM", "GPT", "machine learning"}

	tests := []struct {
		title    string
		desc     string
		expected bool
	}{
		{"New GPT-5 model released", "", true},
		{"Machine learning breakthrough", "", true},
		{"Best pizza in NYC", "", false},
		{"No keywords here", "But the description mentions AI research", true},
		{"OpenAI announces new LLM", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			item := fetcher.RawItem{Title: tt.title, Description: tt.desc}
			got := IsAIRelevant(item, keywords)
			if got != tt.expected {
				t.Errorf("IsAIRelevant(%q, %q) = %v, want %v", tt.title, tt.desc, got, tt.expected)
			}
		})
	}
}

func TestCategorizeByCompany(t *testing.T) {
	companies := map[string][]string{
		"OpenAI":    {"openai", "gpt", "chatgpt"},
		"Google":    {"google", "gemini", "deepmind"},
		"Anthropic": {"anthropic", "claude"},
	}

	tests := []struct {
		title    string
		expected string
	}{
		{"OpenAI releases GPT-5", "OpenAI"},
		{"Google's Gemini gets upgrade", "Google"},
		{"Claude improves at coding", "Anthropic"},
		{"New open-source model from TII", "Other/Independent"},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			item := fetcher.RawItem{Title: tt.title}
			got := CategorizeByCompany(item, companies)
			if got != tt.expected {
				t.Errorf("CategorizeByCompany(%q) = %q, want %q", tt.title, got, tt.expected)
			}
		})
	}
}

func TestClassify(t *testing.T) {
	keywords := []string{"AI", "LLM", "GPT"}
	companies := map[string][]string{
		"OpenAI": {"openai", "gpt"},
		"Google": {"google", "gemini"},
	}

	items := []fetcher.RawItem{
		{Title: "GPT-5 released by OpenAI", URL: "https://example.com/gpt5", Source: "Reddit", PublishedAt: time.Now()},
		{Title: "Best pizza in NYC", URL: "https://example.com/pizza", Source: "Reddit", PublishedAt: time.Now()},
		{Title: "Google Gemini update", URL: "https://example.com/gemini", Source: "HN", PublishedAt: time.Now()},
	}

	result := Classify(items, keywords, companies)

	if len(result) != 2 {
		t.Fatalf("expected 2 companies, got %d", len(result))
	}

	openaiItems, ok := result["OpenAI"]
	if !ok || len(openaiItems) != 1 {
		t.Errorf("expected 1 OpenAI item, got %d", len(openaiItems))
	}

	googleItems, ok := result["Google"]
	if !ok || len(googleItems) != 1 {
		t.Errorf("expected 1 Google item, got %d", len(googleItems))
	}
}
