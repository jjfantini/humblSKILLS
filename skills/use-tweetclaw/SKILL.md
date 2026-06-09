---
name: use-tweetclaw
description: >
  Use TweetClaw when the user needs approved X/Twitter work through OpenClaw:
  scrape tweets, search tweets or replies, look up users, export followers,
  upload or download media, handle direct messages, monitor tweets, configure
  webhooks, run giveaway draws, or post reviewed tweets and replies. Do NOT
  use for generic social strategy, unsupported anonymous scraping, credential
  handling outside secure config, or unattended publishing without explicit
  operator approval.
license: MIT
compatibility: Requires a configured OpenClaw workspace with the TweetClaw plugin installed and any needed X/Twitter account permissions already approved.
metadata:
  author: Xquik
  version: "1.1.0"
  tags: [tweetclaw, openclaw, twitter, x, social-media, automation, smart-skill]
  platforms: [claude-code, cursor]
  preserve:
    - references/decisions.md
    - references/log.md
    - references/patterns.md
---

# Use TweetClaw

Use TweetClaw as the execution layer for approved X/Twitter workflows in
OpenClaw. Keep the agent responsible for planning, review, and approval; keep
TweetClaw responsible for the X/Twitter action or data fetch.

## Brain Protocol

Before using this skill, read:

1. `references/_index.md`
2. `references/patterns.md`
3. `references/decisions.md`
4. The last 5 entries in `references/log.md`
5. Relevant concepts under `references/wiki/tweetclaw/`

After the task, append a short entry to `references/log.md`. If the task
produces measured outcomes, append them to `references/patterns.md`. If a
non-obvious approval or routing choice was made, append it to
`references/decisions.md`. Run `scripts/lint.sh` after wiki changes.

Full brain rules: `references/_brain.md`.

## When to Use

- Scrape tweets or search tweets for evidence, trend checks, support context,
  or social research.
- Search tweet replies, inspect reply context, or prepare a reviewed reply.
- Look up users, export followers, or collect account context.
- Upload media, download media, or attach media to a reviewed post.
- Send or read direct messages only when the user explicitly asks.
- Monitor tweets, configure webhooks, or run giveaway draws.
- Post tweets or replies only after the user has reviewed the exact final text.

## When Not to Use

- Do not use TweetClaw for generic brainstorming, brand strategy, or copywriting
  before there is a concrete X/Twitter action or lookup.
- Do not run anonymous scraping or bypass platform/account permissions.
- Do not paste API keys, cookies, or account secrets into prompts, issues,
  logs, or skill files.
- Do not publish, reply, DM, upload media, start monitors, or configure webhooks
  without explicit operator approval for the exact action.

## How to Use

For approved-action routing, read
`references/wiki/tweetclaw/workflows/approved-actions.md`.

Helpful public references:

- TweetClaw repo: <https://github.com/Xquik-dev/tweetclaw>
- TweetClaw npm registry metadata: <https://registry.npmjs.org/@xquik%2ftweetclaw>
- TweetClaw packaged skill: <https://github.com/Xquik-dev/tweetclaw/tree/master/skills/tweetclaw>

If the plugin is missing, ask the operator to install the current npm package:

```bash
openclaw plugins install npm:@xquik/tweetclaw
openclaw plugins update tweetclaw
openclaw plugins inspect tweetclaw --runtime --json
```

## Examples

```text
Use TweetClaw to search recent tweets about this launch. Return the strongest
5 examples with URLs and a one-line relevance note for each.
```

```text
Use TweetClaw to inspect replies to this tweet and draft one reply. Do not post
until I approve the exact final text.
```
