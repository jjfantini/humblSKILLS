"""Offline, zero-network end-to-end demo of the SkillOpt training loop.

This runs the REAL ``ReflACTTrainer`` over the ``exactmatch`` env, but stubs the
LLM boundary so it needs no API keys and spends no tokens:

* **Target (the agent)** is an oracle that returns the gold answer wrapped in
  ``<answer>`` tags *only when the current skill contains the formatting rule*;
  otherwise it returns verbose prose that fails exact-match. So the skill text
  genuinely controls the reward.
* **Optimizer (the analyst)** proposes a single ``append`` edit that adds that
  exact formatting rule.

The validation **gate accepts** the edit and held-out accuracy goes 0 → 1.0.

NOTE ON WHAT THIS PROVES: the acceptance is *rigged by construction* — we wrote
both the oracle and the matching edit. This is a **wiring/plumbing proof** that
the whole pipeline (rollout → reflect → optimize → gate → ``best_skill.md``) is
connected correctly. It is NOT evidence that SkillOpt improves a real skill;
that requires a live model and a genuine task (see the README).

Run as a script:
    uv run python -m skillopt_lab.offline_demo
"""
from __future__ import annotations

import json
import os
import re
from typing import Callable

REPO_ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
_DATA = os.path.join(REPO_ROOT, "skillopt_lab", "data", "exactmatch", "raw.jsonl")
CONFIG = os.path.join(REPO_ROOT, "skillopt_lab", "configs", "exactmatch.yaml")
_MARKER = "wrap the final answer in <answer>"


def _gold_map() -> dict[str, list[str]]:
    gold: dict[str, list[str]] = {}
    with open(_DATA) as f:
        for line in f:
            line = line.strip()
            if line:
                row = json.loads(line)
                gold[row["question"].strip()] = row["answers"]
    return gold


def install_stub_backend() -> Callable[[], None]:
    """Patch ``skillopt.model.azure_openai`` with a deterministic oracle target +
    formatting-rule optimizer. Returns a callable that restores the originals.
    """
    os.environ["OPTIMIZER_BACKEND"] = "openai_chat"
    os.environ["TARGET_BACKEND"] = "openai_chat"

    from skillopt.model import azure_openai as O
    from skillopt.model.common import CompatAssistantMessage

    gold = _gold_map()

    def _question(user: str) -> str:
        m = re.search(r"## Question\n(.*)", user, re.DOTALL)
        text = m.group(1).strip() if m else user.strip()
        return text.splitlines()[0].strip() if text else ""

    def fake_target(system, user, **_kw):
        golds = gold.get(_question(user), [""])
        usage = {"prompt_tokens": 1, "completion_tokens": 1, "total_tokens": 2}
        if _MARKER in (system or "").lower():
            return f"<answer>{golds[0]}</answer>", usage
        return f"The answer to your question is {golds[0]}, I hope that helps!", usage

    patch = json.dumps(
        {
            "patch": {
                "edits": [
                    {
                        "op": "append",
                        "content": f"## Formatting\nAlways {_MARKER}</answer> tags and output ONLY the final answer with no extra words.",
                    }
                ]
            },
            "analysis": "stub: enforce answer formatting",
        }
    )

    def fake_opt(system, user, **_kw):
        return patch, {"prompt_tokens": 1, "completion_tokens": 1, "total_tokens": 2}

    def fake_opt_msgs(messages, **kw):
        usage = {"prompt_tokens": 1, "completion_tokens": 1, "total_tokens": 2}
        if kw.get("return_message"):
            return CompatAssistantMessage(content=patch, tool_calls=[]), usage
        return patch, usage

    names = (
        "chat_target", "chat_optimizer", "chat_with_deployment",
        "chat_optimizer_messages", "chat_target_messages", "chat_messages_with_deployment",
    )
    originals = {name: getattr(O, name) for name in names}
    O.chat_target = fake_target
    O.chat_optimizer = fake_opt
    O.chat_with_deployment = lambda deployment, system, user, **kw: fake_opt(system, user, **kw)
    O.chat_optimizer_messages = fake_opt_msgs
    O.chat_target_messages = fake_opt_msgs
    O.chat_messages_with_deployment = lambda deployment, messages, **kw: fake_opt_msgs(messages, **kw)

    def _restore() -> None:
        for name, fn in originals.items():
            setattr(O, name, fn)

    return _restore


def run_offline_demo(out_root: str) -> dict:
    """Run the stubbed end-to-end loop (trainer built directly) and return the summary."""
    restore = install_stub_backend()
    try:
        import scripts.train as T

        import skillopt_lab.train  # noqa: F401  (registers the exactmatch env)

        class _Args:
            config = CONFIG
            cfg_options = [
                "model.optimizer_backend=openai_chat",
                "model.target_backend=openai_chat",
                "model.optimizer=stub",
                "model.target=stub",
            ]

            def __getattr__(self, _name):
                return None

        cfg = T.load_config(_Args())
        cfg["out_root"] = os.path.abspath(out_root)
        adapter = T.get_adapter(cfg)

        from skillopt.engine.trainer import ReflACTTrainer

        return ReflACTTrainer(cfg, adapter).train()
    finally:
        restore()


if __name__ == "__main__":
    import tempfile

    out = os.path.join(tempfile.gettempdir(), "skillopt_offline_demo")
    summary = run_offline_demo(out)
    print("\n" + "=" * 60)
    print("OFFLINE DEMO SUMMARY  (wiring proof — acceptance is rigged by design)")
    print("=" * 60)
    print(f"  steps              : {summary.get('total_steps')}")
    print(f"  accepted edits     : {summary.get('total_accepts')}")
    print(f"  rejected edits     : {summary.get('total_rejects')}")
    print(f"  baseline test hard : {summary.get('baseline_test_hard')}")
    print(f"  best test hard     : {summary.get('test_hard')}  (delta {summary.get('test_delta_hard'):+})")
    print(f"  best_skill.md      : {os.path.join(out, 'best_skill.md')}")
