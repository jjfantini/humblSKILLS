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

### 2026-06-12 | Default to worktrees and Vibe mode on deferral
- Context: The skill needs safe defaults when the user says "I defer to you" but still must avoid clobbering parallel local work.
- Options: (A) in-place branches by default, (B) worktrees by default, (C) always stop until the user chooses.
- Chose: B - worktrees by default, with Vibe mode autonomy unless the user asks for HITL.
- Why: Worktrees isolate agent work from dirty local branches and parallel Codex, Claude, or Cursor sessions. Vibe mode matches the requested autonomous flow while still requiring green tests, lint, verification, and CI/CD before merges.
- Result: TBD after first real use.
