"""Programmatic reward for the ExactMatch demo environment.

SkillOpt optimizes a skill *against a numeric reward*. This module is that
reward: it extracts the model's final answer and scores it against the gold
answer(s). The scoring is deliberately strict on format so that the *skill*
(which controls how the model formats its answer) measurably moves the score
— that is what makes the "textual gradient" real and observable.

Mirrors the structure of ``skillopt/envs/searchqa/evaluator.py`` and
``skillopt/envs/docvqa/evaluator.py`` (answer extraction + normalized match),
but with a simple exact-match metric instead of ANLS/F1.
"""
from __future__ import annotations

import re
from typing import Any


def _normalize(text: Any) -> str:
    """Lowercase, drop punctuation, collapse whitespace."""
    lowered = str(text or "").strip().lower()
    no_punct = re.sub(r"[^a-z0-9 ]+", " ", lowered)
    return " ".join(no_punct.split())


def extract_answer(text: str) -> str:
    """Pull the final answer out of a model response.

    Priority:
      1. The last ``<answer>...</answer>`` block (the format the skill should teach).
      2. Otherwise the last non-empty line (lenient fallback).
    """
    lower = text.lower()
    start = lower.rfind("<answer>")
    end = lower.rfind("</answer>")
    if start != -1 and end != -1 and end > start:
        return text[start + len("<answer>"):end].strip()
    lines = [line.strip() for line in text.splitlines() if line.strip()]
    return lines[-1] if lines else text.strip()


def _token_f1(pred: str, gold: str) -> float:
    """Token-overlap F1 — a graded ``soft`` signal between 0 and 1."""
    pred_tokens = pred.split()
    gold_tokens = gold.split()
    if not pred_tokens and not gold_tokens:
        return 1.0
    if not pred_tokens or not gold_tokens:
        return 0.0
    overlap = 0
    remaining = list(gold_tokens)
    for token in pred_tokens:
        if token in remaining:
            overlap += 1
            remaining.remove(token)
    if overlap == 0:
        return 0.0
    precision = overlap / len(pred_tokens)
    recall = overlap / len(gold_tokens)
    return 2 * precision * recall / (precision + recall)


def evaluate(prediction_text: str, gold_answers: Any) -> dict:
    """Score a raw model response against the gold answer(s).

    Returns
    -------
    dict
        ``em``   : 1.0 if the normalized answer exactly matches any gold, else 0.0
                   (this becomes ``hard`` — the gate/accuracy metric).
        ``f1``   : best token-overlap with any gold (this becomes ``soft``).
        ``predicted_answer`` : the extracted answer string.
        ``gold_answers``     : the gold list (for logging).
    """
    predicted = extract_answer(prediction_text)
    predicted_norm = _normalize(predicted)

    golds = gold_answers if isinstance(gold_answers, list) else [gold_answers]
    golds_norm = [_normalize(g) for g in golds if str(g).strip()]

    em = 1.0 if predicted_norm and predicted_norm in golds_norm else 0.0
    soft = max((_token_f1(predicted_norm, g) for g in golds_norm), default=0.0)

    return {
        "em": em,
        "f1": soft,
        "predicted_answer": predicted,
        "gold_answers": golds,
    }
