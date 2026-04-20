# Published eval reports

Single-file HTML dashboards published from `humblskills eval` runs. Each report is self-contained (inline Plotly) and opens standalone. The `latest-*.html` files track the most recent publication per showcase and are safe to link to from blog posts or announcements.

## Latest: adaptive-brand-voice-discovery (2026-04-20)

The canonical Smart Skills compounding-learning showcase. Six sessions, three arms, ten idiosyncratic brand-voice rules for a fictional Toronto fintech.

**Headline numbers (cursor-agent, fresh run):**

| Arm          | pass_rate  | tokens (mean) |
|--------------|:----------:|:-------------:|
| smart_skill  | **0.935**  | 63,519        |
| flat_skill   | 0.679      | 72,266        |
| no_skill     | 0.740      | 193,785       |

- **smart vs flat**: +0.256 pass_rate (**+37.7%**), −12.1% tokens
- **smart vs none**: +0.194 pass_rate (**+26.3%**), −67.2% tokens
- Session 5 (pure retention, no in-prompt feedback): **smart = 0 violations, flat = 10, no_skill = 9**

[**→ Open the full interactive report in a new tab**](adaptive-brand-voice-discovery-2026-04-20.html){ target="_blank" }

Or [view the "latest" alias](latest-brand-voice.html){ target="_blank" } (always points at the most recent brand-voice publication).

### Live preview

<iframe src="adaptive-brand-voice-discovery-2026-04-20.html" width="100%" height="900" style="border: 1px solid #24262f; border-radius: 8px; background: #0c0d12;" loading="lazy"></iframe>

## Reports index

| Date       | Showcase                         | Skill                       | Runner        | Link                                                                                                              |
|------------|----------------------------------|-----------------------------|---------------|-------------------------------------------------------------------------------------------------------------------|
| 2026-04-20 | adaptive-brand-voice-discovery   | use-smart-humanize-text     | cursor-agent  | [dated](adaptive-brand-voice-discovery-2026-04-20.html) · [latest](latest-brand-voice.html)                       |

## Reproducing locally

```sh
humblskills eval brand-voice
# equivalent long form:
humblskills eval run use-smart-humanize-text --scenario adaptive-brand-voice-discovery --open
```

`humblskills eval brand-voice` runs the three-arm showcase on `use-smart-humanize-text` and opens the generated report in your browser. See [Scenarios](../scenarios.md) for scenario design notes.

## Publishing a new report

1. Run the eval: `humblskills eval brand-voice` (or any scenario).
2. Locate the freshly-generated `report.html` (path is printed after the run; also visible via `humblskills eval ls <skill>`).
3. Copy into this directory with a date-stamped name: `cp report.html docs/eval/reports/<showcase>-YYYY-MM-DD.html`.
4. Update the `latest-<showcase>.html` symlink: `(cd docs/eval/reports && ln -sf <dated-file>.html latest-<showcase>.html)`.
5. Add a row to the table above with the headline numbers.
6. Open a PR — GitHub Actions will publish the new HTML file to GitHub Pages under `https://jjfantini.github.io/humblSKILLS/eval/reports/`.
