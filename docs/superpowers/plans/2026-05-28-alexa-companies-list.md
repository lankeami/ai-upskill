# Alexa Companies List Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `amazon-alexa` as a scoped company entry to the daily report system, with clear filtering rules to distinguish Alexa-specific content from broader Amazon company news.

**Architecture:** We'll create a companies reference document that serves as the single source of truth for filtering decisions. This document includes both a one-liner reference (for developers embedding rules in code) and detailed filtering criteria (for consistent decision-making). The identifier `amazon-alexa` will be added to report frontmatter metadata, and the section header `## Amazon Alexa` will be used in report body sections.

**Tech Stack:** Markdown documentation, Jekyll (static site generation), existing Go CLI report generation

---

### Task 1: Create or Update Companies Reference Document

**Files:**
- Create or modify: `docs/companies-reference.md`

- [ ] **Step 1: Check if companies-reference.md exists**

Run: `ls -la docs/companies-reference.md`

If file exists, proceed to Step 2. If not, proceed to Step 3.

- [ ] **Step 2: Read the existing companies-reference.md**

Run: `cat docs/companies-reference.md`

Observe the existing structure and format so you can match the style when adding Alexa.

- [ ] **Step 3: Create companies-reference.md if it doesn't exist**

Create the file with the following content:

```markdown
# Companies Reference

This document serves as the source of truth for company identifiers, section headers, and filtering criteria used in daily AI reports.

## Company Identifiers

Each company has a unique identifier used in report frontmatter's `companies` array:
- `anthropic` → ## Anthropic
- `google` → ## Google
- `openai` → ## OpenAI
- `xai` → ## XAI
- `apple` → ## Apple
- `amazon-alexa` → ## Amazon Alexa
- `Other/Independent` → Catch-all for unscoped content

## Amazon Alexa

**Identifier:** `amazon-alexa`  
**Section Header:** `## Amazon Alexa`

**One-liner for code comments:**
Voice assistant, Echo devices, smart home integrations, and Alexa-adjacent announcements

**Include:**
- Alexa voice assistant features and updates
- Echo and Echo-adjacent devices (Show, Dot, Auto, Sub, Hub, etc.)
- Alexa integrations with third-party smart home systems
- Voice-activated smart home announcements powered by Alexa
- Third-party products that feature Alexa as a primary capability

**Exclude:**
- AWS (Amazon Web Services) — goes to Other/Independent
- Prime Video, retail, logistics, and other Amazon divisions — go to Other/Independent
- General Amazon company announcements unrelated to Alexa

**Edge Case Decision Rule:**
When an article mentions both Alexa and another Amazon product, place it in `## Amazon Alexa` if the primary focus is Alexa ecosystem. If the primary focus is the other product and Alexa is incidental, place it in Other/Independent.
```

If the file already exists (from Step 2), add the "## Amazon Alexa" section with the content above, matching the existing format for other companies.

- [ ] **Step 4: Verify the file is readable and well-formatted**

Run: `cat docs/companies-reference.md | head -40`

Expected: First 40 lines show clear structure with company identifiers and section headers.

- [ ] **Step 5: Commit the companies reference document**

```bash
git add docs/companies-reference.md
git commit -m "docs: add amazon-alexa company reference with filtering criteria"
```

---

### Task 2: Verify Companies List in Example Report

**Files:**
- Examine: `reports/2026-05-08.md` (or most recent report)
- No modifications needed in this task; verification only

- [ ] **Step 1: Check the current companies array in a recent report**

Run: `head -20 reports/2026-05-08.md`

Observe the `companies` array in the frontmatter. Expected format:
```yaml
companies: ["anthropic", "google", "openai", "xai", "apple", "Other/Independent"]
```

- [ ] **Step 2: Understand where reports are generated**

Run: `grep -r "companies:" reports/ | head -3`

Observe that reports have the frontmatter pattern. This confirms that the Go CLI that generates reports sets the companies array.

- [ ] **Step 3: Locate the Go CLI code that generates reports**

Run: `find . -name "*.go" -type f | xargs grep -l "companies" | head -5`

Expected: Find Go files in `cmd/ai-report/` or similar that define the frontmatter/companies structure.

- [ ] **Step 4: Read the Go CLI code that sets the companies array**

Open the relevant Go file (likely `cmd/ai-report/main.go` or `cmd/ai-report/generate.go` or similar) and locate the section that defines:
- The companies array
- Hardcoded company list
- Or frontmatter generation logic

Record the exact line numbers and structure.

- [ ] **Step 5: Verify structure matches documentation**

Confirm that the companies list in the Go code matches what we documented in `companies-reference.md`. If the Go code has a hardcoded list that needs updating, document it for Task 3.

- [ ] **Step 6: Commit verification notes (optional)**

If no changes are needed, no commit required. If you found that the Go code needs updating, you've prepared the groundwork for Task 3.

---

### Task 3: Update Go CLI Companies List (If Needed)

**Files:**
- Modify: `cmd/ai-report/main.go` or wherever the companies list is defined (exact path from Task 2, Step 4)

- [ ] **Step 1: Determine if Go code needs updating**

From Task 2, Step 5: Did you find a hardcoded companies list in the Go code that needs `amazon-alexa` added?

If **NO** (the companies are dynamically sourced or the list is complete), skip this task and proceed to Task 4.

If **YES** (hardcoded list exists and needs updating), proceed to Step 2.

- [ ] **Step 2: Read the relevant Go file**

Run: `cat cmd/ai-report/main.go | sed -n '<START_LINE>,<END_LINE>p'`

Replace `<START_LINE>` and `<END_LINE>` with the line numbers from Task 2, Step 4.

Observe the exact structure of the companies list (e.g., array, slice, map).

- [ ] **Step 3: Write a test for the companies list (TDD)**

Create a test file `cmd/ai-report/main_test.go` (or update if exists) with:

```go
package main

