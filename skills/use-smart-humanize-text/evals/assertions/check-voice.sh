#!/usr/bin/env bash
# check-voice.sh - deterministic plain-docs-voice checker.
#
# Usage:   check-voice.sh <file>
# Stdout:  JSON {"violations": ["...","..."], "count": N}
# Exit:    0 if count == 0, 1 otherwise
#
# Rules (chosen because a frontier model will slip at least one per pass):
#   1. no sentence over 20 words
#   2. zero occurrences of the banned-word list below
#   3. no em dash
#   4. no sycophantic opener ("I hope this", "I'm thrilled", "we are thrilled")
#
# This is the feedback loop for the smart-skill compounding eval. The
# agent calls this script, feeds the JSON "violations" list into
# patterns.md, and the NEXT session reads that patterns.md on brain-
# protocol startup. Smart skill should converge to zero violations;
# flat skill has no mechanism to learn specifics.

set -euo pipefail

file="${1:-}"
if [ -z "$file" ] || [ ! -f "$file" ]; then
  printf '{"error":"file not found: %s","violations":[],"count":0}\n' "$file"
  exit 2
fi

violations=()

# --- Rule 1: sentences > 20 words ---------------------------------------
# Split on .!? followed by space/newline. Count words per segment.
while IFS= read -r sentence; do
  [ -z "$(printf '%s' "$sentence" | tr -d '[:space:]')" ] && continue
  n=$(printf '%s' "$sentence" | wc -w | tr -d '[:space:]')
  if [ "$n" -gt 20 ]; then
    snip=$(printf '%s' "$sentence" | head -c 50 | tr '\n' ' ' | tr -d '"')
    violations+=("long sentence ($n words): ${snip}...")
  fi
done < <(tr '\n' ' ' < "$file" | sed 's/\([.!?]\)[[:space:]]\+/\1\n/g')

# --- Rule 2: banned words (case-insensitive, word-boundary) -------------
# Deliberately long so every session has high odds of hitting at least
# one fresh word. The smart-skill arm accumulates these into patterns.md.
banned=(
  utilize leverage comprehensive robust seamless unlock thrilled delve
  meticulous vibrant pivotal groundbreaking empower crucial foster
  harness streamline cutting-edge state-of-the-art synergy synergies
  innovative transformative bespoke elevate revolutionize
)
for word in "${banned[@]}"; do
  # -i: case-insensitive, -w: word boundary, -E for fixed string fine
  # -F with -w: literal word match; hyphens okay since we search text
  if grep -qiE "(^|[^a-zA-Z-])${word}([^a-zA-Z-]|$)" "$file"; then
    violations+=("banned word: ${word}")
  fi
done

# --- Rule 3: em dash ----------------------------------------------------
if grep -q '—' "$file"; then
  violations+=("em dash present")
fi

# --- Rule 4: sycophantic openers ----------------------------------------
# Look in the first 200 chars only.
head_text=$(head -c 200 "$file" | tr '\n' ' ')
for opener in "I hope this" "hope this finds you well" "I'm thrilled" "We are thrilled" \
              "thrilled to announce" "thrilled to reach out"; do
  if printf '%s' "$head_text" | grep -qi -- "$opener"; then
    violations+=("sycophantic opener: ${opener}")
  fi
done

# --- Emit JSON ----------------------------------------------------------
printf '{"violations":['
for i in "${!violations[@]}"; do
  [ "$i" -gt 0 ] && printf ','
  # Escape double-quotes and backslashes for JSON safety.
  v=${violations[$i]//\\/\\\\}
  v=${v//\"/\\\"}
  printf '"%s"' "$v"
done
printf '],"count":%d}\n' "${#violations[@]}"

[ "${#violations[@]}" -eq 0 ]
