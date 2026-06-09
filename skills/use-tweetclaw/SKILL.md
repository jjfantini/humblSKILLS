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
  version: "1.0.0"
  tags: [tweetclaw, openclaw, twitter, x, social-media, automation]
  platforms: [claude-code, cursor]
---

# Use TweetClaw

Use TweetClaw as the execution layer for approved X/Twitter workflows in
OpenClaw. Keep the agent responsible for planning, review, and approval; keep
TweetClaw responsible for the X/Twitter action or data fetch.

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

## Setup Check

Before using TweetClaw, verify that the workspace has the plugin installed and
configured through secure OpenClaw config.

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

## Workflow

1. Restate the requested X/Twitter job and identify whether it is read-only or
   write-like.
2. For read-only work, collect the minimum query, user, tweet URL, or time range
   needed to run the lookup.
3. For write-like work, draft the exact final action, show it to the operator,
   and wait for explicit approval before calling TweetClaw.
4. Run the TweetClaw tool that matches the job.
5. Summarize only the relevant result, including source tweet URLs or IDs when
   useful for verification.

## Prompt Patterns

```text
Use TweetClaw to search recent tweets about this launch. Return the strongest
5 examples with URLs and a one-line relevance note for each.
```

```text
Use TweetClaw to inspect replies to this tweet and draft one reply. Do not post
until I approve the exact final text.
```

```text
Use TweetClaw to export followers for this account, then summarize notable
developer-tool, media, and agency accounts.
```

```text
Use TweetClaw to prepare this tweet with the attached media. Show me the final
text and media list before posting.
```
