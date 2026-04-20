---
title: "Runner Interface - One Contract, Six Backends"
context: eval
category: runner
concept: interface
description: "Every eval runner (claudecode, cursor-agent, codex, anthropic-api, openai-api, mock) satisfies one Go interface. Scenarios are portable across runners."
tags: eval, runner, humblskills
sources: []
last_ingested: 2026-04-19
---

## Contract

```go
type Runner interface {
    Name() string
    Capabilities() Capabilities
    DoctorCheck(ctx context.Context) DoctorCheck
    Execute(ctx context.Context, req Request) (*Result, error)
}
```

`Request` is a skill path + prompt + input files + output dir. `Result`
is tokens + duration + transcript + tool call counts + output file list.

Two behaviour classes:

| Class             | Runners                               | Auth path                                |
|-------------------|---------------------------------------|------------------------------------------|
| CLI wrappers      | claudecode, cursor-agent, codex       | Agent's own login (`claude`, `cursor-agent`, `codex`) |
| Direct API        | anthropic-api, openai-api             | `secrets.Store` (env > keyring > file)   |

Both classes share:

- A scratch directory per request (skill staged under `scratch/skill`,
  inputs under `scratch/inputs/`).
- The same 5-tool sandbox for API runners (`Read` / `Write` / `Bash` /
  `Glob` / `Grep` - all scoped to scratch).
- The same stream-json event parse convention for CLI runners (tool_use
  and usage fields).

## Why it's a flat list, not a plugin system

`cli/internal/eval/evalruntime.DefaultRegistry(store)` returns the
canonical six-runner registry. The order is the auto-detect priority.
Adding a seventh runner = add a backend package + prepend it to the
registry list. No plugin loader, no YAML, no generics.

## Sources

- [cli/internal/eval/runner/runner.go](../../../../../../cli/internal/eval/runner/runner.go) - the interface
- [cli/internal/eval/runner/clitool/clitool.go](../../../../../../cli/internal/eval/runner/clitool/clitool.go) - shared glue for the three CLI backends
- [cli/internal/eval/runner/toolbox/toolbox.go](../../../../../../cli/internal/eval/runner/toolbox/toolbox.go) - sandboxed Read/Write/Bash/Glob/Grep
