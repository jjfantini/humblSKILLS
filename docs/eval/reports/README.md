# Published eval reports

Single-file HTML dashboards published from `humblskills eval` runs. Each report is self-contained (inline Plotly) and opens standalone. The `latest-*.html` symlinks track the most recent publication per showcase and are safe to link to from docs or blog posts.

## Reports

| Date       | Showcase                         | Skill                       | Runner        | Link                                                                                 | Headline                                                                                                   |
|------------|----------------------------------|-----------------------------|---------------|--------------------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------|
| 2026-04-20 | adaptive-brand-voice-discovery   | use-smart-humanize-text     | cursor-agent  | [dated](adaptive-brand-voice-discovery-2026-04-20.html) · [latest](latest-brand-voice.html) | smart_skill pass_rate **0.935** vs no_skill 0.740 (**+26.3%**) and flat_skill 0.679 (**+37.7%**); smart uses **67% fewer tokens** than no_skill. |

## Reproducing

```sh
humblskills eval brand-voice
# or:
humblskills eval run use-smart-humanize-text --scenario adaptive-brand-voice-discovery --open
```

`humblskills eval brand-voice` runs the canonical three-arm showcase on `use-smart-humanize-text` and opens the generated report in your browser. See [../scenarios.md](../scenarios.md) for scenario design notes.

## Publishing a new report

1. Run the eval: `humblskills eval brand-voice` (or any scenario).
2. Locate the freshly-generated `report.html` (path is printed after the run; also visible via `humblskills eval ls <skill>`).
3. Copy into this directory with a date-stamped name: `cp report.html docs/eval/reports/<showcase>-YYYY-MM-DD.html`.
4. Update the `latest-<showcase>.html` symlink: `(cd docs/eval/reports && ln -sf <dated-file>.html latest-<showcase>.html)`.
5. Add a row to the table above with the headline numbers.
