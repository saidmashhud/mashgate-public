# Notifications

Two SDK resources cover outbound messaging:

- **`notify`** (`mgNotify`) — transactional messages across channels: email, SMS, Telegram, plus reusable templates.
- **`mail`** — a full mailbox capability (a Mashgate core primitive, ADR-0019): per-subject mailboxes, sendable/receivable messages, folders, and tenant mail domains with DKIM/SPF/DMARC.

Use `notify` for fire-and-forget transactional sends (OTP codes, receipts, alerts). Use `mail` when your product *is* email — inboxes, threads, and domain ownership.

## When to use

- **notify** — send an SMS OTP, a templated "order shipped" email, or a Telegram alert; manage reusable templates; read delivery logs.
- **mail** — give each user a real mailbox, list/read/send/organize messages, and register sending domains for a tenant.

## Key operations — notify

### Send an SMS

Either pass raw `text`, or a `templateKey` with `vars` to render a stored template.

**Go**

```go
log, err := client.Notify.SendSms(ctx, mashgate.SendSmsRequest{
    TenantID:    tenantID,
    To:          "+998901234567",
    TemplateKey: "otp",
    Vars:        map[string]string{"code": "4821"},
})
```

**TypeScript**

```ts
const log = await client.notify.sendSms({
  tenantId,
  to: "+998901234567",
  templateKey: "otp",
  vars: { code: "4821" },
});
```

**Python**

```python
log = client.notify.send_sms(
    tenant_id=tenant_id,
    to="+998901234567",
    text="Your code is 4821",
)
```

### Send an email

The Go and TypeScript SDKs send via a required `templateKey`; the Python SDK also accepts an inline `subject` + `body_html`/`body_text`.

**Go**

```go
log, err := client.Notify.SendEmail(ctx, mashgate.SendEmailRequest{
    TenantID:    tenantID,
    To:          "user@example.com",
    TemplateKey: "order-shipped",
    Vars:        map[string]string{"orderId": "4821"},
})
```

**TypeScript**

```ts
const log = await client.notify.sendEmail({
  tenantId,
  to: "user@example.com",
  templateKey: "order-shipped",
  vars: { orderId: "4821" },
});
```

**Python**

```python
log = client.notify.send_email(
    tenant_id=tenant_id,
    to="user@example.com",
    subject="Your order shipped",
    body_html="<p>Order #4821 is on the way.</p>",
)
```

### Send a Telegram message (Go)

Telegram delivery is exposed on the Go SDK. The chat-id pairing flow (resolving a Telegram `chatId` from a `/start <token>`) stays in your app; notify-service is delivery-only.

```go
res, err := client.Notify.SendTelegram(ctx, mashgate.SendTelegramRequest{
    TenantID: tenantID,
    ChatID:   "123456789",
    Text:     "Your order is ready",
})
// res.Status is "sent" | "failed"
```

### Templates

**Go**

```go
tmpl, err := client.Notify.CreateTemplate(ctx, mashgate.CreateTemplateRequest{
    TenantID:     tenantID,
    TemplateKey:  "order-shipped",
    Channels:     []string{"email", "sms"},
    EmailSubject: "Your order shipped",
    EmailBody:    "<p>Order {{orderId}} is on the way.</p>",
    SmsText:      "Order {{orderId}} shipped",
    Vars:         []string{"orderId"},
})
templates, err := client.Notify.ListTemplates(ctx, tenantID)
```

**TypeScript**

```ts
const tmpl = await client.notify.createTemplate({
  tenantId,
  templateKey: "order-shipped",
  channels: ["email", "sms"],
  emailSubject: "Your order shipped",
  emailBodyHtml: "<p>Order {{orderId}} is on the way.</p>",
  smsText: "Order {{orderId}} shipped",
  vars: ["orderId"],
});
const templates = await client.notify.listTemplates(tenantId);
```

**Python**

```python
tmpl = client.notify.create_template(
    tenant_id=tenant_id,
    name="order-shipped",
    channel="email",
    body_template="<p>Order {{orderId}} is on the way.</p>",
    subject_template="Your order shipped",
)
templates = client.notify.list_templates(tenant_id)
```

### Delivery logs

**Go**

```go
logs, err := client.Notify.ListLogs(ctx, tenantID, 1) // page
```

**TypeScript**

```ts
const logs = await client.notify.listLogs({ tenantId, page: 1 });
```

**Python**

```python
logs = client.notify.list_logs(tenant_id=tenant_id, page=1)
```

## Key operations — mail

Auth differs by scope: pass a **user JWT** for self-service mailbox operations (`mail:read` / `mail:write`), or **admin/service-account** credentials for tenant-wide operations (`mail:admin`).

> **Python:** the mail resource is not yet available — use Go or TypeScript.

### Read your mailbox and messages

**Go**

```go
box, err := client.Mail.GetMyMailbox(ctx)

page, err := client.Mail.ListMessages(ctx, mashgate.ListMailMessagesQuery{
    Folder: mashgate.MessageFolderInbox,
    Limit:  25,
})
// page.NextCursor for the next page

full, err := client.Mail.GetMessage(ctx, messageID)
```

