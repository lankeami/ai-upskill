# Alexa Companies List Design

**Date:** 2026-05-28  
**Objective:** Add Amazon Alexa as a scoped company entry to daily reports, capturing only Alexa ecosystem content (not broader Amazon company news).

---

## Scope & Naming

**Identifier:** `amazon-alexa` (used in frontmatter companies array)

**Section Header:** `## Amazon Alexa` (title-cased for readability in the report)

**Rationale:** The hyphenated identifier clearly scopes Alexa within Amazon's broader portfolio, while the title-cased header is consistent with other company sections (e.g., `## Google`, `## Apple`).

---

## Content Inclusion Criteria

### Include:
- Alexa voice assistant features and updates
- Echo and Echo-adjacent devices (Show, Dot, Auto, Sub, Hub, etc.)
- Alexa integrations with third-party smart home systems
- Voice-activated smart home announcements powered by Alexa
- Third-party products that feature Alexa as a primary capability

### Exclude:
- AWS (Amazon Web Services) — goes to Other/Independent
- Prime Video, retail, logistics, and other Amazon divisions — go to Other/Independent
- General Amazon company announcements unrelated to Alexa

---

## Decision Making for Edge Cases

When an article mentions both Alexa and another Amazon product (e.g., Alexa + Fire TV, Alexa + Amazon smart display), apply this judgment:

**Primary focus = Alexa ecosystem?** → Goes in `## Amazon Alexa`  
**Primary focus = other product/service, Alexa is incidental?** → Goes in `Other/Independent`

Example: An article about "Alexa gains new smart home capabilities" goes to Alexa. An article about "Amazon's Q1 earnings beat expectations" goes to Other/Independent.

---

## Implementation

### Frontmatter
Add `"amazon-alexa"` to the `companies` array in report frontmatter:
```yaml
companies: ["anthropic", "google", "openai", "xai", "apple", "amazon-alexa", "Other/Independent"]
```

### Quick Reference (Code Comment)
Include in code comments or wiki:
```
amazon-alexa: Voice assistant, Echo devices, smart home integrations, and Alexa-adjacent announcements
```

### Full Documentation
Include in a companies reference doc or README:
```
Amazon Alexa

Include articles about:
- Alexa voice assistant features and updates
- Echo and Echo-adjacent devices (Show, Dot, Auto, etc.)
- Alexa integrations with smart home systems
- Voice-activated smart home announcements powered by Alexa
- Third-party products with Alexa integration

Exclude:
- AWS (Amazon Web Services) — goes to Other/Independent
- Prime Video, retail, logistics — goes to Other/Independent
- General Amazon company news unrelated to Alexa
```

---

## Success Criteria

- `amazon-alexa` appears in report frontmatter alongside other companies
- Articles about Alexa voice assistant and Echo devices are consistently placed under the Alexa section
- Broader Amazon company news is correctly routed to Other/Independent
- The one-liner and fuller documentation guide future filtering decisions consistently
