# Log

Append-only session log. Every session MUST append at least one entry. Never
edit old entries - they are the historical record. Most recent entries appear
at the bottom.

Entry shape:

```markdown
[INGEST|QUERY|LINT <YYYY-MM-DD>] <one-line summary>
  <optional indented detail line(s)>
```

---

[INGEST 2026-06-12] Scaffolded smart-worktree-flow.
  - Added brain meta files, raw user request, workflow wiki concepts, and scripts
  - Initial defaults: worktree isolation, Vibe mode on deferral, cleanup enabled

[LINT 2026-06-12] 6 wiki, 1 raw. Hard: 0, Soft: 0. Regenerated _index.md.

[LINT 2026-07-30] 6 wiki, 1 raw. Hard: 0, Soft: 0. Regenerated _index.md.

[LINT 2026-07-30] 6 wiki, 1 raw. Hard: 0, Soft: 0. Regenerated _index.md.

[LINT 2026-07-30] 6 wiki, 1 raw. Hard: 0, Soft: 0. Regenerated _index.md.

[LINT 2026-07-30] 6 wiki, 1 raw. Hard: 0, Soft: 0. Regenerated _index.md.

[LINT 2026-08-04] 6 wiki, 1 raw. Hard: 0, Soft: 0. Regenerated _index.md.

[LINT 2026-08-04] 6 wiki, 1 raw. Hard: 0, Soft: 0. Regenerated _index.md.

[QUERY 2026-09-01] Two-branch release path (develop pre, main stable + brew).
  - Updated SKILL.md steps 4–6, release-pr.md, main-gate.md, version 1.2.0

[QUERY 2026-09-01] Ground truth: skill was main-only + brew post-check; Jennings wants develop pre-releases.
  - Chose release-please pre channel over a second goreleaser workflow
  - Vibe still auto-merges develop→main and both release PRs; brew post-check after main only

[LINT 2026-09-01] 6 wiki, 1 raw. Hard: 0, Soft: 0. Regenerated _index.md.

[LINT 2026-09-01] 6 wiki, 1 raw. Hard: 0, Soft: 0. Regenerated _index.md.

[LINT 2026-09-01] 6 wiki, 1 raw. Hard: 0, Soft: 0. Regenerated _index.md.

[QUERY 2026-09-01] First-class beta channel: second brew formula + profile field.
  - humblskills-pre on pre tags only; stable formula never jumps to -pre.N
  - Same profile.channel for profile get/set, TUI install channel, upgrade --channel
  - Skill version 1.2.1

[QUERY 2026-09-01] Shared manifest blocked develop→main stable.
  - run 33548192550 skipped: last release matched manifest 2.52.0-pre
  - Split manifests; promote merge waits for the stable release PR
  - Skill version 1.2.2

[QUERY 2026-09-01] Smoking gun 33548192550: last-saw 375eb41; #270 unparsed.
  - Document exact unstick: promote fix to main (that push is the run)
  - Do not re-run release.yml on current main; Release-As is fallback only

[LINT 2026-09-01] 6 wiki, 1 raw. Hard: 0, Soft: 0. Regenerated _index.md.

[LINT 2026-09-01] 6 wiki, 1 raw. Hard: 0, Soft: 0. Regenerated _index.md.

[QUERY 2026-09-01] After v2.52.0, develop must not increment 2.52.0-pre.N.
  - 2.52.0-pre.3 < 2.52.0 so beta stays on stable; next pre is 2.52.1-pre.1
  - Reset develop manifest to graduated stable; Release-As only for this cut
  - Skill version 1.2.3

[LINT 2026-09-01] 6 wiki, 1 raw. Hard: 0, Soft: 0. Regenerated _index.md.
