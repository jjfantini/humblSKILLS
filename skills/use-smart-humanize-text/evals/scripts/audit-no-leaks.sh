#!/usr/bin/env bash
# audit-no-leaks.sh
# -----------------------------------------------------------------------------
# Post-eval auditor for the `indie-launch-copy-iteration` scenario.
#
# Proves that the only learning channel across sessions is the smart_skill
# brain. Specifically:
#
#   1. For no_skill and flat_skill, the session-5 and session-6 prompts and
#      transcripts must contain NONE of the rule-disclosure fragments that
#      appeared in sessions 2, 3, or 4. If a fragment leaks, the retention
#      test is invalid — the flat / no arms could have learned from
#      transcript-level carryover rather than being genuinely amnesic.
#
#   2. For flat_skill, every session's brain-snapshot-before/patterns.md
#      (if it exists) must be empty of scenario entries — proving the
#      harness's "re-derive flat before every session" guarantee holds.
#
#   3. For smart_skill, patterns.md must grow monotonically across sessions
#      — a sanity check that the brain is in fact accumulating lessons
#      (not strictly a leak check, but paired here so one script answers
#      "is the three-arm comparison honest?" end to end).
#
# Usage:
#   bash audit-no-leaks.sh <iteration-dir>
#
# Iteration dir is the per-iteration workspace, e.g.
#   $XDG_STATE_HOME/humblskills/evals/use-smart-humanize-text/iteration-3
#
# Emits a short summary and exits non-zero on any finding.

set -euo pipefail

iter="${1:-}"
if [ -z "$iter" ] || [ ! -d "$iter" ]; then
  echo "usage: $0 <iteration-dir>" >&2
  exit 2
fi

scenario_id="indie-launch-copy-iteration"
findings=0

# --- static scenarios.json audit -------------------------------------------
# Proves the scenario definition itself is clean: rule-disclosure text
# appears ONLY in the session where the rule is introduced. Independent of
# runner — catches leaks that transcript-level audits would miss when the
# runner truncates prompts (e.g. mock).
sj=""
for cand in \
  "${HUMBLSKILLS_ROOT:-}/skills/use-smart-humanize-text/evals/scenarios.json" \
  "$(dirname "$0")/../scenarios.json"; do
  if [ -f "$cand" ]; then sj="$cand"; break; fi
done
if [ -n "$sj" ]; then
  # Extract the prompt field of each session for our scenario, keyed by n.
  # A rule-disclosure fragment introduced in session N must appear in exactly
  # one session's prompt (session N). Finding it in any other session's
  # prompt proves a leak.
  python3 - "$sj" "$scenario_id" <<'PY' || findings=$((findings + 1))
import json, sys, re
path, sid = sys.argv[1], sys.argv[2]
with open(path) as f:
    doc = json.load(f)
for sc in doc.get("scenarios", []):
    if sc.get("id") != sid:
        continue
    prompts = {s["n"]: s["prompt"] for s in sc["sessions"]}
    # Fragments by origin session.
    fragments = {
        2: ["powerful", "seamless", "for <group>"],
        3: ["leverage", "unleash", "concrete number"],
        4: ["intuitive", "effortless", "revolutionary", "game-changer",
            "cutting-edge", "first-person sentence", "honest limitation"],
    }
    bad = 0
    for origin, needles in fragments.items():
        for n in needles:
            # Use verbatim substring (case-insensitive) with a word-boundary
            # guard for short words to avoid false positives on "AI"-style
            # substrings. All our needles are multi-char so a plain
            # case-insensitive find is fine.
            for sess_n, prompt in prompts.items():
                if sess_n == origin:
                    continue
                if n.lower() in prompt.lower():
                    print(f"STATIC LEAK: fragment '{n}' from session {origin} appears in session {sess_n} prompt")
                    bad += 1
    sys.exit(1 if bad else 0)
# Scenario id not found at all — treat as a structural error.
print("STATIC WARNING: scenario id not found in scenarios.json", file=sys.stderr)
sys.exit(2)
PY
fi


