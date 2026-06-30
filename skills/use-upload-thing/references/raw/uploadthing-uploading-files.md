# Uploading Files (captured from https://docs.uploadthing.com/uploading-files)

Source: https://docs.uploadthing.com/uploading-files (captured 2026-06-30)

Two ways to upload: client-side (via File Routes + adapters) and server-side.
For a server-side / SDK-less consumer, you generate presigned URLs and PUT files
to the regional ingest host.

## Generating presigned URLs (self-signed, advanced)

File key = Sqids(appId, { minLength: 12 }) concatenated with a url-safe file
seed (e.g. base64 of a chosen seed). Reference implementations exist for JS,
Python, PHP, Go, Rust.

> If you struggle to implement it for your language, you can also request one
> from the `/v7/prepareUpload` REST endpoint. Keep in mind that this adds extra
> latency to your uploads.

Upload URL format: `https://{{ REGION_ALIAS }}.ingest.uploadthing.com/{FILE_KEY}`

Self-signed presigned URL query params:

```
const searchParams = new URLSearchParams({
  // Required
  expires: Date.now() + 60 * 60 * 1000,   // ms since epoch
  "x-ut-identifier": "MY_APP_ID",
  "x-ut-file-name": "my-file.png",
  "x-ut-file-size": 131072,
  "x-ut-slug": "MY_FILE_ROUTE",            // omit for server-side uploads
  // Optional
  "x-ut-file-type": "image/png",
  "x-ut-custom-id": "MY_CUSTOM_ID",
  "x-ut-content-disposition": "inline",
  "x-ut-acl": "public-read",
});
const url = new URL(`https://{{ REGION_ALIAS }}.ingest.uploadthing.com/${fileKey}`);
url.search = searchParams.toString();
const signature = `hmac-sha256=${hmacSha256(url, apiKey)}`;
url.searchParams.append("signature", signature);
```

## Uploading the file

> Uploading the files is as simple as submitting a PUT request to the signed URL.

```
const formData = new FormData();
formData.append("file", file);
await fetch(presigned.url, { method: "PUT", body: formData });
```

Resumable: send a HEAD to the presigned URL to read `x-ut-range-start`, then PUT
with a `Range: bytes=<start>-` header and the sliced body.

## Server-side uploads

> Generating presigned URLs is the same for server-side uploads as for
> client-side uploads. The only difference is that you do not have to include
> the `x-ut-slug` search parameter. There is no need to register the upload or
> handle callbacks.

This is why `use-upload-thing` uses `/v7/prepareUpload` (no slug) + PUT: it is a
server-side upload that needs only the API key, no Sqids/HMAC implementation.

Live endpoint catalog: https://docs.uploadthing.com/api-reference/openapi-spec
