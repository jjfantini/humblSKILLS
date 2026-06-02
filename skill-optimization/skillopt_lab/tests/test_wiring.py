"""Offline wiring tests for the SkillOpt ↔ pindex integration.

These tests prove the *integration* is correct without spending any tokens:
the only thing stubbed is the LLM dispatch boundary
(``skillopt.model.azure_openai.chat_target`` / ``.chat_optimizer``). Everything
else — env registration, ratio-splitting, the exact-match reward, rollout I/O
(including the ``conversation.json`` the reflect stage depends on), and the
reflect contract — runs for real.

Run:  uv run pytest -q
"""
from __future__ import annotations

import os

import pytest
import skillopt.model as M
from skillopt.model import azure_openai as _openai

REPO_ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
DATA_PATH = os.path.join(REPO_ROOT, "skillopt_lab", "data", "exactmatch", "raw.jsonl")


# ── 1. Reward ──────────────────────────────────────────────────────────────

def test_evaluator_exact_match_and_extraction():
    from skillopt_lab.envs.exactmatch.evaluator import evaluate, extract_answer

    assert extract_answer("blah\n<answer>Paris</answer>") == "Paris"
    assert extract_answer("The answer is Paris.") == "The answer is Paris."

    good = evaluate("<answer>Paris</answer>", ["paris"])
    assert good["em"] == 1.0
    verbose = evaluate("The capital of France is Paris.", ["paris"])
    assert verbose["em"] == 0.0  # the skill must teach terse output to score
    assert 0.0 < verbose["f1"] <= 1.0


# ── 2. Dataloader materializes the ratio split ─────────────────────────────

def test_dataloader_materializes_ratio_split(tmp_path):
    from skillopt_lab.envs.exactmatch.adapter import ExactMatchAdapter

    adapter = ExactMatchAdapter(
        data_path=DATA_PATH, split_mode="ratio", split_ratio="2:1:1", split_seed=42
    )
    adapter.setup({"out_root": str(tmp_path), "env": "exactmatch"})
    dl = adapter.get_dataloader()
    total = len(dl.train_items) + len(dl.val_items) + len(dl.test_items)
    assert total == 20  # all raw items land in some split
    assert len(dl.train_items) > 0 and len(dl.val_items) > 0 and len(dl.test_items) > 0
    item = dl.train_items[0]
    assert set(item) >= {"id", "question", "answers", "task_type"}


# ── 3. The launcher registers our env in SkillOpt's live registry ──────────

def test_launcher_registers_env():
    import scripts.train as skillopt_train

    import skillopt_lab.train  # noqa: F401  (import patches the registrar)
    from skillopt_lab.envs.exactmatch.adapter import ExactMatchAdapter

    adapter = skillopt_train.get_adapter(
        {"env": "exactmatch", "data_path": DATA_PATH, "split_mode": "ratio"}
    )
    assert isinstance(adapter, ExactMatchAdapter)


# ── 4. Rollout calls the target, scores, and writes conversation.json ──────

@pytest.fixture
def _stub_target(monkeypatch):
    """Stub the target LLM: always answer 'paris' in the taught format."""
    def fake_chat_target(system, user, **kwargs):
        return "<answer>paris</answer>", {"prompt_tokens": 0, "completion_tokens": 0, "total_tokens": 0}

    M.set_target_backend("openai_chat")
    monkeypatch.setattr(_openai, "chat_target", fake_chat_target)
    return fake_chat_target


def test_rollout_writes_conversation_and_scores(tmp_path, _stub_target):
    from skillopt_lab.envs.exactmatch.adapter import ExactMatchAdapter

    adapter = ExactMatchAdapter(data_path=DATA_PATH, split_mode="ratio", workers=2)
    adapter.setup({"out_root": str(tmp_path), "env": "exactmatch"})
    items = adapter.build_eval_env(env_num=4, split="val", seed=1)
    out_dir = str(tmp_path / "run")
    results = adapter.rollout(items, skill_content="be terse", out_dir=out_dir)

    assert len(results) == len(items)
    for r in results:
        assert {"id", "hard", "soft"} <= set(r)
        assert r["hard"] in (0, 1)
        assert 0.0 <= float(r["soft"]) <= 1.0
        conv = os.path.join(out_dir, "predictions", str(r["id"]), "conversation.json")
        assert os.path.exists(conv), f"missing trajectory file for {r['id']}"
    # The reward genuinely fires end-to-end: the val item whose gold is 'paris'
    # (if present in this split) scores 1; correctness itself is covered by
    # test_evaluator. Here we only assert the rollout produced valid scores.


# ── 5. Reflect consumes trajectories and returns a patch list ──────────────

def test_reflect_returns_patch_list(tmp_path, _stub_target, monkeypatch):
    from skillopt_lab.envs.exactmatch.adapter import ExactMatchAdapter

    def fake_chat_optimizer(system, user, **kwargs):
        return '{"patch": {"edits": []}, "analysis": "stub"}', {
            "prompt_tokens": 0, "completion_tokens": 0, "total_tokens": 0
        }

    M.set_optimizer_backend("openai_chat")
    monkeypatch.setattr(_openai, "chat_optimizer", fake_chat_optimizer)

    adapter = ExactMatchAdapter(
        data_path=DATA_PATH, split_mode="ratio", workers=2, analyst_workers=2, minibatch_size=2
    )
    adapter.setup({"out_root": str(tmp_path), "env": "exactmatch"})
    items = adapter.build_eval_env(env_num=4, split="val", seed=1)
    out_dir = str(tmp_path / "run")
    results = adapter.rollout(items, skill_content="be terse", out_dir=out_dir)

    patches = adapter.reflect(results, skill_content="be terse", out_dir=out_dir)
    assert isinstance(patches, list)  # the reflect contract: list[dict | None]


# ── 6. Full end-to-end loop (stubbed LLM) accepts a beneficial gated edit ──

def test_offline_end_to_end_gradient(tmp_path):
    """The real trainer runs the whole loop and the gate ACCEPTS an edit that
    raises validation accuracy — i.e. a genuine textual-gradient step lands."""
    from skillopt_lab.offline_demo import run_offline_demo

    out_root = str(tmp_path / "run")
    summary = run_offline_demo(out_root)

    assert summary.get("total_accepts", 0) >= 1, "gate never accepted an improving edit"
    assert summary.get("test_hard", 0.0) > summary.get("baseline_test_hard", 0.0), (
        "held-out accuracy did not improve over baseline"
    )

    best = os.path.join(out_root, "best_skill.md")
    assert os.path.exists(best)
    assert "<answer>" in open(best).read(), "learned skill is missing the format rule"


# ── 7. The documented CLI entrypoint actually runs ─────────────────────────

def test_documented_cli_entrypoint(tmp_path, monkeypatch):
    """Drive `python -m skillopt_lab.train --config ... ` exactly as the README
    instructs (stubbed LLM), proving the headline command produces best_skill.md."""
    import sys

    import skillopt_lab.train as launcher
    from skillopt_lab.offline_demo import CONFIG, install_stub_backend

    out_root = str(tmp_path / "cli_run")
    restore = install_stub_backend()
    argv = [
        "skillopt_lab.train", "--config", CONFIG, "--out_root", out_root,
        "--cfg-options",
        "model.optimizer_backend=openai_chat", "model.target_backend=openai_chat",
        "model.optimizer=stub", "model.target=stub",
    ]
    try:
        monkeypatch.setattr(sys, "argv", argv)
        launcher.main()
    finally:
        restore()

    assert os.path.exists(os.path.join(out_root, "best_skill.md"))
