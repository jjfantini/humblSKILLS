---
name: smart-handoff
description: >
  Capture a running agentic session - objective, verified vs assumed progress,
  next steps, artifacts, gotchas, secrets access - into one portable handoff
  markdown file another agent, harness, or agent team can pick up cold.
  Redacts credentials, suggests the skills the receiving agent should install,
  and references specs, plans, ADRs, issues and diffs by path instead of
  copying them. Use when the user says "hand this off", "handoff", "handover",
  "write a handoff doc", "pass this to codex", "moving to cursor", "continue
  this in Claude", "another agent is taking over", "context dump for the next
  session", or is about to switch harness, model, or teammate mid-task. Do NOT
  use for transferring ownership of a managed worktree, for coordinating
  agents you keep supervising yourself, or for authoring commits or PR
  descriptions.
license: MIT
compatibility: Requires bash, git (optional but recommended), POSIX utilities (grep, sed, find), python3 for scripts/lint.sh, and a writable /tmp.
allowed-tools: "Bash(bash:*) Bash(git:*) Bash(mkdir:*) Bash(humblskills:*) Read Write Edit Glob Grep"
metadata:
  author: jjfantini
  version: "1.0.0"
  category: meta
  tags: [handoff, context-transfer, session, agent-interop, codex, cursor, humblskill]
  platforms: [claude-code, cursor, codex]
  preserve:
    - references/raw/
    - references/wiki/
    - references/decisions.md
    - references/log.md
    - references/patterns.md
---

# Smart Handoff

Turn a live session into one document a cold agent can act on. The receiving
agent may be Claude, Codex, Cursor, or a team of them — so the doc carries
**intent, state, and access instructions**, and links to everything else.

## Brain Protocol (read BEFORE creating anything)

1. `references/_index.md`       - what this skill knows (map)
2. `references/patterns.md`     - what worked, with numbers
3. `references/decisions.md`    - past reasoning, don't repeat mistakes
4. `references/log.md`          - last 5 session entries
5. Relevant `references/wiki/handoff/<category>/` concepts per task

After completing work, UPDATE the brain:
- Performance data (if reported) -> `patterns.md`
- Non-obvious decisions -> `decisions.md`
- Session summary (always) -> append to `log.md`
- Run `scripts/lint.sh` to regenerate `_index.md` and verify structure

_Full spec: `references/_brain.md`._

## Workflow

| Step | What | Driver |
|------|------|--------|
| 1 | Seed from the arguments | Agent |
| 2 | Ask: temp or persist? | Agent (one question) |
| 3 | Infer the receiving harness | Agent (ask only if no signal) |
| 4 | Snapshot repo + skills | `scripts/preflight.sh` |
| 5 | Draft the doc | Agent + `assets/handoff-template.md` |
| 6 | Scan for leaked credentials | `scripts/scan-secrets.sh` |
| 7 | Report the path + pickup prompt | Agent |

### Step-by-step

1. **Seed from the arguments.** Anything the user passed is the *goal of the
   next session*, not a title. It drives what you investigate and what lands in
   Objective and Next Steps. "handoff so codex can finish the retry logic"
   means the doc is about the retry logic, tailored to Codex — everything
   unrelated gets one line in Artifacts or nothing at all.

2. **Ask temp or persist** unless the arguments already answered it. Default is
   temp. See `references/wiki/handoff/lifecycle/temp-vs-persist.md`.
   - temp → `/tmp/.humblskills/handoffs/`
   - persist → `<repo>/.humblskills/handoffs/`

3. **Infer the receiving harness** from the arguments and conversation; ask one
   question only if there is no signal. See
   `references/wiki/handoff/targets/harness-tailoring.md`.

4. **Snapshot.** `bash scripts/preflight.sh --slug <descriptive-name>` returns
   the date, both resolved file paths, branch, HEAD, dirty and unpushed counts,
   status, diffstat, recent commits, which instruction files exist, and the
   installed skill list. Read-only — it creates nothing.

5. **Draft.** Copy `assets/handoff-template.md`, fill it, delete genuinely empty
   sections, keep the order. Filename:
   `<descriptive-name>-handoff-<YYYY-MM-DD>.md`. Two rules do most of the work:
   - Label every completion claim `[verified]` (a command was run, cite it) or
     `[assumed]` (written, never exercised).
   - Never restate a spec, plan, ADR, issue, PR, or diff — link it.

   See `references/wiki/handoff/document/structure.md` and
   `references/wiki/handoff/capture/session-context.md`.

6. **Scan.** `bash scripts/scan-secrets.sh <path>` before you tell the user it
   is ready. Exit 1 means a credential is in the file — fix and re-run. If a
   token genuinely must cross and the user has not flagged it, **raise it
   yourself** and offer keychain transit with a delete-after-use instruction.
   See `references/wiki/handoff/security/secret-handling.md`.

7. **Report.** Print the absolute path and the paste-ready Pickup Prompt. In
   temp mode, state that the receiving agent should delete the file when done.

## Non-Negotiables

- **One question about lifecycle, at most one about the target.** Everything
  else is inferred. Do not interview the user.
- **No duplicated artifact content.** Paths and URLs only.
- **No credential values.** Retrieval commands only — `op read`,
  `security find-generic-password`, `gh auth token`.
- **Flag unflagged secrets.** If a token from this session is needed downstream
  and the user has not mentioned it, say so before writing the doc.
- **Transcript digging is opt-in.** Write from live context and git by default.
- **Never edit `.gitignore` on your own** in persist mode — report and let the
  user decide.
