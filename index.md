---
layout: default
title: AI Daily Report
---

# AI Daily Report

A daily aggregation of AI news from Reddit, Hacker News, and tech RSS feeds, organized by company. Reports are generated automatically each day.

## Reports

{% assign report_pages = site.pages | where_exp: "p", "p.path contains 'reports/'" | where_exp: "p", "p.date" | sort: "date" | reverse %}

{% for report in report_pages %}
{% assign company_count = report.companies | size %}
{% if company_count > 3 %}
  {% assign shown = report.companies | slice: 0, 3 | join: ", " %}
  {% assign remaining = company_count | minus: 3 %}
- [{{ report.date | date: "%B %d, %Y" }}]({{ report.url | relative_url }}) — {{ shown }} + {{ remaining }} more — {{ report.item_count }} items
{% else %}
- [{{ report.date | date: "%B %d, %Y" }}]({{ report.url | relative_url }}) — {{ report.companies | join: ", " }} — {{ report.item_count }} items
{% endif %}
{% endfor %}
