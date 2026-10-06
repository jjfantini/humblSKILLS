# Patterns

Performance memory. Each entry records a concrete attempt, its numeric
outcome, and the lesson. Read before every session; append after every
session where quantified results appear.

Entry shape (see `wiki/brain/patterns/how-to-log-results.md` for the full
worked example):

```
### <YYYY-MM-DD> | <short title>
- Context: <what was attempted, in one line>
- Approach: <the method used>
- Result: <metrics, numbers, outcomes>
- Worked: <what helped>
- Didn't: <what hurt>
- Lesson: <the rule to apply next time>
```

---

### 2026-10-02 | Calibrating the promotion gate on every published skill
- Context: tuning `scripts/promote.sh`'s gate so it blocks real problems without blocking good skills.
- Approach: ran `promote.py check --to humblskills` over all 23 skills in skills/, read every FAIL and WARN, fixed false positives, and compared the stdlib frontmatter parser with registry.json (Go's parse).
- Result: parser matched registry.json on 23/23 skills. First pass failed 6/23; requiring key material after a private-key header (smart-handoff's redaction fixture quotes the bare header) took credential false positives from 1 to 0, leaving 5/23 failures, all real: smart-animation description 1120 > 1024 chars, smart-watch has '<' in metadata.dependencies, smart-humanize-text has no tags, smart-orchestrate and smart-skill mention HappyRobot.
- Worked: calibrating on real skills before shipping; token-level placeholder hints (example, xxxx, your) for credential patterns.
- Didn't: a bare BEGIN PRIVATE KEY regex - fixtures and docs quote the header.
- Lesson: before adding a gate rule, run it over every published skill; a rule that fails a good skill teaches people to reach for overrides.
