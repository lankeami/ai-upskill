# Changelog

## 2026-09-04

- [e4eca68](https://github.com/lankeami/ai-upskill/commit/e4eca683927d1fc500c92b3acd200321ccb45ff4) chore: daily AI report for 2026-09-04

## 2026-09-03

- [2bdf706](https://github.com/lankeami/ai-upskill/commit/2bdf7061c43fc0cc24788be9a0785b255329673c) docs: add manual cookie capture method to README auth setup
  Document Method B for when `notebooklm login` hangs: capture cookies via Chrome DevTools Network tab and convert to storage state JSON. Also fix storage state path to match actual profile directory.
- [9a6cc4f](https://github.com/lankeami/ai-upskill/commit/9a6cc4f973b5af27dd14cf8358c7c8c425b74e92) chore: add podcast URL to 2026-09-03 report
- [423b352](https://github.com/lankeami/ai-upskill/commit/423b35272b2fbab0050c03343891e492b4113e45) Fix daily report CI failure by excluding CHANGELOG.md from Jekyll
  CHANGELOG.md was added in #52 but not added to the Jekyll exclude list, causing test-jekyll-output.sh to fail on unexpected CHANGELOG.html.

## 2026-09-01

- [d855dbf](https://github.com/lankeami/ai-upskill/commit/d855dbf0c9cb00a1b16d20a10242b036daefe074) fix: restructure skill files so they appear in Claude Code sidebar
  Move flat .claude/skills/<name>.md files to .claude/skills/<name>/SKILL.md directory convention that the Claude Code skill loader requires.
- [3a87289](https://github.com/lankeami/ai-upskill/commit/3a872899819ddd69bcd068d89c85a866caeeec0a) feat: skip podcast generation when audio already published (#47)
  Both the GitHub workflow and local daily-pipeline skill now check for an existing podcast-DATE release before running the expensive NotebookLM generation. Add --force flag to override when needed.
