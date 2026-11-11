"""ExactMatch rollout — single-turn answer agent + batch execution.

The target model receives the current skill document as part of its system
prompt, then answers a short question. We score the answer with the
exact-match reward and persist the trajectory so the optimizer's reflect
stage can read it.

IMPORTANT (easy to get wrong): ``run_minibatch_reflect`` →
``fmt_minibatch_trajectories`` reads ``<prediction_dir>/<id>/conversation.json``
for every result. If that file does not exist the trajectory is silently
skipped, so **no patches are produced and the skill never changes**. Therefore
``process_one`` MUST write ``conversation.json`` (including an evaluation
summary the analyst can learn from). This mirrors
``skillopt/envs/searchqa/rollout.py``.

Public API
----------
- :func:`process_one`  — run + evaluate one item
- :func:`run_batch`    — parallel execution of a list of items (resume-aware)
"""
from __future__ import annotations

import json
import os
from concurrent.futures import ThreadPoolExecutor, as_completed

from skillopt.model import chat_target

from skillopt_lab.envs.exactmatch.evaluator import evaluate

# Fixed task framing. The skill is layered ON TOP of this; a good skill teaches
# the model to comply with the <answer> contract and answer tersely, which is
# exactly what raises exact-match accuracy.
_BASE_INSTRUCTIONS = (
    "You answer short factual questions. Give the single best final answer."
)


def _build_system(skill_content: str) -> str:
    skill = skill_content.strip()
    if skill:
        return f"{_BASE_INSTRUCTIONS}\n\n## Skill\n{skill}"
    return _BASE_INSTRUCTIONS


def _build_user(question: str) -> str:
    return f"## Question\n{question}"


def process_one(
    item: dict,
    out_root: str,
    skill_content: str,
    exec_timeout: int = 120,
    max_completion_tokens: int = 1024,
) -> dict:
    """Run the answer agent on a single item and score it.

    Parameters
    ----------
    item : dict
        Must have ``id``, ``question``, ``answers``. ``task_type`` is optional.
    out_root : str
        Predictions are saved under ``<out_root>/predictions/<id>/``.
    skill_content : str
        Current skill document text (the thing being optimized).

    Returns
    -------
    dict
        Rollout result with ``id``/``hard``/``soft`` plus extras consumed by
        the analyst (``task_description``, ``task_type``, ``fail_reason`` ...).
    """
    item_id = str(item["id"])
    question = str(item.get("question", ""))
    gold_answers = item.get("answers", [])
    task_type = str(item.get("task_type") or "general")

    result = {
        "id": item_id,
        "question": question,
        "task_description": question,
        "task_type": task_type,
        "hard": 0,
        "soft": 0.0,
        "predicted_answer": "",
        "gold_answers": gold_answers,
        "response": "",
        "fail_reason": "",
        "agent_ok": False,
        "n_turns": 0,
    }

    pred_dir = os.path.join(out_root, "predictions", item_id)
    os.makedirs(pred_dir, exist_ok=True)

    try:
        system = _build_system(skill_content)
        user = _build_user(question)

        response, _usage = chat_target(
            system=system,
            user=user,
            max_completion_tokens=max_completion_tokens,
            retries=4,
            stage="rollout",
            timeout=exec_timeout,
        )

        eval_result = evaluate(response, gold_answers)
        result["response"] = response
        result["agent_ok"] = True
        result["n_turns"] = 1
        result["predicted_answer"] = eval_result["predicted_answer"]
        result["hard"] = int(eval_result["em"])
        result["soft"] = float(eval_result["f1"])
        if eval_result["em"] < 1.0:
            result["fail_reason"] = (
                f"EM=0: predicted {eval_result['predicted_answer']!r} "
                f"but expected one of {gold_answers!r}"
            )

        # The conversation the analyst will read. The trailing system message
        # carries the evaluation verdict so the optimizer can reason about WHY
        # the answer scored as it did.
        conversation = [
            {"type": "message", "turn": 1, "content": response},
            {
                "role": "system",
                "content": (
                    "[EVALUATION RESULT]\n"
                    f"Question: {question}\n"
                    f"Predicted answer: {eval_result['predicted_answer']!r}\n"
                    f"Gold answers: {gold_answers!r}\n"
                    f"Exact match (hard): {result['hard']}\n"
                    f"Token-F1 (soft): {result['soft']:.4f}"
                ),
            },
        ]
        with open(os.path.join(pred_dir, "target_system_prompt.txt"), "w") as f:
            f.write(system)
        with open(os.path.join(pred_dir, "target_user_prompt.txt"), "w") as f:
            f.write(user)
        with open(os.path.join(pred_dir, "conversation.json"), "w") as f:
            json.dump(conversation, f, ensure_ascii=False, indent=2)

    except Exception as exc:  # noqa: BLE001
        result["fail_reason"] = f"error: {type(exc).__name__}: {exc}"
        # Still write a minimal trajectory so the failure is visible to reflect.
        with open(os.path.join(pred_dir, "conversation.json"), "w") as f:
            json.dump(
                [{"role": "system", "content": f"[ERROR] {result['fail_reason']}"}],
                f,
                ensure_ascii=False,
                indent=2,
            )

    return result


def run_batch(
    items: list[dict],
    out_root: str,
    skill_content: str,
    exec_timeout: int = 120,
    workers: int = 4,
    max_completion_tokens: int = 1024,
) -> list[dict]:
    """Run the answer agent on all items in parallel. Resume-aware.

    Already-scored items (by ``id``) are read back from ``results.jsonl`` so a
    re-run / resume does not re-spend tokens.
    """
    os.makedirs(out_root, exist_ok=True)
    results_path = os.path.join(out_root, "results.jsonl")

    done_ids: set[str] = set()
    existing: list[dict] = []
    if os.path.exists(results_path):
        with open(results_path) as f:
            for line in f:
                line = line.strip()
                if not line:
                    continue
                try:
                    row = json.loads(line)
                    done_ids.add(str(row["id"]))
                    existing.append(row)
                except Exception:  # noqa: BLE001
                    pass

    pending = [it for it in items if str(it["id"]) not in done_ids]
    if not pending:
        return existing

    total = len(existing) + len(pending)
    completed = len(existing)
    correct = sum(1 for r in existing if r.get("hard", 0))
    results = list(existing)

    def _run_one(item: dict) -> dict:
        return process_one(
            item,
            out_root,
            skill_content,
            exec_timeout=exec_timeout,
            max_completion_tokens=max_completion_tokens,
        )

    with open(results_path, "a") as outf:
        with ThreadPoolExecutor(max_workers=max(1, workers)) as ex:
            futs = {ex.submit(_run_one, it): it for it in pending}
            for fut in as_completed(futs):
                item = futs[fut]
                try:
                    res = fut.result()
                except Exception as exc:  # noqa: BLE001
                    res = {
                        "id": str(item["id"]),
                        "question": item.get("question", ""),
                        "task_description": item.get("question", ""),
                        "task_type": item.get("task_type") or "general",
                        "hard": 0,
                        "soft": 0.0,
                        "predicted_answer": "",
                        "fail_reason": f"unexpected: {type(exc).__name__}: {exc}",
                        "agent_ok": False,
                        "n_turns": 0,
                    }
                results.append(res)
                completed += 1
                if res.get("hard", 0):
                    correct += 1
                acc = correct / completed if completed else 0.0
                print(
                    f"    [rollout] {completed}/{total} (acc={acc:.3f}) "
                    f"id={res['id']} hard={res.get('hard', '?')}",
                    flush=True,
                )
                outf.write(json.dumps(res, ensure_ascii=False) + "\n")
                outf.flush()

    return results
