# Podcast Audio Customization

How to control what NotebookLM talks about in the daily audio podcast.
## Where audio generation is configured

`scripts/generate-podcast.py` — two call sites:

| Function | Line | Used by |
|---|---|---|
| `generate_podcast()` | ~255 | Legacy monolithic mode (`--date`, `--media-type` flags) |
| `start_generation()` | ~133 | Step-by-step mode (`start` subcommand, used by CI) |

Both call `client.artifacts.generate_audio()`.

---

## The `instructions` parameter

`generate_audio()` accepts an optional `instructions: str` argument:

```python
await client.artifacts.generate_audio(
    nb.id,
    audio_format=AudioFormat.DEEP_DIVE,
    audio_length=AudioLength.SHORT,
    instructions="Your custom instructions here.",
)
```

**Without `instructions`:** NotebookLM picks 1–2 topics from the report and discusses them in depth. Most of the report is ignored.

**With `instructions`:** The hosts follow your direction. You can tell them to cover every headline, change tone, focus on a theme, etc.

---

## Explain every headline (recommended)

Paste this as the `instructions` value:

```
Go through every headline in the report in order. For each one, read the headline aloud and then explain in 1–2 sentences what it means and why it matters to someone learning about AI.
```

Both call sites need updating. In `generate_podcast()` (~line 255) and `start_generation()` (~line 133), change:

```python
# Before
status = await client.artifacts.generate_audio(
    nb.id,
    audio_format=AudioFormat.DEEP_DIVE,
    audio_length=AudioLength.SHORT,
)

# After
AUDIO_INSTRUCTIONS = (
    "Go through every headline in the report in order. "
    "For each one, read the headline aloud and then explain in 1–2 sentences "
    "what it means and why it matters to someone learning about AI."
)

status = await client.artifacts.generate_audio(
    nb.id,
    audio_format=AudioFormat.DEEP_DIVE,
    audio_length=AudioLength.SHORT,
    instructions=AUDIO_INSTRUCTIONS,
)
```

---

## Audio length

`AudioLength` controls episode duration. Options: `SHORT`, `DEFAULT`, `LONG`.

Currently set to `SHORT` in both call sites. Change to `DEFAULT` or `LONG` if the report has many sections and the hosts are rushing through content.

```python
audio_length=AudioLength.DEFAULT,  # ~10–15 min, better for full-headline coverage
```
