package processor

import (
	"strings"

	"github.com/jaychinthrajah/ai-upskill/internal/fetcher"
)

type ClassifiedItem struct {
	fetcher.RawItem
	Company string
}

func IsAIRelevant(item fetcher.RawItem, keywords []string) bool {
	text := strings.ToLower(item.Title + " " + item.Description)
	for _, kw := range keywords {
		if strings.Contains(text, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}

func CategorizeByCompany(item fetcher.RawItem, companies map[string][]string) string {
	text := strings.ToLower(item.Title + " " + item.Description)
	for company, keywords := range companies {
		for _, kw := range keywords {
			if strings.Contains(text, strings.ToLower(kw)) {
				return company
			}
		}
	}
	return "Other/Independent"
}

func isRelevant(item fetcher.RawItem, aiKeywords []string, companies map[string][]string) bool {
	if IsAIRelevant(item, aiKeywords) {
		return true
	}
	// Also consider company-specific keywords as AI-relevant signals.
	for _, kws := range companies {
		if IsAIRelevant(item, kws) {
			return true
		}
	}
	return false
}

func Classify(items []fetcher.RawItem, aiKeywords []string, companies map[string][]string) map[string][]fetcher.RawItem {
	result := make(map[string][]fetcher.RawItem)

	for _, item := range items {
		if !isRelevant(item, aiKeywords, companies) {
			continue
		}
		company := CategorizeByCompany(item, companies)
		result[company] = append(result[company], item)
	}

	return result
}
