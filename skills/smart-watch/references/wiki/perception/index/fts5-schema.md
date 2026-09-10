---
title: "One SQLite FTS5 Index Across Every Watched Source"
context: perception
category: index
concept: fts5-schema
description: "Schema, tokenizer choice (unicode61, no stemming) and query shapes for ask.sh; why porter was dropped"
tags: sqlite, fts5, index, search, schema, tokenizer, bm25
sources:
  - "references/raw/probe-findings-2026-09-10.md"
last_ingested: 2026-09-10
command: scripts/ask.sh
---

## Schema

`~/.local/state/smart-watch/index.db` (WAL), created by `lib.sh::db_init`:

```sql
sources(id PK, path, duration, fps, has_audio, created_at)
segments(rowid, source_id, kind 'ocr'|'speech', t_start, t_end, frame, text)
segments_fts  -- FTS5 external-content table over segments.text, kept in sync by triggers
```

One DB for every source is what makes `ask.sh all "<query>"` (cross-recording search) free.
`<id>` is the first 12 hex of the source file's SHA-256, so re-watching the same file is a no-op.

**Incorrect (porter stemmer):**

```sql
tokenize='porter unicode61'
-- "deployment" indexes as deploy, the query "deploy" stems to deploi: zero hits
```

**Correct:**

```sql
tokenize='unicode61 remove_diacritics 2'
-- whole-word tokens; use deploy* for prefix matches, "exact phrase" for phrases
```

Query shapes `ask.sh` accepts (FTS5 syntax passed through): `error`, `timeout OR failed`,
`"all passed"`, `deploy*`, `error NOT warning`. Results rank by `bm25` then time; `--from/--to`
bound `t_start`, `--kind` restricts to `ocr` or `speech`, `--timeline` dumps in time order.

sqlite gotchas baked into `lib.sh`:
- `PRAGMA busy_timeout` returns a row; `db()` silences it with `.output /dev/null`.
- `sqlite3` ignores stdin once given SQL arguments; bulk inserts go through `db_file` (`.read`).
- `%` truncates REALs, so hh:mm:ss is rendered by `sql_hms` with explicit casts.
- Quote escaping needs the quote in a variable: `q="'"; ${v//$q/$q$q}`.

## Sources

- `references/raw/probe-findings-2026-09-10.md` - FTS5 availability check and tokenizer observation.
