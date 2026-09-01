# Changelog

## 2026-09-01

- [d855dbf](https://github.com/lankeami/ai-upskill/commit/d855dbf0c9cb00a1b16d20a10242b036daefe074) fix: restructure skill files so they appear in Claude Code sidebar
  Move flat .claude/skills/<name>.md files to .claude/skills/<name>/SKILL.md directory convention that the Claude Code skill loader requires.
- [3a87289](https://github.com/lankeami/ai-upskill/commit/3a872899819ddd69bcd068d89c85a866caeeec0a) feat: skip podcast generation when audio already published (#47)
  Both the GitHub workflow and local daily-pipeline skill now check for an existing podcast-DATE release before running the expensive NotebookLM generation. Add --force flag to override when needed.
