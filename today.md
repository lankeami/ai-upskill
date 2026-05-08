---
permalink: /today
---
{% assign reports = site.pages | where_exp: "item", "item.path contains 'reports/'" | sort: "name" | last %}
<meta http-equiv="refresh" content="0; url={{ site.baseurl }}{{ reports.url }}">
<script>window.location.replace("{{ site.baseurl }}{{ reports.url }}");</script>
<p>Redirecting to latest report... <a href="{{ site.baseurl }}{{ reports.url }}">click here</a> if not redirected.</p>
