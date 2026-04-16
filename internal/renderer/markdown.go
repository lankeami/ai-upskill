package renderer

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jaychinthrajah/ai-upskill/internal/processor"
)

var companyOrder = []string{
	"OpenAI", "Google", "Anthropic", "Meta", "Microsoft",
	"Mistral", "Apple", "Stability AI", "xAI", "Other/Independent",
}

func RenderMarkdown(classified map[string][]processor.DeduplicatedItem, date time.Time, sources []string) string {
	// Collect ordered companies and total item count for front matter
	orderedCompanies := make([]string, 0)
	rendered := make(map[string]bool)
	for _, company := range companyOrder {
		items, ok := classified[company]
		if !ok || len(items) == 0 {
			continue
		}
		rendered[company] = true
		orderedCompanies = append(orderedCompanies, company)
	}
	remaining := make([]string, 0)
	for company := range classified {
		if !rendered[company] {
			remaining = append(remaining, company)
		}
	}
	sort.Strings(remaining)
	orderedCompanies = append(orderedCompanies, remaining...)

	itemCount := 0
	for _, company := range orderedCompanies {
		itemCount += len(classified[company])
	}

	quotedCompanies := make([]string, len(orderedCompanies))
	for i, c := range orderedCompanies {
		quotedCompanies[i] = fmt.Sprintf("%q", c)
	}

	var b strings.Builder

	// Jekyll front matter
	b.WriteString("---\n")
	b.WriteString("layout: report\n")
	b.WriteString(fmt.Sprintf("title: \"AI Daily Report \u2014 %s\"\n", date.Format("2006-01-02")))
	b.WriteString(fmt.Sprintf("date: %s\n", date.Format("2006-01-02")))
	b.WriteString(fmt.Sprintf("companies: [%s]\n", strings.Join(quotedCompanies, ", ")))
	b.WriteString(fmt.Sprintf("item_count: %d\n", itemCount))
	b.WriteString("---\n\n")

	b.WriteString(fmt.Sprintf("# AI Daily Report \u2014 %s\n\n", date.Format("2006-01-02")))

	for _, company := range orderedCompanies {
		renderSection(&b, company, classified[company])
	}

	b.WriteString("---\n")
	b.WriteString(fmt.Sprintf("*Generated at %s | Sources: %s*\n",
		time.Now().UTC().Format(time.RFC3339),
		strings.Join(sources, ", ")))

	return b.String()
}

func renderSection(b *strings.Builder, company string, items []processor.DeduplicatedItem) {
	b.WriteString(fmt.Sprintf("## %s\n", company))

	sort.Slice(items, func(i, j int) bool {
		return len(items[i].Sources) > len(items[j].Sources)
	})

	for _, item := range items {
		sourceLinks := make([]string, 0, len(item.Sources))
		for _, src := range item.Sources {
			sourceLinks = append(sourceLinks, fmt.Sprintf("[%s](%s)", src.Name, src.URL))
		}
		b.WriteString(fmt.Sprintf("- **%s** \u2014 %s\n", item.Title, strings.Join(sourceLinks, " | ")))
	}
	b.WriteString("\n")
}
