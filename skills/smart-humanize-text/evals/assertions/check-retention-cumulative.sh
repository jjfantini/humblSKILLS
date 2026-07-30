#!/usr/bin/env bash
# check-retention-cumulative.sh
# -----------------------------------------------------------------------------
# Outcome-based cumulative retention check for the indie-launch-copy-iteration
# scenario. Reads the violation-count sidecars for session 5 (pure retention)
# and session 6 (generalization), sums them, and exits 0 iff the total is
# within the cap.
#
# This asserts the skill's real-world value: across the two no-feedback tail
# sessions, how many rule violations survived? A smart skill that retained the
# learned rules should hit 0 or 1 total; flat and no-skill should regress and
# hit >= 2.
#
# Usage (from within session-06's OutputDir cwd):
#   bash check-retention-cumulative.sh <s5-check-json> <s6-check-json> <max-combined>
#
# The harness sets $EVAL_WORK_DIR to the session 6 run dir; the caller is
# expected to compose the session 5 path from that.

set -euo pipefail

s5="${1:-}"
s6="${2:-}"
max="${3:-1}"

if [ ! -f "$s5" ] || [ ! -f "$s6" ]; then
  echo "cumulative: missing check file(s): s5=$s5 s6=$s6" >&2
  exit 2
fi

c5=$(python3 -c "import json,sys; print(json.load(open(sys.argv[1]))['count'])" "$s5")
c6=$(python3 -c "import json,sys; print(json.load(open(sys.argv[1]))['count'])" "$s6")
total=$((c5 + c6))

echo "cumulative retention: S5=$c5 + S6=$c6 = $total (cap $max)"
[ "$total" -le "$max" ]
