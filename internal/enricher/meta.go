package enricher

import (
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// Meta holds extracted metadata from a web page's <head>.
type Meta struct {
	Description string
	Title       string // og:title, for potential future use
}

var httpClient = &http.Client{Timeout: 5 * time.Second}

// FetchMeta fetches the URL and extracts meta description and og:title from
// the HTML <head>. Extraction priority for description:
//
//  1. og:description
//  2. meta[name=description]
//  3. twitter:description
//
// Returns an empty Meta on any failure — enrichment is best-effort.
func FetchMeta(url string) Meta {
	resp, err := httpClient.Get(url)
	if err != nil {
		return Meta{}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Meta{}
	}

	return parseMeta(resp.Body)
}

// parseMeta reads an HTML stream and extracts og:title and the best available
// description. It stops once it exits the <head> element.
func parseMeta(r io.Reader) Meta {
	tokenizer := html.NewTokenizer(r)

	var ogDescription, metaDescription, twitterDescription, ogTitle string
	inHead := true

	for inHead {
		tt := tokenizer.Next()
		switch tt {
		case html.ErrorToken:
			inHead = false
		case html.StartTagToken, html.SelfClosingTagToken:
			name, hasAttr := tokenizer.TagName()
			tagName := string(name)

			if tagName == "body" {
				inHead = false
				break
			}

			if tagName != "meta" || !hasAttr {
				break
			}

			attrs := extractAttrs(tokenizer)
			property := strings.ToLower(attrs["property"])
			metaName := strings.ToLower(attrs["name"])
			content := attrs["content"]

			switch {
			case property == "og:description":
				ogDescription = content
			case property == "og:title":
				ogTitle = content
			case metaName == "description":
				metaDescription = content
			case metaName == "twitter:description" || property == "twitter:description":
				twitterDescription = content
			}

		case html.EndTagToken:
			name, _ := tokenizer.TagName()
			if string(name) == "head" {
				inHead = false
			}
		}
	}

	description := firstNonEmpty(ogDescription, metaDescription, twitterDescription)
	return Meta{Description: description, Title: ogTitle}
}

// extractAttrs drains remaining attributes from the current token and returns
// them as a map of lowercase key → value.
func extractAttrs(t *html.Tokenizer) map[string]string {
	attrs := make(map[string]string)
	for {
		key, val, more := t.TagAttr()
		attrs[strings.ToLower(string(key))] = string(val)
		if !more {
			break
		}
	}
	return attrs
}

// firstNonEmpty returns the first non-empty string from the provided values.
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
