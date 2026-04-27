# Published eval reports

Single-file HTML dashboards published from `humblskills eval` runs. Each report is self-contained (inline Plotly) and opens standalone. Every published report is referenced by a dated filename — there is no `latest-*` alias. To browse the per-skill scenario pages with live previews, use the sidebar (e.g. [`use-smart-humanize-text` reports](use-smart-humanize-text/index.md)).

## Reports index

| Date       | Showcase                         | Skill                       | Runner        | Link                                                                                                          |
|------------|----------------------------------|-----------------------------|---------------|---------------------------------------------------------------------------------------------------------------|
| 2026-04-20 | adaptive-brand-voice-discovery   | use-smart-humanize-text     | cursor-agent  | [HTML](use-smart-humanize-text/adaptive-brand-voice-discovery-2026-04-20.html) · [scenario page](use-smart-humanize-text/adaptive-brand-voice-discovery.md) |
| 2026-04-27 | indie-launch-copy-iteration      | use-smart-humanize-text     | cursor-agent  | [HTML](use-smart-humanize-text/indie-launch-copy-iteration-2026-04-27.html) · [scenario page](use-smart-humanize-text/indie-launch-copy-iteration.md)       |

## Reproducing locally

```sh
humblskills eval brand-voice
# equivalent long form:
humblskills eval run use-smart-humanize-text --scenario adaptive-brand-voice-discovery --open

# indie-launch:
humblskills eval run use-smart-humanize-text --scenario indie-launch-copy-iteration --runner cursor-agent
```

See [Scenarios](../scenarios.md) for scenario design notes.

## Publishing a new report

1. Run the eval: `humblskills eval brand-voice` (or any scenario).
2. Locate the freshly-generated `report.html` (path is printed after the run; also visible via `humblskills eval ls <skill>`).
3. Copy into this directory with a date-stamped name: `cp report.html docs/eval/reports/<showcase>-YYYY-MM-DD.html`.
4. Add a row to the table above with the dated link, and update the relevant scenario page under `use-smart-humanize-text/` to point its iframe at the new dated filename.
5. Open a PR — GitHub Actions will publish the new HTML file to GitHub Pages under `https://jjfantini.github.io/humblSKILLS/eval/reports/`.
