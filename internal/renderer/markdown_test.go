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

func TestRenderMarkdownIncludesFrontMatter(t *testing.T) {
	date := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)

	classified := map[string][]processor.DeduplicatedItem{
		"OpenAI": {
			{
				Title:   "GPT-5 Released",
				Sources: []processor.SourceRef{{Name: "Reddit", URL: "https://reddit.com/gpt5"}},
			},
		},
		"Google": {
			{
				Title:   "Gemini Update",
				Sources: []processor.SourceRef{{Name: "HN", URL: "https://hn.com/gemini"}},
			},
			{
				Title:   "Bard News",
				Sources: []processor.SourceRef{{Name: "RSS", URL: "https://rss.com/bard"}},
			},
		},
	}

	result := RenderMarkdown(classified, date, []string{"Reddit"})

	if !strings.HasPrefix(result, "---\n") {
		t.Fatal("report must start with front matter delimiter")
	}
	if !strings.Contains(result, "layout: report") {
		t.Error("missing layout field in front matter")
	}
	if !strings.Contains(result, "date: 2026-04-15") {
		t.Error("missing date field in front matter")
	}
	if !strings.Contains(result, "item_count: 3") {
		t.Error("missing or incorrect item_count in front matter")
	}
	if !strings.Contains(result, "\"OpenAI\"") {
		t.Error("missing OpenAI in companies list")
	}
	if !strings.Contains(result, "\"Google\"") {
		t.Error("missing Google in companies list")
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

func TestCleanDescription(t *testing.T) {
	t.Run("empty returns empty", func(t *testing.T) {
		if got := cleanDescription(""); got != "" {
			t.Errorf("expected empty, got %q", got)
		}
	})

	t.Run("strips HTML tags", func(t *testing.T) {
		got := cleanDescription("<p>Hello <b>world</b></p>")
		if got != "Hello world" {
			t.Errorf("expected %q, got %q", "Hello world", got)
		}
	})

	t.Run("truncates long description at 200 runes", func(t *testing.T) {
		long := strings.Repeat("a", 250)
		got := cleanDescription(long)
		runes := []rune(got)
		// 200 chars + … suffix
		if len(runes) != 201 {
			t.Errorf("expected 201 runes (200 + ellipsis), got %d", len(runes))
		}
		if !strings.HasSuffix(got, "…") {
			t.Error("truncated description should end with …")
		}
	})

	t.Run("short description not truncated", func(t *testing.T) {
		short := "Short desc"
		got := cleanDescription(short)
		if got != short {
			t.Errorf("expected %q, got %q", short, got)
		}
	})
}

func TestRenderSectionWithDescriptions(t *testing.T) {
	date := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)

	t.Run("item with description renders description line", func(t *testing.T) {
		classified := map[string][]processor.DeduplicatedItem{
			"OpenAI": {
				{
					Title:       "GPT-5 Released",
					Description: "A major new model from OpenAI.",
					Sources:     []processor.SourceRef{{Name: "Reddit", URL: "https://reddit.com/gpt5"}},
				},
			},
		}
		result := RenderMarkdown(classified, date, []string{"Reddit"})
		if !strings.Contains(result, "  A major new model from OpenAI.") {
			t.Error("description should appear indented with 2 spaces")
		}
	})

	t.Run("item without description has no extra blank line", func(t *testing.T) {
		classified := map[string][]processor.DeduplicatedItem{
			"OpenAI": {
				{
					Title:   "GPT-5 Released",
					Sources: []processor.SourceRef{{Name: "Reddit", URL: "https://reddit.com/gpt5"}},
				},
			},
		}
		result := RenderMarkdown(classified, date, []string{"Reddit"})
		// The bullet line should be directly followed by a newline then the section trailing newline,
		// not an extra indented blank line.
		if strings.Contains(result, "**GPT-5 Released**\n\n  ") {
			t.Error("item without description should not have an indented description line")
		}
	})

	t.Run("long description truncated at 200 chars", func(t *testing.T) {
		longDesc := strings.Repeat("x", 300)
		classified := map[string][]processor.DeduplicatedItem{
			"OpenAI": {
				{
					Title:       "Big Article",
					Description: longDesc,
					Sources:     []processor.SourceRef{{Name: "Reddit", URL: "https://reddit.com/big"}},
				},
			},
		}
		result := RenderMarkdown(classified, date, []string{"Reddit"})
		if !strings.Contains(result, "…") {
			t.Error("long description should be truncated with … suffix")
		}
		// Verify rendered description line is not more than 200 runes + "  " prefix + "…"
		for _, line := range strings.Split(result, "\n") {
			if strings.HasPrefix(line, "  ") && strings.HasSuffix(line, "…") {
				content := strings.TrimPrefix(line, "  ")
				runes := []rune(content)
				if len(runes) > 201 { // 200 chars + …
					t.Errorf("description line too long: %d runes", len(runes))
				}
			}
		}
	})

	t.Run("HTML in description is stripped", func(t *testing.T) {
		classified := map[string][]processor.DeduplicatedItem{
			"OpenAI": {
				{
					Title:       "HTML Article",
					Description: "<p>Some <strong>bold</strong> text</p>",
					Sources:     []processor.SourceRef{{Name: "Reddit", URL: "https://reddit.com/html"}},
				},
			},
		}
		result := RenderMarkdown(classified, date, []string{"Reddit"})
		if strings.Contains(result, "<p>") || strings.Contains(result, "<strong>") {
			t.Error("HTML tags should be stripped from description")
		}
		if !strings.Contains(result, "Some bold text") {
			t.Error("description text content should be preserved after stripping HTML")
		}
	})
}
