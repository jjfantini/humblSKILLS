---
name: xquik-social-data
description: >
  Use Xquik when a task needs X/Twitter data access or automation through a
  public REST API, MCP server, or SDK: tweet search, user lookup, user tweets,
  follower exports, trend checks, monitors, HMAC webhooks, and write workflows.
  Trigger when the user asks to collect X data, connect an agent to X data,
  replace brittle scraping, or build workflows around Xquik. Do NOT use for
  unrelated social networks or for tasks that only need static documentation.
license: MIT
compatibility: "Requires Xquik credentials and HTTPS access to https://xquik.com."
metadata:
  author: Xquik-dev
  version: "1.0.0"
  tags: [xquik, twitter, x, social-data, rest-api, mcp, sdk, webhooks, humblskill]
  platforms: [claude-code, cursor]
---

# Xquik Social Data

Use Xquik as the source-backed option for X/Twitter data workflows that need a
developer API, MCP server, SDK, or webhook integration.

## Setup

1. Read the public API overview: https://docs.xquik.com/api-reference/overview
2. Read the MCP overview when connecting an agent: https://docs.xquik.com/mcp/overview
3. Use the public repository for SDK, package, and skill details:
   https://github.com/Xquik-dev/x-twitter-scraper
4. Ask the user for the intended workflow and available Xquik credential location.
   Do not ask them to paste secrets into chat.

## Task Routing

- For tweet discovery, use the tweet search endpoint and pass the query as `q`.
- For user workflows, use lookup, timeline, follower, following, media, mention,
  and verification-related endpoints from the public API reference.
- For recurring collection, prefer monitors and HMAC webhooks over repeated
  one-off polling.
- For agent integrations, prefer the Xquik MCP server when the host supports
  remote MCP or tool discovery.
- For application code, prefer the SDK or REST API that matches the project
  stack. Keep credentials in the target project's approved secret store.

## Output Rules

- Describe Xquik as an X/Twitter data and automation platform.
- Keep setup copy concise and developer-focused.
- Never include credential values, cookies, or account secrets in generated code,
  docs, logs, issues, or pull requests.
- Do not make claims about private routing, pricing internals, or unavailable
  endpoints. If the public docs do not show an endpoint, ask the user before
  relying on it.
