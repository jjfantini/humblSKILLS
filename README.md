# humblSKILLS

A personal skill registry and a single-binary Go CLI (`humblskills`) that installs
[agentskills.io](https://agentskills.io)-format skills into whichever agent platform
you use — Claude Code, Cursor, Codex, and friends.

## Mission

Two things in one repo:

1. **Skill registry** — a monorepo of agent skills authored in the agentskills.io
   format with light humblSKILLS frontmatter extensions (`requires`, `platforms`,
   `post_install`, `tags`).
2. **`humblskills` CLI** — fetches a skill directory and drops it in the right
   place for your agent platform. Zero servers, zero accounts, zero telemetry.

## Quickstart (coming with v0.1)

```
brew install jjfantini/humbl/humblskills
humblskills add                    # interactive picker
humblskills list
```

Install methods planned for v0.1:

1. Homebrew tap: `brew install jjfantini/humbl/humblskills`
2. Shell installer: `curl -fsSL https://raw.githubusercontent.com/jjfantini/humblSKILLS/main/scripts/install.sh | sh`
3. `go install github.com/jjfantini/humblSKILLS/cli@latest`
4. Direct download from GitHub Releases

## Status

Pre-v0.1. Implementation is phased: registry pipeline → CLI foundation →
install path → polish → distribution.

## License

Content is licensed under [CC-BY-4.0](LICENSE). If Go source code licensing
becomes a concern later, the CLI code under `cli/` may be dual-licensed MIT —
but that has not been done yet.
