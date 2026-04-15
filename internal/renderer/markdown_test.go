package renderer

import (
	"strings"
	"testing"
	"time"

	"github.com/jaychinthrajah/ai-upskill/internal/processor"
)

func TestRenderMarkdown(t *testing.T) {
	date := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)

	classified := map[string][]processor.DeduplicatedItem{
		"OpenAI": {
			{
				Title: "GPT-5 Released",
				Sources: []processor.SourceRef{
					{Name: "Reddit", URL: "https://reddit.com/gpt5"},
					{Name: "TechCrunch", URL: "https://techcrunch.com/gpt5"},
				},
			},
		},
		"Google": {
			{
				Title: "Gemini Update",
				Sources: []processor.SourceRef{
					{Name: "Hacker News", URL: "https://hn.com/gemini"},
				},
			},
		},
	}

	sources := []string{"Reddit", "Hacker News", "TechCrunch"}
	result := RenderMarkdown(classified, date, sources)

	if !strings.Contains(result, "# AI Daily Report — 2026-04-15") {
		t.Error("missing report header")
	}
	if !strings.Contains(result, "## OpenAI") {
		t.Error("missing OpenAI section")
	}
	if !strings.Contains(result, "## Google") {
		t.Error("missing Google section")
	}
	if !strings.Contains(result, "**GPT-5 Released**") {
		t.Error("missing GPT-5 item")
	}
	if !strings.Contains(result, "[Reddit]") {
		t.Error("missing Reddit source link")
	}
	if !strings.Contains(result, "[TechCrunch]") {
		t.Error("missing TechCrunch source link")
	}
}

func TestRenderMarkdownOmitsEmptySections(t *testing.T) {
	date := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)

	classified := map[string][]processor.DeduplicatedItem{
		"OpenAI": {
			{
				Title:   "GPT-5",
				Sources: []processor.SourceRef{{Name: "Reddit", URL: "https://reddit.com/gpt5"}},
			},
		},
	}

	result := RenderMarkdown(classified, date, []string{"Reddit"})

	if strings.Contains(result, "## Google") {
		t.Error("should not contain empty Google section")
	}
}
