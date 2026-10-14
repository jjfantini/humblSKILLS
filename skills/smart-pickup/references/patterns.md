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

(no entries yet)

### 2026-10-14 | Test selection and configuration without external writes
- Context: Validate workstream isolation and configuration safety before installing the shared skill.
- Approach: Run isolated standard-library tests with temporary handoffs and private configs.
- Result: 15 tests pass; package gate reports 0 failures and 0 warnings.
- Worked: Exact directory-segment mapping, explicit-file handling, and atomic config saves isolate user state.
- Didn't: The first fixture compared logical temporary paths against resolved paths; canonicalizing the fixture fixed 6 assertion failures. New Notion provisioning and conversational trigger rates are not measured.
- Lesson: Resolve fixture paths on systems with symlinked temporary directories, and keep provisioning claims separate from offline request tests.

### 2026-10-14 | Verify real Notion setup and producer-to-consumer pickup
- Context: Validate new task-list provisioning and handoff discovery in tandem.
- Approach: Create a private test database through the connected tool, fetch the returned schema, query open tasks, save an isolated profile, and exercise directory fixtures.
- Result: 1 database created; 5 properties and 5 status states verified; 0 open test tasks; 19 offline pickup/integration tests and 22 handoff checks pass.
- Worked: Use returned data source IDs outside the package and the same starting-folder path for both skills.
- Didn't: Prior setup only tested the prepared request. Conversational selection rates across fresh clients are still not measured.
- Lesson: Verify the complete create-fetch-query-configure-load sequence while preserving the main personal config.
