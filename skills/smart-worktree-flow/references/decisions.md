# Decisions

Reasoning memory. Each entry records a non-obvious choice: the context, the
options considered, what was chosen, why, and the observed result.

Entry shape:

```markdown
### <YYYY-MM-DD> | <short title>
- Context: <the situation that required a choice>
- Options: (A) <opt>, (B) <opt>, (C) <opt>
- Chose: <letter and name>
- Why: <the rationale>
- Result: <what happened after, or "TBD">
```

---

### 2026-09-01 | after stable graduate, reset develop last-released to that stable
- Context: main tagged `v2.52.0`. develop kept incrementing `2.52.0-pre.N` (`v2.52.0-pre.3` in #278). Beta = max(stable, pre). Semver: `2.52.0-pre.3` < `2.52.0`, so users on channel=beta already on 2.52.0 never see the update-notice banner. Jennings wants the next pre to be `2.52.1-pre.1` (patch, #271's feat already consumed as pre.3) without hand-tagging or pushing main.
- Options: (A) change `versioning` from `prerelease` to `default` so every develop commit bumps the base (loses intra-cycle `2.52.1-pre.2`), (B) one-off `Release-As: 2.52.1-pre.1` only, (C) keep `versioning: prerelease` and after each stable rewrite the develop manifest from `X.Y.Z-pre.N` to `X.Y.Z` so the next feat/fix starts a new pre line.
- Chose: C, plus a one-time `Release-As: 2.52.1-pre.1` because leftover feats on develop since `v2.52.0` would otherwise open `2.53.0-pre.1`.
- Why: `PrereleasePatchVersionUpdate` on a non-prerelease `2.52.0` yields `2.52.1-pre`. On `2.52.0-pre.3` it yields `2.52.0-pre.4`. The split manifests already made last-stable and last-pre independent; the missing rule was "last-pre base must move past last-stable after a graduate." A/B stay one-off. C is the default: `record_stable_on_develop` in `release.yml` runs `scripts/sync-develop-pre-after-stable.sh` after a non-`-pre` tag. Intra-cycle `2.52.1-pre.2` is unchanged.
- Result: TBD after `chore(develop): release 2.52.1-pre.1` tags `v2.52.1-pre.1` (prerelease, `humblskills-pre` only).

### 2026-09-01 | second brew formula + first-class profile channel
- Context: Jennings wants a pre-release *install* path, then made the beta vs stable channel first-class in the CLI and the existing TUI — not a hidden profile.json key.
- Options: (A) mutate `Formula/humblskills.rb` on develop, (B) `humblskills@beta` versioned formula, (C) second formula `humblskills-pre` plus one `channel` field on the existing profile, wired to `profile get`/`set`, `upgrade --channel`, and the Profile TUI.
- Chose: C.
- Why: A would make `brew upgrade humblskills` jump to `-pre.N`. B is an illegal Homebrew class (`@` only maps with digits). One profile field is the source of truth Homebrew, `upgrade`, and the existing settings TUI already share — no extra config file, no second TUI.
- Result: TBD after first `humblskills-pre` tap commit.

### 2026-09-01 | split manifests; do not share last-released version
- Context: #270 merged develop→main after `v2.52.0-pre` existed. Shared `.release-please-manifest.json` was `{ ".": "2.52.0-pre" }`. release.yml run 33548192550 on main skipped: "No user facing commits found since v2.52.0-pre". No `v2.52.0`, brew still 2.51.0.
- Options: (A) keep one manifest and rely on `prerelease: false` / `versioning: prerelease` on main to graduate, (B) `last-release-sha` or `Release-As: 2.52.0` on every promote, (C) a job that invents `vX.Y.Z`, (D) two manifests — develop records last pre, main records last stable.
- Chose: D.
- Why: release-please matches GitHub releases to the manifest version, then walks commits *after that SHA* (`src/manifest.ts` `expectedVersion === tagName.version`, then `changelogEmpty` in `src/strategies/base.ts`). The official "toggle `prerelease`" pairing never reaches versioning when the latest matching release is the `-pre` tag and the only later commit is `Merge pull request #N` (not a conventional commit). A/B are sticky glue; C is a second source of truth. D deletes the bad assumption that one last-released version can describe both channels.
- Result: TBD after this reaches main and `release.yml` opens `v2.52.0` from the existing `2.51.0` stable marker.

### 2026-09-01 | release-please pre channel, not a second goreleaser workflow
- Context: Jennings wants develop to cut GitHub pre-releases and main to cut the real version + brew. Two honest implementations exist.
- Options: (A) release-please `prerelease` + `versioning: prerelease` + `prerelease-type: pre` on develop, same goreleaser job, `skip_upload: auto`, (B) a dedicated develop workflow that invents the next `-pre` tag and calls goreleaser itself.
- Chose: A.
- Why: release-please already documents pre-release branches that graduate on merge to main. goreleaser already consumes a semver `-pre.N` suffix (`prerelease: auto`, `skip_upload: auto`). B would add a version-computation script and a second source of truth. Tags `vX.Y.Z-pre.N` cannot equal `vX.Y.Z`, GitHub `/releases/latest` and `go install @latest` ignore prereleases, and brew stays on the last main formula. No develop Homebrew channel.
- Result: First develop pre (`v2.52.0-pre`) landed. Graduation on main failed until manifests were split (see 2026-09-01 split-manifests entry).

### 2026-09-01 | Two-branch release: develop is pre, main is stable + brew
- Context: The live repo watched only `main` (`release.yml` `on.push.branches: [main]`). The canonical skill described feature→develop→main plus a third release-please PR on main, then `brew upgrade` as a post-check — not GitHub prereleases on develop.
- Options: (A) keep documenting that single-release-branch path, (B) document develop pre-release (`vX.Y.Z-pre.N`) and main stable + Homebrew as the new path, keep brew as a post-check after main only.
- Chose: B - match `.github/workflows/release.yml` and `release-please-config.develop.json`.
- Why: Jennings' new ask is develop → pre-release, main → real version + tap. The skill must not tell agents develop does not release, or that brew follows develop.
- Result: TBD after first develop pre-release lands.

### 2026-06-12 | Default to worktrees and Vibe mode on deferral
- Context: The skill needs safe defaults when the user says "I defer to you" but still must avoid clobbering parallel local work.
- Options: (A) in-place branches by default, (B) worktrees by default, (C) always stop until the user chooses.
- Chose: B - worktrees by default, with Vibe mode autonomy unless the user asks for HITL.
- Why: Worktrees isolate agent work from dirty local branches and parallel Codex, Claude, or Cursor sessions. Vibe mode matches the requested autonomous flow while still requiring green tests, lint, verification, and CI/CD before merges.
- Result: TBD after first real use.
