"""ExactMatch task dataloader.

Subclasses :class:`skillopt.datasets.base.SplitDataLoader`. The base class does
almost everything:

- ``split_mode="ratio"``  → reads ``data_path`` (a single ``.json``/``.jsonl``
  file), shuffles deterministically, and materializes ``train/``, ``val/``,
  ``test/`` directories of ``items.json``.
- ``split_mode="split_dir"`` → reads an existing pre-split directory.

We only override :meth:`load_raw_items` to normalize each raw record into the
shape our rollout/evaluator expect and to guarantee a stable ``id``.
"""
from __future__ import annotations

import json
from pathlib import Path

from skillopt.datasets.base import SplitDataLoader


def _normalize_item(raw: dict, index: int) -> dict:
    """Normalize one raw record. The only hard requirement is a string ``id``."""
    answers = raw.get("answers")
    if answers is None:
        single = raw.get("answer") or raw.get("ground_truth") or ""
        answers = [single] if single else []
    if not isinstance(answers, list):
        answers = [answers]
    return {
        "id": str(raw.get("id") or raw.get("uid") or f"item_{index:04d}"),
        "question": str(raw.get("question") or raw.get("prompt") or ""),
        "answers": [str(a) for a in answers],
        "task_type": str(raw.get("task_type") or raw.get("category") or "general"),
    }


def _read_records(path: str) -> list[dict]:
    text = Path(path).read_text(encoding="utf-8").strip()
    if not text:
        return []
    # JSON array first, then JSONL.
    if text[0] == "[":
        data = json.loads(text)
        if not isinstance(data, list):
            raise ValueError(f"Expected a JSON array at top level of {path}")
        return data
    records: list[dict] = []
    for line in text.splitlines():
        line = line.strip()
        if line:
            records.append(json.loads(line))
    return records


class ExactMatchLoader(SplitDataLoader):
    """Loader for the ExactMatch demo benchmark."""

    def load_raw_items(self, data_path: str) -> list[dict]:
        records = _read_records(data_path)
        return [_normalize_item(raw, i) for i, raw in enumerate(records)]
