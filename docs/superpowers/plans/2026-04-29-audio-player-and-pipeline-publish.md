# Audio Player + Pipeline Publish Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Show a native HTML5 audio player on each report page and make the local daily-pipeline skill publish the MP3 to a GitHub Release and write `podcast_url` into the report's front matter.

**Architecture:** Two independent changes — (1) swap the YouTube iframe block in the Jekyll layout for an `<audio>` element keyed on `page.podcast_url`, and (2) insert a publish step into the daily-pipeline skill that runs `gh release create`, builds the deterministic asset URL, then uses an inline Python one-liner to inject `podcast_url` into the report's YAML front matter before committing.

**Tech Stack:** Jekyll/Liquid (layout), `gh` CLI (release), Python 3 inline (front matter edit), Bash (skill steps)

---

### Task 1: Replace YouTube block with HTML5 audio player

**Files:**
- Modify: `_layouts/report.html` — replace `{% if page.youtube_url %}` block

- [ ] **Step 1: Locate the YouTube block**

Open `_layouts/report.html` and find the block that starts with `{% if page.youtube_url %}` and ends with `{% endif %}` (approximately lines 60–70). It contains an `<iframe>` YouTube embed.

- [ ] **Step 2: Replace the block**

Replace the entire `{% if page.youtube_url %}...{% endif %}` block with:

```html
{% if page.podcast_url %}
<div class="podcast-player" style="margin-bottom: 1.5rem; padding: 1rem; background: #f6f8fa; border-radius: 6px;">
  <h3 style="font-size: 0.9rem; margin-bottom: 0.5rem; color: #656d76; text-transform: uppercase; letter-spacing: 0.05em;">Listen to this report</h3>
  <audio controls style="width: 100%; margin-bottom: 0.5rem;">
    <source src="{{ page.podcast_url }}" type="audio/mpeg">
  </audio>
  <p style="margin: 0; font-size: 0.85rem;"><a href="{{ page.podcast_url }}">Download MP3</a></p>
</div>
{% endif %}
```

- [ ] **Step 3: Verify Jekyll builds cleanly**

```bash
bundle exec jekyll build 2>&1 | tail -5
```

Expected output ends with: `done in X.XXX seconds.` and no errors or warnings about `youtube_url` or `podcast_url`.

- [ ] **Step 4: Smoke-test the player renders correctly**

Temporarily add `podcast_url: "https://example.com/test.mp3"` to the front matter of any report in `reports/`, rebuild, and check that `_site/reports/<date>/index.html` contains `<audio controls`. Remove the temporary field afterward.

```bash
# Add temp field
python3 -c "
import re
path = 'reports/2026-04-29.md'
content = open(path).read()
parts = content.split('---', 2)
parts[1] = parts[1].rstrip('\n') + '\npodcast_url: \"https://example.com/test.mp3\"\n'
open(path, 'w').write('---'.join(parts))
"
bundle exec jekyll build 2>&1 | tail -3
grep -c '<audio controls' _site/reports/2026-04-29/index.html
# Expected: 1

# Remove temp field
git checkout reports/2026-04-29.md
```

- [ ] **Step 5: Commit**

```bash
git add _layouts/report.html
git commit -m "feat: replace YouTube embed with HTML5 audio player for podcast_url"
```

---

### Task 2: Add publish step to daily-pipeline skill

**Files:**
- Modify: `.claude/skills/daily-pipeline.md` — insert Step 4 (publish release + front matter), renumber old Step 4 → Step 5, update Done summary

- [ ] **Step 1: Insert Step 4 between the download step and the commit step**

In `.claude/skills/daily-pipeline.md`, after the `## Step 3: Generate Audio` section (which ends after Step 3c Download), insert a new section:

