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

(no entries yet)

### 2026-10-14 | Separate personal task configuration from the shared package
- Context: Two clients must resume the same work without shipping the author's account references to other users.
- Options: (A) bake account IDs into the skill, (B) duplicate per-client configurations, (C) keep one external personal config and share one skill through client links.
- Chose: C - external personal configuration and a shared skill directory.
- Why: Skill updates and public promotion cannot carry real task IDs; both clients resolve the same settings.
- Result: Configuration validation, scope selection, and permission checks pass in isolated tests.

### 2026-10-14 | Provision Notion through the connected agent tool
- Context: A setup helper can prepare a task-list schema but cannot directly invoke the host's MCP catalog.
- Options: (A) collect a separate API credential, (B) prepare a request for the connected tool, (C) require manual database creation.
- Chose: B - agent-mediated provisioning after an explicit create choice.
- Why: Reuses the user's authenticated connection and returned source ID without collecting tokens or creating duplicate task lists.
- Result: Creation request is covered by offline checks; actual database provisioning remains untested.

### 2026-10-14 | Keep strict pickup scope and repair the producer
- Context: The persisted handoff helper saved at the Git root, which was invisible to pickup in a nested workstream.
- Options: (A) search ancestors automatically, (B) change the producer's destination, (C) select a cross-workstream fallback silently.
- Chose: B - align the producer with the consumer and keep pickup folder-local.
- Why: Automatic ancestor/sibling selection mixes distinct work; explicit-path transfers cover exceptions.
- Result: Nested Git and non-Git round trips, explicit session-start directory, and workstream ignore checks pass.

### 2026-10-14 | Use one task profile for handoff writes and pickup reads
- Context: Handoff task changes must appear in the next pickup without keeping separate client or skill settings.
- Options: (A) separate profiles, (B) one external profile and helper context, (C) embed task IDs in the skill package.
- Chose: B - handoff loads pickup's existing profile and scope; pickup reconciles task links with live status.
- Why: Keeps private account data outside the package and prevents stale snapshots from recreating completed work.
- Result: A connected Notion smoke test verified scoped task creation, verified completion, duplicate reuse, and the next pickup showing only the remaining open task.
