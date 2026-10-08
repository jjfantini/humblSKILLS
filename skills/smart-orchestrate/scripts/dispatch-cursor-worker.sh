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
# references/wiki/orchestrate/routing/cursor-cli-models.md (measured 2026-08-13
# and re-measured 2026-10-08 on CLI 2026.10.01, with controls). Headlines:
#
#   NEVER  auto                        0/12  never succeeded once
#   NEVER  any cursor-grok-4.6-*       ~3/24 across all 8 tiers
#   NEVER  composer-2.5[-fast]         1/4, 1/3, 0/3 again on 2026-10-08
#   NEVER  grok-4.7-*                  1/6 across two tiers, stream teardown
#   NEVER  gpt-5.4-nano-*              invalid ID - --list-models over-reports
#   GOOD   claude-haiku-5-5-thinking-medium  6/6  5-6s  <- the default below
#          (real brief 15s; replaced gpt-5.3-codex-low-fast, 27/27, because
#          every gpt-* ID is expected to leave Cursor at OpenAI's proposed
#          2026-11-12 cutoff - route GPT-6 briefs to `codex exec` instead)
#   GOOD   claude-sonnet-5-5-medium     6/6   5-8s  mid-tier
#   GOOD   gpt-5.6-luna-high            3/3   5s    fastest verified
#   GOOD   claude-opus-5-thinking-high  3/3   9-28s hard briefs (Opus 5.5 throttled)
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
MODEL="${3:-${CURSOR_WORKER_MODEL:-claude-haiku-5-5-thinking-medium}}"
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

  # Only the transport failure and provider capacity are retryable. Anything
  # else is a real error the parent must read, not paper over.
  # ponytail: linear backoff on resource_exhausted; exponential if it keeps biting.
  if printf '%s' "$out" | grep -q 'resource_exhausted'; then
    sleep $((attempt * 15))
  elif ! printf '%s' "$out" | grep -q 'WritableIterable is closed'; then
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

  echo "cursor-worker: attempt $attempt/$MAX_ATTEMPTS failed retryably (stream or capacity), retrying." >&2
done

echo "cursor-worker: gave up after $MAX_ATTEMPTS attempts." >&2
exit 1
