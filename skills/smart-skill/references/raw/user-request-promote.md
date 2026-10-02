# User Request

Date: 2026-10-02

Give forward-deployed engineers (FDEs) a clean way to take a smart skill they
created locally and promote it into a skill registry as a pull request for the
maintainer (Jennings) to review. Keep the bar high.

Scope:

1. Net-new skill addition only. Promoting learnings or behavior changes back
   into an existing skill is a separate path and out of scope for now.
2. Fit how users already install and use humblskills: canonical local skills
   live in `~/.humblskills/skills`. Do not invent a parallel library.
3. Prefer extending existing mechanisms over greenfield.

Updated direction:

1. Put a `promote.sh` script inside smart-skill, the skill users already use
   to create a new smart skill.
2. After create, the agent asks whether to promote the skill:
   - to humblSKILLS (the public repo), or
   - to happySKILLS, if the skill is HappyRobot-specific.
3. Promotion must ensure the skill is tagged and categorized, so its role and
   responsibility are filled in properly and the CLI can organize skills.

Deliverables called out:

- A clear end-to-end path: local create, validate against the quality bar,
  promote, open a PR targeting the right base branch.
- Guardrails that keep quality high (required fields, naming, structure
  checks, lint/tests) without being bureaucratic.
- Documentation of the flow for FDEs.
- Never merge as part of promotion; the maintainer reviews and merges.
