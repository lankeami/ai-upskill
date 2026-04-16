#!/usr/bin/env bash
set -euo pipefail

echo "Building Jekyll site..."
bundle exec jekyll build --destination _site 2>&1

echo ""
echo "Checking _site output..."

# Collect all HTML files in _site
html_files=$(find _site -name '*.html' -type f | sort)

echo "HTML files found:"
echo "$html_files"
echo ""

# Check each HTML file is either index.html or under reports/
fail=0
while IFS= read -r file; do
  # Strip _site/ prefix
  rel="${file#_site/}"
  case "$rel" in
    index.html) ;;
    reports/*.html) ;;
    *)
      echo "UNEXPECTED FILE: $rel"
      fail=1
      ;;
  esac
done <<< "$html_files"

if [ "$fail" -eq 1 ]; then
  echo ""
  echo "FAIL: Jekyll built unexpected files. Update _config.yml exclude list."
  exit 1
fi

echo ""
echo "PASS: Jekyll output contains only index.html and reports/*.html"
