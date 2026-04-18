# Decisions

Reasoning memory. Each entry records a non-obvious choice: the context,
the options considered, what was chosen, why, and the observed result.
Never delete entries - if a decision is reversed, add a new entry that
references the old one.

Entry shape:

```
### <YYYY-MM-DD> | <short title>
- Context: <the situation that required a choice>
- Options: (A) <opt>, (B) <opt>, (C) <opt>
- Chose: <letter and name>
- Why: <the rationale, ideally citing evidence>
- Result: <what happened after, or "TBD">
```

---

### 2026-04-17 | Add `compatibility` frontmatter, skip `allowed-tools`
- Context: agentskills.io spec defines 4 optional SKILL.md frontmatter fields (`license`, `compatibility`, `metadata`, `allowed-tools`). Needed to decide which to propagate through this skill and its scaffold.
- Options: (A) add both `compatibility` and `allowed-tools`, (B) add `compatibility` only, (C) add neither and only document in a wiki concept.
- Chose: B - add `compatibility` to this skill's SKILL.md, emit a loud TODO placeholder in scaffold.sh, and propagate through workflow/validation docs. Skip `allowed-tools` entirely.
- Why: `compatibility` carries real signal (this skill has bash scripts, consumers need to know). Max 500 chars, cheap, per-spec appropriate. `allowed-tools` is flagged experimental by the spec itself, agent support varies, and this is a meta-skill with open-ended tool use - pre-approving a list would either be decorative (too permissive) or restrictive (breaks the skill). Scaffold emits a TODO-or-delete line (not silent omit) so new-skill authors consciously decide instead of forgetting the field exists.
- Result: SKILL.md now declares bash + POSIX requirements. Scaffold forces explicit decision. Future skill authors see the field in the new `wiki/smart/spec/skill-frontmatter.md` concept.
