# Design: Move Other/Independent Section to Bottom of Reports

**Date:** 2026-05-07
**Status:** Approved

## Problem

The "Other/Independent" section was listed in `companyOrder` in `internal/renderer/markdown.go`, placing it at position 9 (after xAI). Any companies not in `companyOrder` were appended alphabetically after it, meaning "Other/Independent" was not actually last.

## Change

Remove `"Other/Independent"` from the `companyOrder` slice in `internal/renderer/markdown.go:32-35`.

The existing fallback logic (lines 49–64) already handles this correctly: it separates "Other/Independent" from other unrecognised companies, appends unrecognised companies alphabetically, then appends "Other/Independent" last.

### Before

```go
var companyOrder = []string{
    "OpenAI", "Google", "Anthropic", "Meta", "Microsoft",
    "Mistral", "Apple", "Stability AI", "xAI", "Other/Independent",
}
```

### After

```go
var companyOrder = []string{
    "OpenAI", "Google", "Anthropic", "Meta", "Microsoft",
    "Mistral", "Apple", "Stability AI", "xAI",
}
```

## Result

Section order in generated reports:

1. Named companies in priority order (OpenAI → xAI)
2. Any unrecognised companies (alphabetical)
3. **Other/Independent** (always last)

## Files Affected

- `internal/renderer/markdown.go` — one-line change
- `internal/renderer/markdown_test.go` — update/add test asserting Other/Independent is last

## Out of Scope

No changes to classification logic, front matter, or any other pipeline step.
