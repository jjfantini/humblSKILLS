# Decisions

Reasoning memory. Append new decisions; do not rewrite history.

---

### 2026-07-29 | Preserve upstream and separate review from implementation
- Context: Port upstream design guidance into a progressive-disclosure humblSKILL.
- Options: (A) rewrite upstream in place, (B) preserve raw sources and distill wiki concepts, (C) ship only the upstream flat skill.
- Chose: B - immutable raw sources plus thin router and atomic wiki concepts.
- Why: Maintains attribution and exact source fidelity while making runtime loading selective. Review stays read-only unless implementation is explicitly requested.
- Result: Raw Markdown is byte-identical to upstream and every concept cites its source.