import (
	"testing"
)

func TestCompaniesListIncludesAlexa(t *testing.T) {
	expectedCompanies := []string{
		"anthropic",
		"google",
		"openai",
		"xai",
		"apple",
		"amazon-alexa",
		"Other/Independent",
	}
	
	actualCompanies := GetCompaniesForReport() // or however the function is named
	
	if len(actualCompanies) != len(expectedCompanies) {
		t.Errorf("Expected %d companies, got %d", len(expectedCompanies), len(actualCompanies))
	}
	
	for _, expected := range expectedCompanies {
		found := false
		for _, actual := range actualCompanies {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected company %q not found in companies list", expected)
		}
	}
}
```

- [ ] **Step 4: Run the test to verify it fails**

Run: `go test ./cmd/ai-report -v -run TestCompaniesListIncludesAlexa`

Expected: FAIL with message indicating `amazon-alexa` not found in the companies list.

- [ ] **Step 5: Update the Go code to add amazon-alexa**

Modify the hardcoded companies list in the Go file to include `"amazon-alexa"` in the appropriate position (between `"apple"` and `"Other/Independent"` based on the design spec).

Exact code change depends on the structure from Step 2, but typically:

```go
companies := []string{
    "anthropic",
    "google",
    "openai",
    "xai",
    "apple",
    "amazon-alexa",  // Add this line
    "Other/Independent",
}
```

- [ ] **Step 6: Run the test again to verify it passes**

Run: `go test ./cmd/ai-report -v -run TestCompaniesListIncludesAlexa`

Expected: PASS.

- [ ] **Step 7: Run all tests to ensure no regressions**

Run: `go test ./cmd/ai-report -v`

Expected: All tests pass (or same pass/fail status as before, with no new failures).

- [ ] **Step 8: Commit the Go code update**

```bash
git add cmd/ai-report/main.go cmd/ai-report/main_test.go
git commit -m "feat: add amazon-alexa to companies list in report generation"
```

---

### Task 4: Add Code Comment with One-Liner Reference

**Files:**
- Modify: The Go file from Task 2, Step 4 (same file where companies are defined)

- [ ] **Step 1: Locate the companies list definition in Go code**

From earlier tasks, you know the exact line numbers. Open that file and position at the companies list.

- [ ] **Step 2: Add a comment block above the companies list**

Insert a comment that explains what each company identifier means. The one-liner for amazon-alexa (from companies-reference.md) is:

```
amazon-alexa: Voice assistant, Echo devices, smart home integrations, and Alexa-adjacent announcements
```

Add it to the comment block alongside other companies:

```go
// Companies included in daily reports.
// Each identifier maps to a section header in the generated report.
//
// Identifier mapping:
// - anthropic: AI research company
// - google: Search, AI, and cloud services
// - openai: AI research company
// - xai: Elon Musk's AI company
// - apple: Consumer electronics and software
// - amazon-alexa: Voice assistant, Echo devices, smart home integrations, and Alexa-adjacent announcements
// - Other/Independent: Catch-all for unscoped content
var companies = []string{
    ...
}
```

- [ ] **Step 3: Verify the comment is clear and accurate**

Read the comment aloud to yourself. Does it accurately convey the scope and intent of `amazon-alexa` for a developer who's filtering content?

- [ ] **Step 4: Run tests to ensure no breakage**

Run: `go test ./cmd/ai-report -v`

Expected: All tests pass (comments don't affect execution).

- [ ] **Step 5: Commit the code comment**

```bash
git add cmd/ai-report/main.go
git commit -m "docs: add amazon-alexa filtering one-liner to companies list comment"
```

---

### Task 5: Verify End-to-End Integration

**Files:**
- Examine: Any new reports generated after changes
- No modifications; verification only

- [ ] **Step 1: Rebuild the Go CLI to ensure it compiles**

Run: `go build -o bin/ai-report ./cmd/ai-report`

Expected: No errors. Executable created at `bin/ai-report`.

- [ ] **Step 2: Check that companies reference doc is in exclude list**

Run: `grep -A 20 "exclude:" _config.yml | grep companies-reference`

If `docs/companies-reference.md` is NOT in the exclude list and does NOT need to be served by Jekyll, add it:

Edit `_config.yml` and add `docs/companies-reference.md` to the exclude list.

Reasoning: This is internal documentation, not a public report. It should not be served as a live page.

- [ ] **Step 3: Verify Jekyll build doesn't break**

Run: `scripts/test-jekyll-output.sh` (or equivalent Jekyll validation command)

Expected: Jekyll builds successfully and serves only the expected pages (index.md, today.md, reports/*.md).

- [ ] **Step 4: Commit Jekyll config update (if needed)**

If you made changes to `_config.yml`:

```bash
git add _config.yml
git commit -m "docs: add docs/companies-reference.md to Jekyll exclude list"
```

---

## Success Criteria

- [ ] `docs/companies-reference.md` exists and documents all companies including `amazon-alexa`
- [ ] Go CLI code includes `amazon-alexa` in the companies list (if hardcoded)
- [ ] Code comment provides the one-liner reference for developers
- [ ] All Go tests pass, including the new `TestCompaniesListIncludesAlexa` test
- [ ] Jekyll build succeeds and companies-reference.md is not served as a public page
- [ ] Commits are frequent and atomic (one logical change per commit)
