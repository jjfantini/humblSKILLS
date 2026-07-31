# Roadmap

This page is a lightweight stub. Track larger direction in [GitHub Issues](https://github.com/jjfantini/humblSKILLS/issues) and the main [README](https://github.com/jjfantini/humblSKILLS).

Planned doc improvements:

- Auto-generated **CLI reference** (Cobra → Markdown → MkDocs) in a follow-up pass. The [Quickstart](getting_started/quickstart.md#command-reference-at-a-glance) carries a hand-maintained command table until then; `--help` on any command is always authoritative.

Already generated at build time:

- The [skill catalog](skills.md) is rendered from `registry.json` on every docs build, so it cannot drift from the released registry.