```markdown
## Step 4: Publish GitHub Release

**Check:** Run `gh release view podcast-DATE 2>/dev/null && echo exists || echo missing`. If it prints `exists`, skip this step and note "Release already exists for DATE".

If not, run:
```bash
gh release create podcast-DATE podcasts/DATE.mp3 \
  --title "Podcast — DATE" \
  --notes "Audio podcast for AI Daily Report DATE"
```

If this command fails, stop the pipeline and show the error.

The asset URL is deterministic — no need to parse output:
```
PODCAST_URL=https://github.com/lankeami/ai-upskill/releases/download/podcast-DATE/DATE.mp3
```

**Check:** Read `reports/DATE.md` and check if `podcast_url` is already in the YAML front matter. If so, skip the injection and note "Front matter already has podcast_url for DATE".

If not, inject it using Python:
```bash
python3 -c "
path = 'reports/DATE.md'
url = 'PODCAST_URL'
content = open(path).read()
parts = content.split('---', 2)
if len(parts) >= 3:
    parts[1] = parts[1].rstrip('\n') + '\npodcast_url: \"' + url + '\"\n'
    open(path, 'w').write('---'.join(parts))
    print('Added podcast_url to', path)
else:
    import sys; print('Error: could not parse front matter', file=sys.stderr); sys.exit(1)
"
```

If the Python command fails, stop the pipeline and show the error.
```

- [ ] **Step 2: Rename old Step 4 to Step 5**

Find `## Step 4: Commit & Push` in the file and rename it to `## Step 5: Commit & Push`.

Update its commit message lines to:

```markdown
Use the appropriate commit message:
- If the report was **newly generated** in Step 2: `git commit -m "chore: daily AI report for DATE"`
- If the report **already existed** and only the front matter was updated: `git commit -m "chore: add podcast URL to DATE report"`
```

Also update `git add reports/DATE.md` to ensure it stages the updated front matter.

- [ ] **Step 3: Update the Done summary**

Find the `## Done` section and replace the summary with:

```markdown
Summarize what was done:
- Report: generated or already existed
- Audio: generated or already existed
- GitHub Release: published or already existed
- Front matter: updated or already correct
- Commit: pushed or nothing to commit
```

- [ ] **Step 4: Verify the skill file is valid**

```bash
head -10 .claude/skills/daily-pipeline.md
grep "## Step" .claude/skills/daily-pipeline.md
```

Expected `grep` output:
```
## Step 1: Build Go CLI
## Step 2: Generate Report
## Step 3: Generate Audio
## Step 4: Publish GitHub Release
## Step 5: Commit & Push
```

- [ ] **Step 5: Commit**

```bash
git add .claude/skills/daily-pipeline.md
git commit -m "feat: add GitHub Release publish step to daily-pipeline skill"
```

---

### Task 3: End-to-end smoke test

- [ ] **Step 1: Confirm today's report has no `podcast_url`**

```bash
grep podcast_url reports/2026-04-29.md || echo "not present"
```

Expected: `not present`

- [ ] **Step 2: Dry-run the publish step manually**

```bash
gh release view podcast-2026-04-29 2>/dev/null && echo "release exists" || echo "release missing"
```

Note the output — if the release already exists from a prior run, Step 4 of the pipeline will skip correctly.

- [ ] **Step 3: Push and open draft PR**

```bash
git push -u origin HEAD
gh pr create --draft \
  --title "feat: audio player on site + pipeline publish step" \
  --body "$(cat <<'BODY'
## Summary
- Replaces YouTube iframe with native HTML5 `<audio>` player using `podcast_url` front matter field
- Adds Step 4 to `/daily-pipeline` skill: creates GitHub Release with MP3, injects `podcast_url` into report front matter
- Aligns local pipeline output with what the CI `podcast.yml` workflow produces

## Test Plan
- [ ] Run `/daily-pipeline` locally for a date without an existing release — confirm release is created and `podcast_url` appears in the report
- [ ] Visit the report on the site — confirm audio player renders
- [ ] Re-run pipeline for same date — confirm release/front matter steps are skipped (idempotent)
BODY
)"
```

