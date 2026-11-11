#!/usr/bin/env python3
"""pindex training launcher for SkillOpt.

This is a thin wrapper around SkillOpt's ``scripts/train.py``. It injects
pindex's custom environments into the framework's live environment registry
(``scripts.train._ENV_REGISTRY``) *without modifying the vendored SkillOpt
source*, so the ``external/SkillOpt`` submodule stays pristine and updatable.

Why this works: ``scripts.train._register_builtins()`` only *adds* keys to the
module-level ``_ENV_REGISTRY`` dict (lazy ``try/except ImportError`` imports).
We wrap that function so it also registers our envs every time it runs.

Usage
-----
    uv run python -m skillopt_lab.train \
        --config skillopt_lab/configs/exactmatch.yaml

Any SkillOpt CLI flag / ``--cfg-options section.key=value`` override works,
because we delegate straight to ``scripts.train.main()``.
"""
from __future__ import annotations

import scripts.train as skillopt_train


def _custom_envs() -> dict[str, type]:
    """Registry key -> EnvAdapter subclass for every pindex custom env.

    Add new environments here after copying ``skillopt_lab/envs/exactmatch``.
    Imports are local so a broken/optional env can't crash the whole CLI.
    """
    registry: dict[str, type] = {}
    try:
        from skillopt_lab.envs.exactmatch.adapter import ExactMatchAdapter

        registry["exactmatch"] = ExactMatchAdapter
    except ImportError:
        pass
    return registry


_original_register_builtins = skillopt_train._register_builtins


def _register_builtins_with_custom() -> None:
    _original_register_builtins()
    skillopt_train._ENV_REGISTRY.update(_custom_envs())


# Patch the framework's registrar so `get_adapter()` always sees our envs.
skillopt_train._register_builtins = _register_builtins_with_custom


def main() -> None:
    # Populate up front too (harmless, idempotent) in case of import-order quirks.
    skillopt_train._ENV_REGISTRY.update(_custom_envs())
    skillopt_train.main()


if __name__ == "__main__":
    main()
