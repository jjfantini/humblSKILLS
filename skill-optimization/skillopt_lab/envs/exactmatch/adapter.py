"""ExactMatch environment adapter for SkillOpt (ReflACT).

This is the smallest realistic ``EnvAdapter``: a dataset-backed, single-turn,
chat-only env with a programmatic exact-match reward. It is BOTH the smoke test
for the integration AND the template you copy to optimize a real skill.

To optimize a different skill, copy ``skillopt_lab/envs/exactmatch`` to
``skillopt_lab/envs/<your_env>``, swap in your own data + reward (the
``evaluator``), point the config at it, and register ``<your_env>`` in
``skillopt_lab/train.py``.

Contract (from ``skillopt/envs/base.py``): a concrete ``EnvAdapter`` MUST
implement the five abstract methods ``build_train_env``, ``build_eval_env``,
``rollout``, ``reflect``, ``get_task_types``. We also override
``build_env_from_batch`` to hand the rollout a plain ``list[dict]`` of items —
the same pattern used by the official ``_template``/``searchqa``/``docvqa``.
"""
from __future__ import annotations

import os

from skillopt.datasets.base import BatchSpec
from skillopt.envs.base import EnvAdapter
from skillopt.gradient.reflect import run_minibatch_reflect

from skillopt_lab.envs.exactmatch.loader import ExactMatchLoader
from skillopt_lab.envs.exactmatch.rollout import run_batch


class ExactMatchAdapter(EnvAdapter):
    """Single-turn QA env with an exact-match reward."""

    def __init__(
        self,
        split_dir: str = "",
        data_path: str = "",
        split_mode: str = "ratio",
        split_ratio: str = "2:1:1",
        split_seed: int = 42,
        split_output_dir: str = "",
        exec_timeout: int = 120,
        workers: int = 4,
        analyst_workers: int = 4,
        failure_only: bool = False,
        minibatch_size: int = 3,
        edit_budget: int = 3,
        seed: int = 42,
        limit: int = 0,
        max_completion_tokens: int = 1024,
    ) -> None:
        self.exec_timeout = exec_timeout
        self.workers = workers
        self.analyst_workers = analyst_workers
        self.failure_only = failure_only
        self.minibatch_size = minibatch_size
        self.edit_budget = edit_budget
        self.max_completion_tokens = int(max_completion_tokens)
        self.dataloader = ExactMatchLoader(
            split_dir=split_dir,
            data_path=data_path,
            split_mode=split_mode,
            split_ratio=split_ratio,
            split_seed=split_seed,
            split_output_dir=split_output_dir,
            seed=seed,
            limit=limit,
        )

    # ── Lifecycle ───────────────────────────────────────────────────────────

    def setup(self, cfg: dict) -> None:
        super().setup(cfg)
        self.dataloader.setup(cfg)

    def get_dataloader(self):
        return self.dataloader

    # ── Batch → env manager (the "env manager" is just the item list) ───────

    def build_env_from_batch(self, batch: BatchSpec, **kwargs):
        return list(batch.payload or [])

    def build_train_env(self, batch_size: int, seed: int, **kwargs):
        batch = self.dataloader.build_train_batch(batch_size=batch_size, seed=seed, **kwargs)
        return self.build_env_from_batch(batch, **kwargs)

    def build_eval_env(self, env_num: int, split: str, seed: int, **kwargs):
        batch = self.dataloader.build_eval_batch(env_num=env_num, split=split, seed=seed, **kwargs)
        return self.build_env_from_batch(batch, **kwargs)

    # ── Rollout: run the skill against a batch of items ─────────────────────

    def rollout(self, env_manager, skill_content: str, out_dir: str, **kwargs) -> list[dict]:
        items: list[dict] = env_manager
        return run_batch(
            items=items,
            out_root=out_dir,
            skill_content=skill_content,
            exec_timeout=self.exec_timeout,
            workers=self.workers,
            max_completion_tokens=self.max_completion_tokens,
        )

    # ── Reflect: turn trajectories into skill edits (generic analysts) ──────

    def reflect(self, results: list[dict], skill_content: str, out_dir: str, **kwargs) -> list[dict | None]:
        prediction_dir = kwargs.get("prediction_dir", os.path.join(out_dir, "predictions"))
        patches_dir = kwargs.get("patches_dir", os.path.join(out_dir, "patches"))
        return run_minibatch_reflect(
            results=results,
            skill_content=skill_content,
            prediction_dir=prediction_dir,
            patches_dir=patches_dir,
            workers=self.analyst_workers,
            failure_only=self.failure_only,
            minibatch_size=self.minibatch_size,
            edit_budget=self.edit_budget,
            random_seed=kwargs.get("random_seed"),
            # None → run_minibatch_reflect loads the generic analyst prompts
            # (skillopt/prompts/analyst_error.md / analyst_success.md).
            error_system=self.get_error_minibatch_prompt(),
            success_system=self.get_success_minibatch_prompt(),
            step_buffer_context=kwargs.get("step_buffer_context", ""),
            meta_skill_context=kwargs.get("meta_skill_context", ""),
            update_mode=getattr(self, "_cfg", {}).get("skill_update_mode", "patch"),
        )

    # ── Stratification hint ─────────────────────────────────────────────────

    def get_task_types(self) -> list[str]:
        return ["factual", "math", "transform"]
