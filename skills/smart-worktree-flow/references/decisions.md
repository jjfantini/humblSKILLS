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

### 2026-09-01 | second brew formula + first-class profile channel
- Context: Jennings wants a pre-release *install* path, then made the beta vs stable channel first-class in the CLI and the existing TUI — not a hidden profile.json key.
- Options: (A) mutate `Formula/humblskills.rb` on develop, (B) `humblskills@beta` versioned formula, (C) second formula `humblskills-pre` plus one `channel` field on the existing profile, wired to `profile get`/`set`, `upgrade --channel`, and the Profile TUI.
- Chose: C.
- Why: A would make `brew upgrade humblskills` jump to `-pre.N`. B is an illegal Homebrew class (`@` only maps with digits). One profile field is the source of truth Homebrew, `upgrade`, and the existing settings TUI already share — no extra config file, no second TUI.
- Result: TBD after first `humblskills-pre` tap commit.

### 2026-09-01 | release-please pre channel, not a second goreleaser workflow
- Context: Jennings wants develop to cut GitHub pre-releases and main to cut the real version + brew. Two honest implementations exist.
- Options: (A) release-please `prerelease` + `versioning: prerelease` + `prerelease-type: pre` on develop, same goreleaser job, `skip_upload: auto`, (B) a dedicated develop workflow that invents the next `-pre` tag and calls goreleaser itself.
- Chose: A.
- Why: release-please already documents pre-release branches that graduate on merge to main. goreleaser already consumes a semver `-pre.N` suffix (`prerelease: auto`, `skip_upload: auto`). B would add a version-computation script and a second source of truth. Tags `vX.Y.Z-pre.N` cannot equal `vX.Y.Z`, GitHub `/releases/latest` and `go install @latest` ignore prereleases, and brew stays on the last main formula. No develop Homebrew channel.
- Result: TBD after first develop pre-release lands.

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
