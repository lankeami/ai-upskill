# ai-upskill Project Instructions

## Jekyll / GitHub Pages

**Jekyll Exclusion Rule:** Jekyll must only serve `index.md`, `today.md`, and `reports/*.md`. All other directories and top-level Markdown files must be listed in `_config.yml`'s `exclude` list. **Do NOT exclude `today.md`** — it provides the `/today` redirect and must be built by Jekyll. When adding new top-level directories or Markdown files to the repo, add them to the exclude list and verify with `scripts/test-jekyll-output.sh`.
