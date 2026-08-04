# {{TITLE}} — Session Handoff

> **Lifecycle:** {{LIFECYCLE}} · **Created:** {{DATE}} · **From:** {{FROM_HARNESS}} → **To:** {{TO_HARNESS}}
> Delete this file once the receiving agent has picked up the work. *(temp mode only — remove this line when persisting)*

## Objective

{{OBJECTIVE}}

*One paragraph. What the next session must accomplish, in the user's terms.
Seeded from the invocation arguments when present.*

## Repository State

| Field | Value |
|-------|-------|
| Repo | `{{REPO_ROOT}}` |
| Remote | `{{REMOTE_URL}}` |
| Branch | `{{BRANCH}}` |
| HEAD | `{{HEAD_SHA}}` |
| Uncommitted files | {{DIRTY_FILES}} |
| Unpushed commits | {{UNPUSHED_COMMITS}} |

{{WORKING_TREE_NOTE}}

## What Is Already Done

- **[verified]** {{done_and_proven}} — evidence: {{command_or_output_reference}}
- **[assumed]** {{done_but_unproven}} — not yet exercised

*Label every line `[verified]` or `[assumed]`. An unlabelled claim is the
single most expensive thing a handoff can carry.*

## Next Steps

1. {{first_action}} — file: `{{path}}:{{line}}`
2. {{second_action}}
3. {{third_action}}

*Ordered, each independently startable, each naming the file it touches.*

## Environment & Commands

```bash
# build / run
{{BUILD_CMD}}
# test (the exact command that must pass before this is done)
{{TEST_CMD}}
# lint / typecheck
{{LINT_CMD}}
```

{{ENV_PREREQS}}

## Artifacts — Read These, Don't Re-Derive

| Artifact | Location | Why it matters |
|----------|----------|----------------|
| {{name}} | `{{path_or_url}}` | {{one_line}} |

*Reference specs, plans, ADRs, issues, PRs, and diffs by path or URL.
Never paste their contents into this document.*

## Access & Secrets

{{SECRETS_SECTION}}

*Retrieval instructions only — never values. If nothing is needed, write
"No credentials required for this task."*

## Suggested Skills

| Skill | Use it for | Install |
|-------|------------|---------|
| `{{skill}}` | {{when_in_this_task}} | `humblskills install {{skill}}` |

*Only skills that are genuinely relevant to the next steps above, verified
against what is installed. No speculative list.*

## Open Questions

- {{question}} — blocks step {{n}}. Ask the user before proceeding.

## Gotchas & Dead Ends

- {{approach_tried}} → failed because {{reason}}. Don't retry it.

## Pickup Prompt

```
{{PICKUP_PROMPT}}
```

*One paste-ready paragraph the user can hand to the receiving agent, pointing
at this file's absolute path.*
