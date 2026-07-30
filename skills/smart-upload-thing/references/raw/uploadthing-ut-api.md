# UTApi server SDK reference (captured from https://docs.uploadthing.com/api-reference/ut-api)

Source: https://docs.uploadthing.com/api-reference/ut-api (captured 2026-06-30)

The UTApi is the server-side helper. It is "basically just a REST API but
better." Mirrors the REST endpoints this skill calls directly. Key methods and
the REST endpoints they map to:

| UTApi method        | REST endpoint                | Notes |
|---------------------|------------------------------|-------|
| `uploadFiles`       | `/v6/uploadFiles` (v6) / v7 ingest | v7 uses prepareUpload + PUT |
| `deleteFiles`       | `/v6/deleteFiles`            | by `fileKeys` or `customIds` |
| `listFiles`         | `/v6/listFiles`              | `{limit, offset}`; admin/debug use only |
| `renameFiles`       | `/v6/renameFiles`            | `{updates:[{fileKey|customId, newName}]}` |
| `generateSignedURL` | (client-side HMAC, no fetch) | private files; since 7.5 |
| `getSignedURL`      | `/v6/requestFileAccess`      | private files; fetch-based |
| `updateACL`         | `/v6/updateACL`              | `{updates:[{fileKey|customId, acl}]}` |
| `getAppInfo`        | `/v7/getAppInfo`             | `{appId, defaultACL, allowACLOverride}` |

Config / auth:
- `token` (default `env.UPLOADTHING_TOKEN`, since 7.0): the v7 base64 token.
- `apiUrl` defaults to `https://api.uploadthing.com`.
- Env config uses `UPLOADTHING_*` constant-case names.

Important guidance from the docs:
> Please note that external API calls will almost always be slower than querying
> your own database. We recommend storing the file data you need in your own
> database ... instead of relying on the API for your application's core data flow.

`listFiles` is best for administrative tasks, one-time sync, or debugging.

Access control:
- ACL is `public-read` or `private`.
- Per-request ACL overrides only apply if the app allows them
  (`allowACLOverride` from getAppInfo).

Live endpoint catalog: https://docs.uploadthing.com/api-reference/openapi-spec
