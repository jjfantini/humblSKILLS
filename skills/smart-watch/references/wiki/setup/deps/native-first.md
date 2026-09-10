---
title: "Dependency Policy: Native First, Brew With Approval, Never Silent"
context: setup
category: deps
concept: native-first
description: "How dependencies are declared in SKILL.md metadata.dependencies, detected by doctor.sh, and installed only after the user says yes"
tags: dependencies, doctor, brew, approval, frontmatter, compatibility, native
sources:
  - "references/raw/probe-findings-2026-09-10.md"
last_ingested: 2026-09-10
command: scripts/doctor.sh
---

## One List, Read by Humans and by doctor.sh

The dependency list lives in **one place**: the `metadata.dependencies:` block of `SKILL.md`, one
flow mapping per line. `doctor.sh` parses that block with awk/sed, so the frontmatter and the
detector cannot drift. Top-level `compatibility:` (the agentskills spec field) carries the prose
summary of the same list.

**Incorrect:**

```yaml
metadata:
  requires: [ffmpeg]   # requires = skill-to-skill deps, validated against the registry:
                       # make registry fails with: unknown dep "ffmpeg"
```

**Correct:**

```yaml
metadata:
  dependencies:
    - { name: ffmpeg, kind: brew, required: true, provides: "frame + audio extraction" }
    - { name: whisper-cpp, kind: brew, required: false, provides: "...", when: "macos_lt_26" }
```

Fields: `name` (binary or formula), `kind` (`brew` | `native`), `required`, `provides`, optional
`install` (non-brew instruction) and `when` (`macos_lt_26` gates the whisper fallback).

Policy, in order:
1. Native macOS first: `screencapture`, Vision, SpeechAnalyzer, `sqlite3`, `shasum`, `curl`, `swiftc`.
2. Homebrew only for well-known public formulae with no native equivalent: `ffmpeg` (required),
   `jq`, `yt-dlp`, `whisper-cpp` (optional).
3. **Nothing installs without approval.** `doctor.sh` reports and prints the exact `brew install`
   line. `doctor.sh --install` prompts on a TTY; without a TTY it refuses unless `--yes` is passed,
   which the agent may only do after the user has agreed in conversation.

Runtime state lives outside the skill directory so `humblskills update` cannot wipe it:
`~/.cache/smart-watch/bin/watchkit` (rebuilt when the Swift source hash changes) and
`~/.local/state/smart-watch/` (index + sources).

## Sources

- `references/raw/probe-findings-2026-09-10.md` - inventory of native tools verified on this machine.
