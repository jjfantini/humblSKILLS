# Log

Append-only session log. Every session MUST append at least one entry.
Never edit old entries - they are the historical record. Most recent
entries appear at the bottom.

Entry shape:

```
[INGEST|QUERY|LINT <YYYY-MM-DD>] <one-line summary>
  <optional indented detail line(s)>
```

---

[INGEST 2026-04-20] Scaffolded create-video-transition manually.
  - Target path /Users/jjfantini/github/humblSKILLS/skills/create-video-transition/ is outside scripts/scaffold.sh's supported --location flags (personal|project), so files were copied from skills/use-smart-skill/ by hand.
  - 12 wiki concepts authored from scratch (sources: [] — flagged by lint as synthesis orphans, acceptable at launch).
  - HTML template forked from ~/.cursor/skills/scroll-stop-prompter/assets/prompt-page-template.html with: gear-icon settings modal (VIDEO_LENGTH / VIDEO_ASPECT_RATIO / VIDEO_AUDIO / VIDEO_FPS), "Provided by user" placeholder tab, Start/End/Video tab labels.

[LINT 2026-04-20] 12 wiki, 0 raw. Hard: 0, Soft: 12. Regenerated _index.md.

[INGEST 2026-04-20] Scaffold complete — skill ready to ship.
  - SKILL.md (< 200 lines) passes trigger/WHAT/WHEN criteria; description includes 4 trigger phrases and a "Do NOT use" negative trigger.
  - 12 wiki concepts across 3 contexts (prompt, workflow, output) with frontmatter triples matching filesystem paths.
  - HTML template forked at assets/prompt-page-template.html with settings modal, gear icon, and "Provided by user" placeholder for skipped tabs.
  - Remaining work for future sessions: smoke-test HTML in a browser (manual), then drop real-world Prompt A/B/C samples into references/raw/ and convert synthesis orphans into source-cited concepts as patterns emerge.


[QUERY 2026-04-21] Extracted Prompt A/B/C templates from wiki concepts into assets/templates/.
  - New files: assets/templates/prompt-a-start-frame.tmpl, prompt-b-end-frame.tmpl, prompt-c-block-timeline.tmpl (raw fill-in-the-blank templates)
  - Wiki concepts rewritten as thin guidance layers: purpose + rules + worked example + incorrect/correct, with a Template section that routes to the asset path
  - Matches the create-scroll-animation pattern (templates in assets/, conceptual rules in wiki/)
  - SKILL.md routes unchanged — concepts still live at the same wiki paths; they now route onward to the asset files
  - Block-count table + per-block rules retained in block-timeline.md (they're rules, not template)

[LINT 2026-04-21] 12 wiki, 0 raw. Hard: 0, Soft: 12. Regenerated _index.md.
