# Storage

The storage module (`mgStorage`) is **S3-compatible object storage** for files: avatars, listing photos, documents, chat and mail attachments. Uploads go directly to object storage over presigned URLs — your backend never proxies the bytes — and downloads are served via presigned (or CDN) URLs.

## When to use

- User-generated media: profile avatars, marketplace listing images, message attachments.
- Any binary blob your app needs to store and serve per tenant.
- As the attachment backend for [Chat](./chat.md) and [Notifications (mail)](./notifications.md).

## Key operations

### 1. Request a presigned upload URL

Ask Mashgate for a short-lived S3 upload URL, then `PUT` the file bytes directly to it from the client or server. The response carries the `fileId` you persist, the `uploadUrl`, the object `key`, and an `expiresAt`.

**Go**

```go
up, err := client.Storage.GenerateUploadURL(ctx, mashgate.GenerateUploadURLRequest{
    TenantID: tenantID,
    Filename: "avatar.png",
    MimeType: "image/png",
})
// up.FileID, up.UploadURL, up.Key, up.ExpiresAt
```

**TypeScript**

```ts
const up = await client.storage.generateUploadUrl({
  tenantId,
  filename: "avatar.png",
  mimeType: "image/png",
});
// up.fileId, up.uploadUrl, up.key, up.expiresAt
```

**Python**

```python
up = client.storage.generate_upload_url(
    tenant_id=tenant_id,
    filename="avatar.png",
    content_type="image/png",
)
```

### 2. Upload the bytes (standard S3 PUT)

The presigned URL is plain S3 — upload with any HTTP client; no Mashgate auth header is needed on this request.

```ts
await fetch(up.uploadUrl, {
  method: "PUT",
  headers: { "Content-Type": "image/png" },
  body: fileBytes,
});
```

### 3. Get a download URL

Returns a presigned download URL (or CDN URL) for the stored object.

**Go**

```go
url, err := client.Storage.GetDownloadURL(ctx, fileID, tenantID)
```

**TypeScript**

```ts
const url = await client.storage.getDownloadUrl(fileId, tenantId);
```

**Python**

```python
res = client.storage.get_download_url(file_id, tenant_id=tenant_id)
```

### List files

**Go**

```go
files, err := client.Storage.ListFiles(ctx, tenantID)
// each: FileID, TenantID, Key, Size, LastModified
```

**TypeScript**

```ts
const files = await client.storage.listFiles(tenantId);
```

**Python**

```python
files = client.storage.list_files(tenant_id)
```

### Delete a file

**Go**

```go
err := client.Storage.DeleteFile(ctx, fileID, tenantID)
```

**TypeScript**

```ts
await client.storage.deleteFile(fileId, tenantId);
```

**Python**

```python
client.storage.delete_file(file_id, tenant_id=tenant_id)
```

## Best practices

- Use the presigned-upload flow — don't proxy file bytes through your own backend; let the client `PUT` straight to S3.
- Persist the returned `fileId` (not the URL): download/CDN URLs are short-lived and regenerated on demand.
- Scope every call to the tenant and never trust a client-supplied tenant id. See [Best practices §4](../best-practices.md).
- For chat/mail attachments, upload first and reference the `fileId` when sending the message.

## See also

- [Chat](./chat.md) — attachments in messages
- [Notifications (mail)](./notifications.md) — `attachment_ids` reference pre-uploaded files
- [Service catalog](./service-catalog.md) — `storage-service` (`StorageService`)
- [Best practices](../best-practices.md)
