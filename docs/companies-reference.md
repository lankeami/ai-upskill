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
