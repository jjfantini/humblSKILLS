---
title: "UploadThing REST Endpoint Catalog"
context: uploadthing
category: api
concept: rest-endpoints
description: "Base URL, auth, and every REST endpoint (v7 + latest v6) with request shapes, callable via api.sh"
tags: rest, endpoints, api, v7, v6, listFiles, deleteFiles, renameFiles, updateACL, getUsageInfo
sources:
  - "references/raw/uploadthing-openapi-spec.json"
  - "references/raw/uploadthing-ut-api.md"
last_ingested: 2026-06-30
command: scripts/api.sh
---

## UploadThing REST Endpoint Catalog

Base URL `https://api.uploadthing.com`. Auth header `x-uploadthing-api-key`. All endpoints are `POST` with a JSON body. Versioning is path-based: the `/v7/*` endpoints are current; management endpoints have no v7 variant, so `/v6/*` IS their latest version (the live spec reports `info.version` 6.10.0).

`scripts/api.sh <path> [json]` calls any endpoint and pretty-prints the JSON. It is the escape hatch for all functionality.

### Endpoints

| Path | Body | Returns |
|------|------|---------|
| `v7/getAppInfo` | `{}` | `{appId, defaultACL, allowACLOverride}` |
| `v7/prepareUpload` | `{fileName, fileSize, fileType?, customId?, contentDisposition?, acl?, expiresIn?}` | `{key, url}` (see upload concept) |
| `v6/listFiles` | `{limit?, offset?}` | `{hasMore, files:[{key, name, size, status, customId, uploadedAt}]}` |
| `v6/deleteFiles` | `{fileKeys:[...]}` or `{customIds:[...]}` | `{success, deletedCount}` |
| `v6/renameFiles` | `{updates:[{fileKey|customId, newName}]}` | `{}` |
| `v6/updateACL` | `{updates:[{fileKey|customId, acl}]}` | `{}` |
| `v6/getUsageInfo` | `{}` | usage stats (bytes, file count, limits) |
| `v6/requestFileAccess` | `{fileKey|customId, expiresIn?}` | `{ufsUrl, url}` (see files concept) |

**Incorrect (wrong field name - deleteFiles ignores it):**

```bash
bash scripts/api.sh v6/deleteFiles '{"keys":["KEY"]}'   # must be "fileKeys"
```

**Correct:**

```bash
bash scripts/api.sh v6/listFiles '{"limit":10}'
bash scripts/api.sh v6/deleteFiles '{"fileKeys":["KEY_1","KEY_2"]}'
bash scripts/api.sh v6/updateACL '{"updates":[{"fileKey":"KEY","acl":"private"}]}'
bash scripts/api.sh v7/getAppInfo
```

For the always-current schema, fetch the live spec (also cached at `references/raw/uploadthing-openapi-spec.json`): https://docs.uploadthing.com/api-reference/openapi-spec

## Command

```bash
bash scripts/api.sh <path> '<json>'
```
