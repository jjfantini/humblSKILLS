#!/bin/bash
# Dispatch one orchestration brief to a cursor-agent worker, retrying the
# known-flaky duplex stream failure.
#
# Usage: dispatch-cursor-worker.sh <worktree-path> <brief-file> [model]
#
# The failure this exists for: cursor-agent's stream to
# agentn.global.api5.cursor.sh dies before the worker does any work, printing
# "Connection lost, reconnecting" x3 then "RetriableError: WritableIterable is
# closed" and exiting 1 after ~30s.
#
# THE PRIMARY FIX IS THE MODEL, NOT THE RETRY. See the full measured table in
# references/wiki/orchestrate/routing/cursor-cli-models.md (30 verified-good
# IDs, 19 verified-bad, measured 2026-08-13 with controls). Headlines:
#
#   NEVER  auto                        0/12  never succeeded once
#   NEVER  any cursor-grok-4.6-*       ~3/24 across all 8 tiers
#   NEVER  composer-2.5[-fast]         1/4, 1/3
#   NEVER  gpt-5.4-nano-*              invalid ID - --list-models over-reports
#   GOOD   gpt-5.3-codex-low-fast      15/15  4-7s  <- the default below
#   GOOD   gpt-5.6-luna-high            3/3   5s    fastest verified
#   GOOD   claude-opus-5-thinking-high  3/3   9s    hard briefs
#   GOOD   cursor-grok-4.5-low-fast     9/9   9s    only safe Grok ID
#
# `auto` lets Cursor route to whatever provider it likes, including one that is
# mid-incident, and that routing failure surfaces as a stream teardown rather
# than an honest upstream error. Always pass an explicit --model.
#
# Reliability is per-model-ID: it does not follow the family, the effort tier,
# or the -fast suffix. Measure any new ID at n>=6 before adopting it.
#
# ponytail: blind retry, safe only because every observed failure happens
# BEFORE the worker writes anything. Cursor also has a distinct mid-turn
# variant ("http/2 stream closed CANCEL", ~15-20min in) that would leave a
# brief half-applied -- hence the dirty-tree guard below. Keep briefs small.
set -uo pipefail

WORKTREE="${1:?worktree path required}"
BRIEF_FILE="${2:?brief file required}"
MODEL="${3:-${CURSOR_WORKER_MODEL:-gpt-5.3-codex-low-fast}}"
MAX_ATTEMPTS="${CURSOR_WORKER_ATTEMPTS:-6}"

BRIEF="$(cat "$BRIEF_FILE")"

# Refuse to retry into a tree a previous attempt already modified -- a rerun
# there double-applies the brief.
dirty() {
  git -C "$WORKTREE" rev-parse --git-dir >/dev/null 2>&1 || return 1
  [ -n "$(git -C "$WORKTREE" status --porcelain)" ]
}

if dirty; then
  echo "cursor-worker: $WORKTREE is already dirty; refusing to dispatch." >&2
  echo "Commit, stash, or reset it first so a retry cannot double-apply." >&2
  exit 2
fi

for attempt in $(seq 1 "$MAX_ATTEMPTS"); do
  out="$(cursor-agent -p --force --trust --model "$MODEL" \
          --workspace "$WORKTREE" "$BRIEF" 2>&1)"
  rc=$?

  if [ "$rc" -eq 0 ]; then
    printf '%s\n' "$out"
    exit 0
  fi

  # Only the transport failure is retryable. Anything else is a real error the
  # parent must read, not paper over.
  if ! printf '%s' "$out" | grep -q 'WritableIterable is closed'; then
    echo "cursor-worker: non-transport failure (rc=$rc), not retrying:" >&2
    printf '%s\n' "$out" >&2
    exit "$rc"
  fi

  if dirty; then
    echo "cursor-worker: stream died AFTER the worker wrote files." >&2
    echo "Tree is dirty -- stopping so a retry cannot double-apply." >&2
    echo "Inspect: git -C '$WORKTREE' status" >&2
    exit 3
  fi

  echo "cursor-worker: attempt $attempt/$MAX_ATTEMPTS lost the stream, retrying." >&2
done

echo "cursor-worker: gave up after $MAX_ATTEMPTS attempts." >&2
exit 1
