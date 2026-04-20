# Eval overview

`humblskills eval` runs a **three-arm** benchmark for any skill:

- **`no_skill`** - baseline without the skill
- **`flat_skill`** - skill injected without smart multi-session state
- **`smart_skill`** - full smart skill; sessions run **in order** so brain state (patterns, decisions, log, wiki) carries across runs

Outputs are graded and summarized in a **single-file HTML** dashboard (plus JSON/Markdown mirrors). For smart skills, you get a **trajectory** that shows compounding across sessions.

## Runners

Pick the runner that matches how you authenticate today, or use the mock runner in CI.

| Runner | Auth | Notes |
|--------|------|-------|
| `claudecode` | Claude Code login | Wraps `claude -p --output-format stream-json` |
| `cursor-agent` | Cursor login | Wraps `cursor-agent` headless CLI |
| `codex` | Codex login | Wraps the OpenAI `codex` CLI |
| `anthropic-api` | `ANTHROPIC_API_KEY` / keyring | Pure-Go Read/Write/Bash/Glob/Grep tool loop |
| `openai-api` | `OPENAI_API_KEY` / keyring | Pure-Go tool loop |
| `mock` | none | Deterministic, zero tokens (CI and dev) |

Next: [Eval quickstart](quickstart.md), [Artifacts](artifacts.md), [Scenarios](scenarios.md).
