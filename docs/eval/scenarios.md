# Scenarios & authoring

Each skill can ship **`evals/scenarios.json`**. Sessions run **in order** so later sessions can depend on smart-skill brain state from earlier ones.

## Assertions

Assertions may be:

- **`llm`** - judged by a model (flexible, less deterministic)
- **Scripted** - for example `path_exists`, `exec`, `regex`, `script`, `json_valid` (prefer when you need stable CI)

Scripted checks are usually better than LLM-only judges when you care about reproducibility.

## Scaffold

```sh
humblskills eval init <skill>
```

## Reference layout

The canonical example with retention checks across sessions is under:

[`skills/use-smart-skill/evals/`](https://github.com/jjfantini/humblSKILLS/tree/develop/skills/use-smart-skill/evals/)

Use that tree as a template when authoring scenarios for a new skill.

For a worked example of pairing **cliché-rejection** with **positive-pattern reinforcement** in a single three-arm scenario — where the checker counts both banned words found AND required voice moves missing — see:

[`skills/use-smart-humanize-text/evals/`](https://github.com/jjfantini/humblSKILLS/tree/develop/skills/use-smart-humanize-text/evals/)

It ships two scenarios side by side: `adaptive-brand-voice-discovery` (surface-rule rejection) and `indie-launch-copy-iteration` (cliché rejection + voice-move reinforcement, plus a `scripts/audit-no-leaks.sh` post-eval auditor that proves no prompt leaks across the three arms).
