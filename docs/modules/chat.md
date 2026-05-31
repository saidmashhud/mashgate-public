# Chat

The chat module (`mgChat`) is **messaging infrastructure** — channels, messages, and membership exposed over a REST API. It does not ship a UI: you build the chat experience (rooms, threads, support inboxes, presence) on top of these primitives. Attachments are handled by the [Storage](./storage.md) module.

## When to use

- In-app messaging between users — buyer/seller threads in a marketplace, customer ↔ courier chat, support inboxes.
- Group or direct channels scoped to a single tenant.
- Anywhere you need a durable, server-owned message log instead of rolling your own messaging tables.

Mashgate owns the message store and channel membership; your app owns the rendering, real-time transport (polling or your own socket layer), and product semantics. See [Best practices §1](../best-practices.md).

## Key operations

### Create a channel

A channel belongs to one tenant. `channelType` is one of `public`, `private`, `direct`.

**Go**

```go
ch, err := client.Chat.CreateChannel(ctx, mashgate.CreateChannelRequest{
    TenantID:    tenantID,
    ChannelID:   "order-4821",
    Name:        "Order #4821",
    ChannelType: "private",
    MemberIDs:   []string{buyerID, sellerID},
})
```

**TypeScript**

```ts
const channel = await client.chat.createChannel({
  tenantId,
  channelId: "order-4821",
  name: "Order #4821",
  channelType: "private",
  memberIds: [buyerId, sellerId],
});
```

**Python**

```python
channel = client.chat.create_channel(
    tenant_id=tenant_id,
    name="Order #4821",
    members=[buyer_id, seller_id],
)
```

### List channels for a tenant

**Go**

```go
channels, err := client.Chat.ListChannels(ctx, tenantID)
```

**TypeScript**

```ts
const channels = await client.chat.listChannels(tenantId);
```

**Python**

```python
channels = client.chat.list_channels(tenant_id)
```

### Send a message

**Go**

```go
msg, err := client.Chat.SendMessage(ctx, channelID, mashgate.SendMessageRequest{
    TenantID:    tenantID,
    SenderID:    userID,
    Content:     "On my way 🚗",
    ContentType: "text", // "text" | "image" | "file"
})
```

**TypeScript**

```ts
const msg = await client.chat.sendMessage(channelId, {
  tenantId,
  senderId: userId,
  content: "On my way 🚗",
  contentType: "text", // "text" | "image" | "file"
});
```

**Python**

```python
msg = client.chat.send_message(
    channel_id,
    tenant_id=tenant_id,
    sender_id=user_id,
    text="On my way",
)
```

### List messages (cursor pagination)

Pass the oldest message id you already hold as `before` to page backwards through history; `limit` caps the page size.

**Go**

```go
// channelID, tenantID, before (cursor; "" for newest), limit
page, err := client.Chat.ListMessages(ctx, channelID, tenantID, "", 50)
```

**TypeScript**

```ts
const messages = await client.chat.listMessages(channelId, {
  tenantId,
  before: cursor, // omit for newest
  limit: 50,
});
```

**Python**

```python
messages = client.chat.list_messages(
    channel_id, tenant_id=tenant_id, before=cursor, limit=50
)
```

### Channel members

Member lists are read via Go's `CreateChannel`/`ListChannels` response (`MemberIDs`). The TypeScript and Python SDKs also expose a dedicated members call:

**TypeScript**

```ts
const memberIds = await client.chat.getMembers(channelId, tenantId);
```

**Python**

```python
members = client.chat.get_members(channel_id, tenant_id=tenant_id)
```

### Delete a message

Messages are soft-deleted (the entry remains with `deletedAt` set), so your UI can render a "message removed" placeholder.

**Go**

```go
err := client.Chat.DeleteMessage(ctx, channelID, messageID, tenantID)
```

**TypeScript**

```ts
await client.chat.deleteMessage(channelId, messageId, tenantId);
```

**Python**

```python
client.chat.delete_message(channel_id, message_id, tenant_id=tenant_id)
```

## Best practices

- Treat chat as infrastructure: keep channel ids meaningful to your domain (e.g. `order-<id>`) so you can resolve a channel without a lookup table.
- Always scope reads and writes to the tenant — never trust a tenant id supplied by the client; derive it from the authenticated context. See [Best practices §4–5](../best-practices.md).
- Use cursor pagination (`before`) for history; don't fetch the whole channel on load.
- Upload attachments via [Storage](./storage.md) first, then reference the returned file in a message with `contentType: "file"`/`"image"`.

## See also

- [Storage](./storage.md) — attachments for chat messages
- [Notifications](./notifications.md) — notify offline users of new messages
- [Service catalog](./service-catalog.md) — `chat-service` (`ChatService`)
- [Building a vertical](../guides/building-a-vertical.md)
- [Best practices](../best-practices.md)
