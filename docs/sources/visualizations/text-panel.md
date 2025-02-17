---
aliases:
  - /docs/grafana/latest/features/panels/text/
  - /docs/grafana/latest/panels/visualizations/text-panel/
  - /docs/grafana/latest/reference/alertlist/
  - /docs/grafana/latest/visualizations/text-panel/
keywords:
  - grafana
  - text
  - documentation
  - panel
title: Text
weight: 1100
---

# Text

The text panel visualization lets you make information and description panels for your dashboards.

## Mode

**Mode** determines how embedded content appears.

### Markdown

This option formats the content as [markdown](https://en.wikipedia.org/wiki/Markdown).

### HTML

This setting renders the content as [sanitized](https://github.com/grafana/grafana/blob/code-in-text-panel/packages/grafana-data/src/text/sanitize.ts) HTML. If you require more direct control over the output, you can set the
[disable_sanitize_html]({{< relref "../setup-grafana/configure-grafana/#disable_sanitize_html" >}}) flag which enables you to directly enter HTML.

### Code

This setting renders content inside a read-only code editor. Select an appropriate language to apply syntax highlighting
to the embedded text.

## Variables

[Variables]({{< relref "../dashboards/variables/variable-syntax" >}}) in the content will be expanded for display.
