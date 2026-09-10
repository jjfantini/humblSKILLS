# Patterns

Performance memory. Each entry records a concrete attempt, its numeric
outcome, and the lesson. Read before every session; append after every session
where quantified results appear.

Entry shape:

```markdown
### <YYYY-MM-DD> | <short title>
- Context: <what was attempted, in one line>
- Approach: <the method used>
- Result: <metrics, numbers, outcomes>
- Worked: <what helped>
- Didn't: <what hurt>
- Lesson: <the rule to apply next time>
```

---

### 2026-09-01 | shared manifest skipped stable after promote
- Context: develop→main after `v2.52.0-pre`; expected `v2.52.0` + brew.
- Approach: one `.release-please-manifest.json` holding `2.52.0-pre` on both branches; main `prerelease: false`.
- Result: release.yml run 33548192550 skipped; 1 unparsed merge commit after the pre tag; 0 stable tags; Homebrew stayed 2.51.0.
- Worked: n/a (path failed).
- Didn't: treating the pre as last-released on main; expecting `prerelease: false` to graduate with an empty changelog.
- Lesson: keep last-stable and last-pre in separate manifests. Promote merge is not the tag.
