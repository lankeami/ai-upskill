# Changelog

## 2026-09-01

- [3a87289](https://github.com/lankeami/ai-upskill/commit/3a872899819ddd69bcd068d89c85a866caeeec0a) feat: skip podcast generation when audio already published (#47)
  Both the GitHub workflow and local daily-pipeline skill now check for an existing podcast-DATE release before running the expensive NotebookLM generation. Add --force flag to override when needed.
