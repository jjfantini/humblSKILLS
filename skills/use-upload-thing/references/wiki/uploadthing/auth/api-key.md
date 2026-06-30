---
title: "Authenticate to an UploadThing Project"
context: uploadthing
category: auth
concept: api-key
description: "How the API key maps to one project, the x-uploadthing-api-key header, and secure key resolution/storage"
tags: auth, api-key, token, x-uploadthing-api-key, security, getAppInfo
sources:
  - "references/raw/uploadthing-v7-migration.md"
  - "references/raw/uploadthing-ut-api.md"
  - "references/raw/uploadthing-openapi-spec.json"
last_ingested: 2026-06-30
command: scripts/auth.sh
---

## Authenticate to an UploadThing Project

One API key maps to exactly one UploadThing app/project. Every REST call - including the `/v7/*` endpoints - authenticates with the `x-uploadthing-api-key` header. The bare `sk_live_...` key is sufficient; the base64 `UPLOADTHING_TOKEN` is only needed by the TS SDK for client-side signing.

### Key resolution order

`scripts/lib.sh` resolves the key without ever printing it:

1. `$UPLOADTHING_API_KEY` (env; never persisted by this skill)
2. `$UPLOADTHING_TOKEN` (v7 base64 token; its `.apiKey` is decoded)
3. `$UT_CONFIG_FILE` (default `${XDG_CONFIG_HOME:-$HOME/.config}/humblskills/uploadthing.env`, mode `0600`)

**Incorrect (leaks the secret into shell history / process args):**

```bash
# DON'T: key visible in `ps`, history, and any command log
bash scripts/auth.sh --save sk_live_xxxxxxxx
curl -H "x-uploadthing-api-key: sk_live_xxxx" ...   # ad-hoc, unmasked
```

**Correct (env for the session, or stdin to persist at 0600):**

```bash
export UPLOADTHING_API_KEY=sk_live_...      # session only
bash scripts/auth.sh --check                # validate via POST /v7/getAppInfo

# or persist once (read from stdin, never an arg), then forget the env var:
printf '%s' "$UPLOADTHING_API_KEY" | bash scripts/auth.sh --save
```

`--check` calls `POST /v7/getAppInfo` and prints `appId`, `defaultACL`, and `allowACLOverride`. Note: if `allowACLOverride` is `false`, per-upload `--acl` is ignored and the app default applies.

To get a key: https://uploadthing.com/dashboard -> select app -> `API Keys` tab.

## Command

```bash
bash scripts/auth.sh --status   # masked source of the resolved key
bash scripts/auth.sh --check    # validate against /v7/getAppInfo
printf '%s' "$KEY" | bash scripts/auth.sh --save   # persist to 0600 file
```
