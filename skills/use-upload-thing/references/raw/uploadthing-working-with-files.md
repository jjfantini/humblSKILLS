# Working with Files (captured from https://docs.uploadthing.com/working-with-files)

Source: https://docs.uploadthing.com/working-with-files (captured 2026-06-30)

## Accessing Public Files

> UploadThing serves all files from a CDN at the following URL pattern:
> `https://<APP_ID>.ufs.sh/f/<FILE_KEY>`

If a `customId` was set, `https://<APP_ID>.ufs.sh/f/<CUSTOM_ID>` also works.

Do not use the raw storage-provider URL (e.g. S3) - UploadThing may move objects
between providers/buckets. The legacy `https://utfs.io/f/<FILE_KEY>` is still
supported but not recommended.

## Accessing Private Files

Generate a presigned URL client-side by HMAC-SHA256 signing the ufs URL with the
API key:

```
const apiKey = "sk_live_...";
const url = new URL("https://<APP_ID>.ufs.sh/f/<FILE_KEY>");
const expires = Date.now() + 1000 * 30;            // ms since epoch
url.searchParams.set("expires", String(expires));
const signature = crypto.createHmac("hmac-sha256", apiKey).update(url.href).digest("hex");
url.searchParams.set("signature", `hmac-sha256=${signature}`);
await fetch(url); // 200 OK
```

> You can also request presigned URLs using the `/requestFileAccess` API endpoint
> (see OpenAPI Specification). However, generating URLs client-side is faster as
> it avoids an additional API call.

`use-upload-thing` uses `POST /v6/requestFileAccess` (returns `ufsUrl`) for
robustness - it works for both public and private files with only the API key,
no HMAC implementation needed.

## Other File Operations

Refer to the UTApi server SDK or call the API directly via the OpenAPI spec:
https://docs.uploadthing.com/api-reference/openapi-spec
