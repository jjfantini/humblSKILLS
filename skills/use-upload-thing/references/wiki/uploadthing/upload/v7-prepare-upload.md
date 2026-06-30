---
title: "Upload a File with the v7 prepareUpload Flow"
context: uploadthing
category: upload
concept: v7-prepare-upload
description: "The v7 two-call upload: POST /v7/prepareUpload for a presigned ingest URL, then PUT the file"
tags: upload, v7, prepareUpload, ingest, presigned, PUT
sources:
  - "references/raw/uploadthing-uploading-files.md"
  - "references/raw/uploadthing-openapi-spec.json"
last_ingested: 2026-06-30
command: scripts/upload.sh
---

## Upload a File with the v7 prepareUpload Flow

v7 generates presigned URLs without the v6 server round-trip. For a server-side
consumer the simplest path is `POST /v7/prepareUpload` (no file-route slug),
then `PUT` the bytes to the returned ingest URL. This avoids implementing Sqids
key generation and HMAC signing by hand.

### Two calls

1. `POST /v7/prepareUpload` with `{fileName, fileSize, fileType?, contentDisposition?, acl?, customId?, expiresIn?}` -> `{key, url}`. `fileName` and `fileSize` (bytes) are required.
2. `PUT` the file to `url` as multipart form-data with field name `file`.

**Incorrect (the prepareUpload response is NOT the finished upload):**

```bash
# Requesting the presigned URL alone leaves the file in "uploading" forever.
bash scripts/api.sh v7/prepareUpload '{"fileName":"a.png","fileSize":1024}'
# ... and then never PUTting the bytes. The file URL will not resolve.
```

**Correct (use upload.sh, which does both calls):**

```bash
bash scripts/upload.sh ./photo.png --content-disposition inline
# -> {"key":"...","ingestUrl":"https://<region>.ingest.uploadthing.com/<key>?...","ufsUrl":"https://<appId>.ufs.sh/f/<key>"}
```

`upload.sh` computes `fileSize` with `wc -c`, detects `fileType` via `file --mime-type`, calls prepareUpload, then `curl -X PUT -F "file=@<path>"` to the ingest URL, and finally builds the public `ufsUrl` from `getAppInfo`.

Notes:
- `acl` is honored only if the app's `allowACLOverride` is `true`; otherwise the app `defaultACL` applies (check with `auth.sh --check`).
- `expiresIn` (seconds) makes the object auto-delete - handy for throwaway test uploads.

## Command

```bash
bash scripts/upload.sh <path> [--name N] [--acl public-read|private] \
  [--content-disposition inline|attachment] [--custom-id ID] [--expires-in SECONDS]
```

Reference: https://docs.uploadthing.com/uploading-files
