---
name: use-upload-thing
description: >
  Drive the UploadThing REST API end-to-end from the shell: authenticate with an
  API key, upload files (v7 prepareUpload + ingest PUT), download by key or URL,
  and list, delete, rename, set ACLs, or read usage. Use when the user wants to
  upload/download/manage files on UploadThing, mentions UploadThing, ufs.sh,
  prepareUpload, an sk_live key, or "put this file on UploadThing". Also guides
  first-time API-key setup. Do NOT use for building an app's File Router / React
  upload components (that is the TS SDK), or for non-UploadThing storage (S3, GCS).
license: MIT
compatibility: "Requires bash, curl, jq, and network access to api.uploadthing.com. Needs an UploadThing API key via UPLOADTHING_API_KEY (or UPLOADTHING_TOKEN), or run scripts/auth.sh --save."
metadata:
  author: jjfantini
  version: "1.0.3"
  category: development
  tags: [uploadthing, file-upload, storage, rest-api, ufs, humblskill]
  platforms: [claude-code, cursor, codex]
  preserve:
    - references/raw/
    - references/wiki/
    - references/decisions.md
    - references/log.md
    - references/patterns.md
---

# Use UploadThing

Talk to the UploadThing REST API (`https://api.uploadthing.com`) from the shell. Authenticate with an API key, upload via the v7 `prepareUpload` flow, download public or private files, and run every management endpoint. One API key maps to one UploadThing project.

## Brain Protocol (read BEFORE doing anything)

1. `references/_index.md`       - what this skill knows (map)
2. `references/patterns.md`     - what worked, with numbers
3. `references/decisions.md`    - past reasoning, don't repeat mistakes
4. `references/log.md`          - last 5 session entries
5. Relevant `references/wiki/uploadthing/<category>/` concepts per task

After completing work, UPDATE the brain:
- Performance data (if reported) -> `patterns.md`
- Non-obvious decisions -> `decisions.md`
- Session summary (always) -> append to `log.md`
- Run `scripts/lint.sh` to regenerate `_index.md` and verify structure

_Full spec: `references/_brain.md`._

## Workflow

1. **Authenticate.** Run `bash scripts/auth.sh --check`. If it reports no key, follow the printed dashboard guidance, then `export UPLOADTHING_API_KEY=sk_live_...` (session) or persist once with `printf '%s' "$KEY" | bash scripts/auth.sh --save` (writes a `0600` file outside the repo). The key is never printed or passed as an argument.
2. **Act.** Use the script matching the task:
   - Upload: `bash scripts/upload.sh <path> [--acl ...] [--expires-in N]`
   - Download: `bash scripts/download.sh <fileKeyOrUrl> [-o out]`
   - Anything else: `bash scripts/api.sh <path> '<json>'` (list/delete/rename/updateACL/getUsageInfo/getAppInfo).
3. **Confirm + clean up.** Verify results (`api.sh v6/listFiles`) and delete throwaway test files (`api.sh v6/deleteFiles '{"fileKeys":["KEY"]}'`).

## Auth and security

- Resolution order: `UPLOADTHING_API_KEY` env -> `UPLOADTHING_TOKEN` (decode `.apiKey`) -> `0600` config file.
- The key is never echoed, logged, embedded in a URL, or passed as a process arg. `--save` reads it from stdin.
- The bare `sk_live_...` key authenticates the entire REST surface, including the `/v7` endpoints. The base64 `UPLOADTHING_TOKEN` is only required by the TS SDK; this skill does not need it.
- See `references/wiki/uploadthing/auth/api-key.md`.

## Scripts

- `scripts/auth.sh` - `--status` / `--check` (`/v7/getAppInfo`) / `--save`.
- `scripts/upload.sh` - v7 `prepareUpload` then `PUT` to the ingest URL; prints `{key, ingestUrl, ufsUrl}`.
- `scripts/download.sh` - resolve a key via `/v6/requestFileAccess` (public + private) then download; or fetch a URL directly.
- `scripts/api.sh` - generic `POST` to any endpoint; the escape hatch for all functionality.
- `scripts/lib.sh` - sourced helpers (key resolution, `ut_api`, preflight). Not run directly.
- `scripts/lint.sh` - brain health check; regenerates `_index.md`.

## How to Use

**Endpoint catalog + request shapes:** `references/wiki/uploadthing/api/rest-endpoints.md`.
**Authenticate / store the key securely:** `references/wiki/uploadthing/auth/api-key.md`.
**Upload a file (v7):** `references/wiki/uploadthing/upload/v7-prepare-upload.md`.
**Download / access / manage files:** `references/wiki/uploadthing/files/access-and-manage.md`.
**Brain protocol, naming, sources contract:** `references/_brain.md`. **Wiki shape:** `references/_template.md`.

Always-current API: https://docs.uploadthing.com/api-reference/openapi-spec

## Examples

### Example 1: "Upload this logo to UploadThing and give me the URL"

Actions:
1. `bash scripts/auth.sh --check` (confirm key + project).
2. `bash scripts/upload.sh ./logo.png` -> capture `ufsUrl`.
3. Return the `ufsUrl` to the user.

Result: file stored; public `https://<appId>.ufs.sh/f/<key>` URL returned.

### Example 2: "Download file key abc123 and then delete it"

Actions:
1. `bash scripts/download.sh abc123 -o abc123.bin`.
2. `bash scripts/api.sh v6/deleteFiles '{"fileKeys":["abc123"]}'`.
3. Confirm with `bash scripts/api.sh v6/listFiles '{"limit":5}'`.

Result: file downloaded locally, then removed from the project.

## Troubleshooting

**`auth FAILED` / HTTP 401:** the key is missing or wrong. Run `auth.sh --status`; re-export `UPLOADTHING_API_KEY` or re-`--save`. Each key is tied to one project.
**Uploaded file URL stays "uploading" / 404:** prepareUpload was called but the `PUT` never happened - use `upload.sh` (does both), not a raw `api.sh v7/prepareUpload`.
**`--acl` ignored:** the app has `allowACLOverride:false` (see `auth.sh --check`); the app default ACL applies.
**`jq`/`curl` not found:** install them; the scripts hard-require both.

## Success Signals

- `auth.sh --check` prints a valid `appId`.
- `upload.sh` then `download.sh` round-trips a file whose bytes match the original.
- `api.sh v6/deleteFiles` removes test files (verified absent via `listFiles`).
- `scripts/lint.sh` exits 0; the key never appears in any output.