- **Under ~150 lines.** Longer means artifact contents leaked in.

## How to Use

**Live enumeration of contexts, categories, and concepts:**
Read `references/_index.md` (auto-regenerated by `scripts/lint.sh`).

**Brain protocol, naming conventions, linking contract, lint checks:**
Read `references/_brain.md`. Wiki concept shape: `references/_template.md`.

### Scripts

- `scripts/preflight.sh` — read-only snapshot: paths, git state, instruction
  files, installed skills. Step 4. `--slug <name>` also resolves both filenames.
- `scripts/scan-secrets.sh` — masked credential/PII sweep. Step 6. Exit 1 on
  any `FAIL`. `--strict` also fails on `WARN`.
- `scripts/lint.sh` — brain health check; regenerates `_index.md`.

### Primary workflows

**What to capture and the `[verified]` / `[assumed]` rule:**
`references/wiki/handoff/capture/session-context.md`

**Section order, file naming, the no-duplication rule:**
`references/wiki/handoff/document/structure.md`

**Suggested Skills from the installed list, with install commands:**
`references/wiki/handoff/document/suggested-skills.md`

**Redaction, keychain/1Password transit, delete-after-use:**
`references/wiki/handoff/security/secret-handling.md`

**temp vs persist paths, cleanup, gitignore posture:**
`references/wiki/handoff/lifecycle/temp-vs-persist.md`

**Tailoring to Claude Code / Codex / Cursor / unknown:**
`references/wiki/handoff/targets/harness-tailoring.md`

## Examples

### Example 1: throwaway handoff to Codex, mid-task

User says: "hand this off so codex can finish the retry logic"

Actions:
1. Target inferred as Codex, subject inferred as the retry work — no question
   about either. Ask only: temp or persist? User says temp.
2. `bash scripts/preflight.sh --slug retry-backoff` → branch `feat/http-retry`,
   3 dirty files, 2 unpushed commits, `AGENTS.md` present, `smart-commit`
   installed.
3. Draft from `assets/handoff-template.md`: Objective = finish 429 handling;
   `[verified]` wrapper at `src/http/client.ts:88` (14/14 tests);
   `[assumed]` `Retry-After` parsing; Next Steps 1-3 each naming a file;
   Artifacts links the open PR and the ADR rather than quoting them; Suggested
   Skills lists `smart-commit` with its install command; instruction file
   referenced as `AGENTS.md`, no slash-command syntax.
4. `bash scripts/scan-secrets.sh /tmp/.humblskills/handoffs/retry-backoff-handoff-2026-08-03.md`
   → 0 FAIL, 1 WARN (a `noreply@` address) → ship.
5. Print the path plus a Pickup Prompt ending in "then delete the file".

Result: one 120-line file; Codex resumes without re-reading the diff, and no
credential or duplicated ADR text left the session.

### Example 2: persisted handoff that needs a token to cross

User says: "write a handover doc and keep it in the repo, the next session
picks up the Linear sync"

Actions:
1. "keep it in the repo" → persist, no lifecycle question. No target named →
   ask once; user says "probably Cursor" → tailor to `.cursor/rules`.
2. `bash scripts/preflight.sh --slug linear-sync` → `PERSIST_FILE=` under
   `<repo>/.humblskills/handoffs/`, `PERSIST_DIR_GITIGNORED=no`.
3. Notice step 3 needs the Linear API token the user pasted earlier and never
   flagged. Raise it: offer `security add-generic-password -s handoff-transit
   -a linear-api` for the user to run in a real terminal, cite the read command
   in the doc, and add the delete-after-use line.
4. `mkdir -p` the persist dir, write the doc with a "persisted — record of…"
   lifecycle line instead of the delete banner, then
   `bash scripts/scan-secrets.sh` → 0 FAIL.
5. Report that `.humblskills/` is not gitignored, so the file will appear in
   `git status` — the user decides whether to commit or ignore it.

Result: durable handoff, the token reachable by one keychain read and scheduled
for deletion, and no unrequested `.gitignore` edit.

## Troubleshooting

**`scan-secrets.sh` exits 1 on a placeholder**
Cause: the placeholder does not match the safe-list (`op://`, `REDACTED`,
`<ANGLE_BRACKETS>`, `$ENV_VAR`, `EXAMPLE`, `example.com`, `noreply@`).
Fix: rewrite the placeholder into one of those forms. Do not add `--strict`
bypasses or delete the check.

**`preflight.sh` prints `REPO_ROOT=none`**
Cause: invoked outside a git repo. Persist mode falls back to `$PWD`.
Fix: confirm the directory with the user and use absolute paths in Artifacts.

**Receiving agent asks questions the doc should have answered**
Cause: Next Steps were written as goals, not actions, or files were unnamed.
Fix: every step names the file it touches and is independently startable.

**Doc is 400 lines**
Cause: artifact contents were pasted instead of linked.
Fix: cut to paths and URLs. See the no-duplication rule in
`references/wiki/handoff/document/structure.md`.

## Success Signals

- The receiving agent's first action is work, not a clarifying question.
- `scripts/scan-secrets.sh` exits 0 on every shipped handoff.
- Every completion claim carries `[verified]` or `[assumed]`.
- Every suggested skill appears in `preflight.sh`'s `installed-skills` block.
- Doc stays under ~150 lines with zero pasted artifact contents.
- At most two questions asked of the user (lifecycle, and target if unclear).
- Temp-mode files do not survive the task — the Pickup Prompt says to delete.