**TypeScript**

```ts
const box = await client.mail.getMyMailbox();

const page = await client.mail.listMessages({
  folder: MessageFolder.Inbox,
  limit: 25,
});

const full = await client.mail.getMessage(messageId);
```

### Send a message

Sends are **idempotent** via `idempotencyKey` — repeating the same key on the same tenant returns the original `message_id`. Delivery is asynchronous; the final outcome arrives as a HookLine event (see [Events](#events)).

**Go**

```go
res, err := client.Mail.SendMessage(ctx, mashgate.SendMailRequest{
    To:             []string{"user@example.com"},
    Subject:        "Welcome",
    BodyHTML:       "<p>Hello!</p>",
    AttachmentIDs:  []string{fileID}, // pre-uploaded via Storage
    IdempotencyKey: idem,
})
// res.Status: SEND_STATUS_QUEUED | SENT | DELIVERED | FAILED
```

**TypeScript**

```ts
const res = await client.mail.sendMessage({
  to: ["user@example.com"],
  subject: "Welcome",
  body_html: "<p>Hello!</p>",
  attachment_ids: [fileId], // pre-uploaded via Storage
  idempotency_key: idem,
});
```

### Update flags and delete

`updateMessage` patches only the fields you pass (`read`, `folder`, `labels`). `deleteMessage` moves to `TRASH`; passing `hardDelete=true` on a message already in `TRASH` removes it permanently.

**Go**

```go
read := true
_, err := client.Mail.UpdateMessage(ctx, messageID, mashgate.UpdateMailMessageRequest{
    Read:   &read,
    Folder: mashgate.MessageFolderInbox,
})
err = client.Mail.DeleteMessage(ctx, messageID, false)
```

**TypeScript**

```ts
await client.mail.updateMessage(messageId, { read: true, folder: MessageFolder.Inbox });
await client.mail.deleteMessage(messageId, false);
```

### Mailboxes and domains (admin, `mail:admin`)

Creating a domain returns it in `pending` with a generated `dkim_public_key`; publish the DKIM/SPF/DMARC TXT records in your DNS, then call `VerifyDomain`. Failed verification returns `verification_errors`.

**Go**

```go
boxes, err := client.Mail.ListMailboxes(ctx, mashgate.ListMailboxesQuery{
    Status: mashgate.MailboxStatusActive, Limit: 50,
})
box, err := client.Mail.CreateMailbox(ctx, mashgate.CreateMailboxRequest{
    SubjectID: userID, Email: "user@mail.example.com",
})

dom, err := client.Mail.CreateDomain(ctx, mashgate.CreateMailDomainRequest{Name: "mail.example.com"})
dom, err = client.Mail.VerifyDomain(ctx, dom.DomainID)
dom, err = client.Mail.RotateDKIM(ctx, dom.DomainID, 2048) // old selector valid ~30d
```

**TypeScript**

```ts
const boxes = await client.mail.listMailboxes({ status: MailboxStatus.Active, limit: 50 });
const box = await client.mail.createMailbox({ subject_id: userId, email: "user@mail.example.com" });

let dom = await client.mail.createDomain({ name: "mail.example.com" });
dom = await client.mail.verifyDomain(dom.domain_id);
dom = await client.mail.rotateDKIM(dom.domain_id, { key_bits: 2048 });
```

## Events

The **mail** module emits HookLine events you can subscribe to via webhooks:

- `mail.received` — an inbound message landed in a mailbox.
- `mail.sent` — a queued message was handed off for delivery.
- `mail.delivered` — the receiving server accepted the message.
- `mail.bounced` — delivery failed.

Because `mail.sendMessage` is asynchronous, treat the `SendMessageResponse.status` as the *initial* state and rely on these events for the final outcome. (The `notify` resource records delivery in its own logs rather than emitting these events.)

## Best practices

- Use `templateKey`/templates instead of inlining copy at every call site, so message content is centrally managed and localizable.
- Make sends idempotent — pass an `idempotencyKey`/`idempotency_key` so retries don't double-send. See [Best practices §2](../best-practices.md).
- Be event-driven for mail outcomes: subscribe to `mail.delivered`/`mail.bounced` and keep handlers idempotent. See [Best practices §3](../best-practices.md).
- Verify mail domains and rotate DKIM keys ahead of expiry; the previous selector stays valid ~30 days for DNS propagation.
- Derive the tenant/subject from the token, not request input. See [Best practices §4–5](../best-practices.md).

## See also

- [Storage](./storage.md) — pre-upload attachments before referencing them in a mail send
- [Chat](./chat.md) — in-app messaging (vs. email/SMS)
- [Flags & Observability](./flags-observability.md) — delivery metrics and logs
- [Service catalog](./service-catalog.md) — `notify-service`, `mail-service`
- [Best practices](../best-practices.md)
