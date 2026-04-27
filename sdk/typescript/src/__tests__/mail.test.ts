import { describe, it, expect, vi, beforeEach } from "vitest";
import { MashgateClient } from "../client.js";
import { MessageFolder, MailboxStatus } from "../resources/mail.js";

interface MockCall {
  url: string;
  init: RequestInit & { headers: Record<string, string> };
}

function mockFetchReturning(body: unknown, status = 200) {
  return vi.fn().mockResolvedValue({
    ok: status >= 200 && status < 300,
    status,
    statusText: "OK",
    headers: new Headers(),
    json: () => Promise.resolve(body),
  });
}

function lastCall(mock: ReturnType<typeof mockFetchReturning>): MockCall {
  const [url, init] = mock.mock.calls[mock.mock.calls.length - 1];
  return { url: String(url), init: init as MockCall["init"] };
}

describe("MailResource", () => {
  let mockFetch: ReturnType<typeof mockFetchReturning>;
  let client: MashgateClient;

  beforeEach(() => {
    mockFetch = mockFetchReturning({});
    client = new MashgateClient({
      baseUrl: "https://api.mashgate.uz",
      apiKey: "mg_test_key",
      fetch: mockFetch,
    });
  });

  it("getMyMailbox GETs /v1/mail/mailboxes/me", async () => {
    mockFetch = mockFetchReturning({
      mailbox_id: "mb-1",
      tenant_id: "tnt-1",
      subject_id: "u-1",
      email: "demo@mail.entry-i.com",
      status: "MAILBOX_STATUS_ACTIVE",
      quota_bytes: 5368709120,
      used_bytes: 0,
      created_at: "2026-04-27T00:00:00Z",
      updated_at: "2026-04-27T00:00:00Z",
    });
    client = new MashgateClient({
      baseUrl: "https://api.mashgate.uz",
      apiKey: "mg_test_key",
      fetch: mockFetch,
    });

    const mb = await client.mail.getMyMailbox();
    expect(mb.email).toBe("demo@mail.entry-i.com");
    expect(mb.status).toBe(MailboxStatus.Active);

    const { url, init } = lastCall(mockFetch);
    expect(url).toBe("https://api.mashgate.uz/v1/mail/mailboxes/me");
    expect(init.method).toBe("GET");
  });

  it("listMessages passes folder and pagination as query", async () => {
    await client.mail.listMessages({
      folder: MessageFolder.Sent,
      limit: 25,
      cursor: "abc",
    });
    const { url } = lastCall(mockFetch);
    expect(url).toContain("/v1/mail/messages");
    expect(url).toContain("folder=MESSAGE_FOLDER_SENT");
    expect(url).toContain("limit=25");
    expect(url).toContain("cursor=abc");
  });

  it("sendMessage POSTs body to /v1/mail/messages with idempotency_key", async () => {
    mockFetch = mockFetchReturning({
      message_id: "msg-1",
      status: "SEND_STATUS_QUEUED",
      queued_at: "2026-04-27T10:00:00Z",
    });
    client = new MashgateClient({
      baseUrl: "https://api.mashgate.uz",
      apiKey: "mg_test_key",
      fetch: mockFetch,
    });

    const out = await client.mail.sendMessage({
      to: ["alice@example.com"],
      subject: "hi",
      body_text: "test",
      idempotency_key: "send-1",
    });
    expect(out.message_id).toBe("msg-1");
    expect(out.status).toBe("SEND_STATUS_QUEUED");

    const { url, init } = lastCall(mockFetch);
    expect(url).toBe("https://api.mashgate.uz/v1/mail/messages");
    expect(init.method).toBe("POST");
    const body = JSON.parse(init.body as string);
    expect(body.to).toEqual(["alice@example.com"]);
    expect(body.subject).toBe("hi");
    expect(body.idempotency_key).toBe("send-1");
  });

  it("updateMessage PATCHes only provided fields", async () => {
    await client.mail.updateMessage("msg-1", {
      read: true,
      folder: MessageFolder.Trash,
    });
    const { url, init } = lastCall(mockFetch);
    expect(url).toBe("https://api.mashgate.uz/v1/mail/messages/msg-1");
    expect(init.method).toBe("PATCH");
    const body = JSON.parse(init.body as string);
    expect(body.read).toBe(true);
    expect(body.folder).toBe(MessageFolder.Trash);
    expect(body.labels).toBeUndefined();
  });

  it("deleteMessage soft-deletes by default", async () => {
    await client.mail.deleteMessage("msg-1");
    const { url, init } = lastCall(mockFetch);
    expect(url).toContain("/v1/mail/messages/msg-1");
    expect(url).toContain("hard_delete=false");
    expect(init.method).toBe("DELETE");
  });

  it("createDomain POSTs name and returns Domain in pending", async () => {
    mockFetch = mockFetchReturning({
      domain_id: "d-1",
      tenant_id: "tnt-1",
      name: "mail.entry-i.com",
      status: "DOMAIN_STATUS_PENDING",
      mx_records: [],
      created_at: "2026-04-27T00:00:00Z",
    });
    client = new MashgateClient({
      baseUrl: "https://api.mashgate.uz",
      apiKey: "mg_test_key",
      fetch: mockFetch,
    });

    const d = await client.mail.createDomain({ name: "mail.entry-i.com" });
    expect(d.status).toBe("DOMAIN_STATUS_PENDING");

    const { url, init } = lastCall(mockFetch);
    expect(url).toBe("https://api.mashgate.uz/v1/mail/domains");
    expect(init.method).toBe("POST");
    const body = JSON.parse(init.body as string);
    expect(body.name).toBe("mail.entry-i.com");
  });

  it("rotateDKIM defaults to 2048 bits", async () => {
    await client.mail.rotateDKIM("d-1");
    const { url, init } = lastCall(mockFetch);
    expect(url).toBe("https://api.mashgate.uz/v1/mail/domains/d-1/dkim/rotate");
    expect(init.method).toBe("POST");
    const body = JSON.parse(init.body as string);
    expect(body.key_bits).toBe(2048);
  });
});