# Rule-disclosure fragments verbatim-unique to each session's prompt.
# If any of these appear in a transcript for a session where they were NOT
# disclosed, rules have leaked across the three-arm boundary.
#
# Each entry is: "<session-N-origin>\t<needle>"
fragments_s2=(
  "don't use the word 'powerful'"
  "don't use 'seamless' or 'seamlessly'"
  "name a specific audience with 'for <group>'"
)
fragments_s3=(
  "don't use the word 'leverage'"
  "don't use 'unleash'"
  "concrete number with a unit"
)
fragments_s4=(
  "don't use 'intuitive'"
  "don't use 'effortless'"
  "don't use 'revolutionary'"
  "don't use 'game-changer'"
  "don't use 'cutting-edge'"
  "first-person sentence"
  "name one honest limitation"
)

# Sessions that must NOT carry forward any of the above fragments (no feedback
# sessions for the three-arm retention test).
later_sessions=(05 06)

# Session paths: iter/<arm>/session-NN-<scenario>/run-1
for arm in no_skill flat_skill; do
  for s in "${later_sessions[@]}"; do
    sess_glob="${iter}/${arm}/session-${s}-${scenario_id}/run-1"
    if [ ! -d "$sess_glob" ]; then
      # Allow running against partial iterations (e.g., smart-only) without
      # failing — just skip. But report it so the user notices.
      echo "note: skipping missing session dir: $sess_glob"
      continue
    fi
    transcript="$sess_glob/transcript.txt"
    prompt_file="$sess_glob/prompt.txt"
    # Some runners emit `prompt.txt`; others inline it in `transcript.txt`.
    for needle_list in "fragments_s2" "fragments_s3" "fragments_s4"; do
      declare -n list="$needle_list"
      for needle in "${list[@]}"; do
        hit=0
        if [ -f "$transcript" ] && grep -Fq -- "$needle" "$transcript"; then
          hit=1
        fi
        if [ -f "$prompt_file" ] && grep -Fq -- "$needle" "$prompt_file"; then
          hit=1
        fi
        if [ "$hit" = 1 ]; then
          echo "LEAK: $arm/session-${s}: found prior-session fragment '$needle'"
          findings=$((findings + 1))
        fi
      done
      unset -n list
    done
  done
done

# Flat arm: patterns.md in every session's brain-snapshot-before should be
# empty of scenario entries. The derived-flat skill snapshots brain files
# from the source, so inheriting a stock patterns.md is fine; what must NOT
# appear is scenario-specific lessons from prior sessions.
for s in 01 02 03 04 05 06; do
  pfile="${iter}/flat_skill/session-${s}-${scenario_id}/run-1/brain-snapshot-before/references/patterns.md"
  [ -f "$pfile" ] || continue
  if grep -qiE 'Liana|Warpshelf|Tabpile|Spritemash|Queuedeck|Thinkmoss|Plipspace' "$pfile"; then
    echo "LEAK: flat_skill/session-${s}: patterns.md snapshot-before contains scenario-specific entries"
    findings=$((findings + 1))
  fi
done

# Smart arm monotonic-growth sanity check. patterns.md byte-size at snapshot-
# after should be >= snapshot-before for every session. If it shrinks the
# brain is not accumulating.
prev_size=-1
for s in 01 02 03 04 05 06; do
  snap="${iter}/smart_skill/session-${s}-${scenario_id}/run-1/brain-snapshot-after/references/patterns.md"
  [ -f "$snap" ] || continue
  size=$(wc -c < "$snap" | tr -d ' ')
  if [ "$prev_size" -ge 0 ] && [ "$size" -lt "$prev_size" ]; then
    echo "WARNING: smart_skill patterns.md shrank from $prev_size to $size bytes at session $s"
    findings=$((findings + 1))
  fi
  prev_size="$size"
done

if [ "$findings" -eq 0 ]; then
  echo "leaks: none"
  exit 0
fi
echo "leaks: found — $findings issue(s)"
exit 1
