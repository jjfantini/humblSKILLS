# evals/ for use-smart-humanize-text

Two-session scenario that proves the de-slop rules compound across
sessions. Session 2 is a retention test: the agent must not reintroduce
the banned vocabulary it stripped in session 1.

## Running

```sh
humblskills eval run use-smart-humanize-text
```

## What it tests

- Vocabulary strips: em dashes, "leverage", "seamless", "thrilled", "game-changer", "synergies", "fast-paced world", "unparalleled".
- Length discipline (session 2 under 70 words).
- Retention: session 2 must hold the rules from session 1 without being re-told.
