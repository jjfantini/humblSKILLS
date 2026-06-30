# Log

Append-only session log. Every session MUST append at least one entry.
Never edit old entries - they are the historical record. Most recent
entries appear at the bottom.

Entry shape:

```
[INGEST|QUERY|LINT <YYYY-MM-DD>] <one-line summary>
  <optional indented detail line(s)>
```

---

[INGEST 2026-06-30] Scaffolded use-upload-thing via scripts/scaffold.sh.
  - Directory layout created: references/{wiki,raw}/, brain meta files, templates
  - Awaiting first raw material and wiki concepts

[INGEST 2026-06-30] Built the v7 UploadThing skill.
  - Captured raw sources: openapi-spec.json (live, info.version 6.10.0), v7-migration,
    uploading-files, working-with-files, ut-api.
  - 4 wiki concepts: auth/api-key, api/rest-endpoints, upload/v7-prepare-upload,
    files/access-and-manage (each cites raw + sets command:).
  - Scripts: lib.sh, auth.sh, upload.sh (v7 prepareUpload + PUT), download.sh
    (requestFileAccess), api.sh (generic), lint.sh.
  - Decision: use /v7/prepareUpload + /v6/requestFileAccess (REST, key-only) instead of
    self-signed Sqids/HMAC ingest URLs - no crypto in bash, works with the sk_live key.
  - Verified live: auth.sh --check returns appId 2zzhd7n3m2 (defaultACL public-read,
    allowACLOverride false).

[LINT 2026-06-30] 4 wiki, 5 raw. Hard: 0, Soft: 0. Regenerated _index.md.
