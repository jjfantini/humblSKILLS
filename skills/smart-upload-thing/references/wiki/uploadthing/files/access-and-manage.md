---
title: "Access, Download, and Manage Uploaded Files"
context: uploadthing
category: files
concept: access-and-manage
description: "ufs.sh URL pattern, requestFileAccess for public/private downloads, and list/delete/rename/ACL management"
tags: download, files, ufs, requestFileAccess, private, delete, rename, acl, usage
sources:
  - "references/raw/uploadthing-working-with-files.md"
  - "references/raw/uploadthing-ut-api.md"
  - "references/raw/uploadthing-openapi-spec.json"
last_ingested: 2026-06-30
command: scripts/download.sh
---

## Access, Download, and Manage Uploaded Files

Public files are served from `https://<APP_ID>.ufs.sh/f/<FILE_KEY>` (or
`.../f/<CUSTOM_ID>`). Private files need a presigned URL. `download.sh` handles
both by calling `POST /v6/requestFileAccess`, which returns a `ufsUrl` valid for
public files and signed for private ones.

**Incorrect (raw storage-provider URL is not stable):**

```bash
# DON'T link to the underlying S3 object - UploadThing may relocate it.
curl -O https://bucket.s3.us-east-1.amazonaws.com/<FILE_KEY>
```

**Correct (by key, works public + private):**

```bash
bash scripts/download.sh <FILE_KEY> -o out.png
# resolves via /v6/requestFileAccess -> ufsUrl -> curl -L

# or download a known URL directly:
bash scripts/download.sh "https://<appId>.ufs.sh/f/<key>" -o out.png
```

Use `--expires-in SECONDS` to bound how long the presigned URL is valid (max 7 days / 604800).

### Managing files (via api.sh)

```bash
bash scripts/api.sh v6/listFiles '{"limit":50}'
bash scripts/api.sh v6/renameFiles '{"updates":[{"fileKey":"KEY","newName":"renamed.png"}]}'
bash scripts/api.sh v6/updateACL  '{"updates":[{"fileKey":"KEY","acl":"private"}]}'
bash scripts/api.sh v6/getUsageInfo
bash scripts/api.sh v6/deleteFiles '{"fileKeys":["KEY"]}'   # cleanup
```

`listFiles` is for admin/debug/one-time sync - the docs recommend storing file
metadata in your own database for the app's core data flow.

## Command

```bash
bash scripts/download.sh <fileKeyOrUrl> [-o OUTPUT] [--expires-in SECONDS]
```

Reference: https://docs.uploadthing.com/working-with-files
