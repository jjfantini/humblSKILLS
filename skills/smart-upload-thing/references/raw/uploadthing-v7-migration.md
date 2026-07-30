# Migrate from v6 to v7 (captured from https://docs.uploadthing.com/v7)

Source: https://docs.uploadthing.com/v7 (captured 2026-06-30)

Version 7 is backed by new infrastructure and a near-full rewrite of the internals:
client bundle reduced, resumable uploads supported, fewer round-trips to the
UploadThing server, polling removed.

## UPLOADTHING_SECRET is now UPLOADTHING_TOKEN (BREAKING)

In prior versions, presigned URLs were generated on UploadThing's servers and
fetched from your backend via authenticated API calls. In v7, presigned-URL
generation moved to your server, so the server must know the app id and region.
These are merged into a single `UPLOADTHING_TOKEN` environment variable.

> The token is a base64 encoded JSON object that contains information such as
> your app id, the app region, as well as the API key. The token is available
> in the UploadThing Dashboard under `API Keys` by selecting the `V7` tab.

Config object keys changed accordingly (token replaces uploadthingSecret/uploadthingAppId).

## The UploadThing REST API has moved (MISC)

> The UploadThing REST API has been moved to a separate domain,
> `api.uploadthing.com`. The old API at `uploadthing.com/api` now redirects to
> the new API.
>
> The new API has explicit path-based versioning, meaning the
> `x-uploadthing-version` header is no longer required.

```
// Old API
curl -X POST https://uploadthing.com/api/listFiles \
  -H 'x-uploadthing-version: 6.12.0' ...

// New API
curl -X POST https://api.uploadthing.com/v6/listFiles ...
```

## Notes relevant to a REST/CLI consumer

- The REST API authenticates with the `x-uploadthing-api-key` header (an
  `sk_live_...` key). The base64 `UPLOADTHING_TOKEN` is what the TS SDK uses to
  self-sign client-side presigned URLs; for the REST API the bare API key is
  sufficient, including for the `/v7/*` endpoints.
- Upload moved to: self-generated presigned ingest URLs OR the
  `/v7/prepareUpload` REST endpoint (then PUT the file to the returned URL).
- Other breaking changes (createRouteHandler, genUploader, log levels,
  skipPolling -> awaitServerData, removed deprecations) are SDK-only and do not
  affect a REST consumer.

Live, always-current endpoint catalog: https://docs.uploadthing.com/api-reference/openapi-spec
