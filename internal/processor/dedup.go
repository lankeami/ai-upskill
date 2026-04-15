package processor

import (
	"strings"

	"github.com/agnivade/levenshtein"
	"github.com/jaychinthrajah/ai-upskill/internal/fetcher"
)

func wordSet(s string) map[string]struct{} {
	words := strings.Fields(strings.ToLower(s))
	set := make(map[string]struct{}, len(words))
	for _, w := range words {
		set[w] = struct{}{}
	}
	return set
}

func wordContainment(a, b string) float64 {
	setA := wordSet(a)
	setB := wordSet(b)

	if len(setA) == 0 && len(setB) == 0 {
		return 1.0
	}

	intersection := 0
	for w := range setA {
		if _, ok := setB[w]; ok {
			intersection++
		}
	}

	minLen := len(setA)
	if len(setB) < minLen {
		minLen = len(setB)
	}
	if minLen == 0 {
		return 0.0
	}

	return float64(intersection) / float64(minLen)
}

type DeduplicatedItem struct {
	Title       string
	URL         string
	Description string
	Sources     []SourceRef
}

type SourceRef struct {
	Name string
	URL  string
}

func titleSimilarity(a, b string) float64 {
	a = strings.ToLower(strings.TrimSpace(a))
	b = strings.ToLower(strings.TrimSpace(b))

	if a == b {
		return 1.0
	}

	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}
	if maxLen == 0 {
		return 1.0
	}

	dist := levenshtein.ComputeDistance(a, b)
	return 1.0 - float64(dist)/float64(maxLen)
}

func Deduplicate(items []fetcher.RawItem, threshold float64) []DeduplicatedItem {
	var result []DeduplicatedItem

	for _, item := range items {
		merged := false
		for i, existing := range result {
			if item.URL == existing.URL || titleSimilarity(item.Title, existing.Title) >= threshold || wordContainment(item.Title, existing.Title) >= threshold {
				result[i].Sources = append(result[i].Sources, SourceRef{
					Name: item.Source,
					URL:  item.URL,
				})
				merged = true
				break
			}
		}

		if !merged {
			result = append(result, DeduplicatedItem{
				Title:       item.Title,
				URL:         item.URL,
				Description: item.Description,
				Sources: []SourceRef{
					{Name: item.Source, URL: item.URL},
				},
			})
		}
	}

	return result
}
