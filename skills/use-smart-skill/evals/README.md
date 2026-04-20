# evals/ for use-smart-skill

Canonical eval suite for the meta-skill. Demonstrates the three-arm +
longitudinal pattern the harness is built for.

## Layout

| File / dir            | Purpose                                                 |
|-----------------------|---------------------------------------------------------|
| `scenarios.json`      | Three scenarios: scaffold-ingest-lint, contradictory-pattern, stale-concept-refresh. |
| `files/`              | Input fixtures (meeting transcripts).                   |
| `assertions/`         | (Reserved for future scripted checks.)                  |

## Running

```sh
humblskills eval run use-smart-skill
humblskills eval showcase              # equivalent convenience entry
```

Results land under `$XDG_STATE_HOME/humblskills/evals/use-smart-skill/iteration-N/`.

## What each scenario proves

- **scaffold-ingest-lint**: three sessions compound. Session 1 creates the
  skill; session 2 teaches the agent kebab-case; session 3 RETAINS that
  lesson even though the user didn't restate it. Smart arm should pass
  the retention check on session 3; flat arm should regress.
- **contradictory-pattern**: append-only discipline. Smart arm keeps the
  2026-04-15 entry immutable and records a follow-up; flat arm is free
  to mutate prior entries.
- **stale-concept-refresh**: freshness audit. Smart arm knows what the
  180-day threshold means because patterns.md + decisions.md record it.
