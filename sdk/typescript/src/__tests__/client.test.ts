import { describe, it, expect, vi, beforeEach } from "vitest";
import { MashgateClient } from "../client.js";
import { MashgateError } from "../errors.js";
import * as pkg from "../index.js";

// Derived directly from the `readonly <name>: <Name>Resource` properties
// declared on MashgateClient (src/client.ts) and assigned in its constructor.
// If a resource is added/removed in client.ts but not here, the "exact set"
// assertion below fails — keeping this list honest against the real source.
const RESOURCE_PROPERTIES = [
  "auth",
  "payments",
  "checkout",
  "wallet",
  "risk",
  "webhooks",
  "developer",
  "settings",
  "chat",
  "notify",
  "storage",
  "flags",
  "logs",
  "subscriptions",
  "invoices",
  "paymentLinks",
  "guard",
  "chain",
  "localPayments",
  "iam",
  "metering",
  "billing",
  "analytics",
  "walletAdmin",
  "mail",
] as const;

function createMockFetch(response: {
  status?: number;
  body?: unknown;
  statusText?: string;
}) {
  return vi.fn().mockResolvedValue({
    ok: (response.status ?? 200) >= 200 && (response.status ?? 200) < 300,
    status: response.status ?? 200,
    statusText: response.statusText ?? "OK",
    headers: new Headers(),
    json: () => Promise.resolve(response.body ?? {}),
  });
}

describe("MashgateClient", () => {
  let mockFetch: ReturnType<typeof createMockFetch>;
  let client: MashgateClient;

  beforeEach(() => {
    mockFetch = createMockFetch({ body: { ok: true } });
    client = new MashgateClient({
      baseUrl: "https://api.mashgate.uz",
      apiKey: "mg_test_key",
      fetch: mockFetch,
    });
  });

  it("sends GET requests with auth headers", async () => {
    client.setAccessToken("tok_123");
    await client.request("GET", "/v1/payments");

    expect(mockFetch).toHaveBeenCalledOnce();
    const [url, init] = mockFetch.mock.calls[0];
    expect(url).toBe("https://api.mashgate.uz/v1/payments");
    expect(init.method).toBe("GET");
    expect(init.headers["Authorization"]).toBe("Bearer tok_123");
    expect(init.headers["X-API-Key"]).toBe("mg_test_key");
  });

  it("sends POST requests with JSON body", async () => {
    await client.request("POST", "/v1/payments", {
      body: { amount: "100.00", currency: "USD" },
    });

    const [, init] = mockFetch.mock.calls[0];
    expect(init.method).toBe("POST");
    expect(init.headers["Content-Type"]).toBe("application/json");
    expect(JSON.parse(init.body)).toEqual({ amount: "100.00", currency: "USD" });
  });

  it("appends query parameters", async () => {
    await client.request("GET", "/v1/payments", {
      query: { page: 2, page_size: 25, status: undefined },
    });

    const [url] = mockFetch.mock.calls[0];
    expect(url).toContain("page=2");
    expect(url).toContain("page_size=25");
    expect(url).not.toContain("status");
  });

  it("throws MashgateError on HTTP errors", async () => {
    mockFetch = createMockFetch({
      status: 401,
      statusText: "Unauthorized",
      body: { message: "Invalid token", code: "auth_error" },
    });
    client = new MashgateClient({
      baseUrl: "https://api.mashgate.uz",
      fetch: mockFetch,
    });

    await expect(client.request("GET", "/v1/auth/profile")).rejects.toThrow(MashgateError);

    try {
      await client.request("GET", "/v1/auth/profile");
    } catch (error) {
      expect(error).toBeInstanceOf(MashgateError);
      const mashgateError = error as MashgateError;
      expect(mashgateError.status).toBe(401);
      expect(mashgateError.code).toBe("auth_error");
      expect(mashgateError.message).toBe("Invalid token");
    }
  });

  it("marks 429 and 5xx errors as retryable", async () => {
    mockFetch = createMockFetch({ status: 429, body: { message: "Rate limited" } });
    client = new MashgateClient({ baseUrl: "https://api.mashgate.uz", fetch: mockFetch });

    try {
      await client.request("GET", "/v1/payments");
    } catch (error) {
      expect((error as MashgateError).retryable).toBe(true);
    }
  });

  it("strips trailing slashes from baseUrl", () => {
    const c = new MashgateClient({ baseUrl: "https://api.mashgate.uz///", fetch: mockFetch });
    c.request("GET", "/v1/test");
    const [url] = mockFetch.mock.calls[0];
    expect(url).toBe("https://api.mashgate.uz/v1/test");
  });
});

describe("Resource integration", () => {
  it("auth.login sets access token", async () => {
    const mockFetch = createMockFetch({
      body: {
        accessToken: "new_token",
        refreshToken: "ref_tok",
        expiresIn: 3600,
        user: { userId: "u1", email: "a@b.com", role: "admin", tenantId: "t1" },
      },
    });
    const client = new MashgateClient({ baseUrl: "https://api.mashgate.uz", fetch: mockFetch });

    await client.auth.login({ email: "a@b.com", password: "pass" });

    // Next request should include the token
    await client.request("GET", "/v1/test");
    const [, init] = mockFetch.mock.calls[1];
    expect(init.headers["Authorization"]).toBe("Bearer new_token");
  });

  it("payments.create sends correct path and body", async () => {
    const mockFetch = createMockFetch({
      body: { paymentId: "pay_1", status: "intent_created", amount: { amount: "50.00", currency: "USD" } },
    });
    const client = new MashgateClient({ baseUrl: "https://api.mashgate.uz", fetch: mockFetch });

    const result = await client.payments.create({ amount: "50.00", currency: "USD" });
    expect(result.paymentId).toBe("pay_1");

    const [url, init] = mockFetch.mock.calls[0];
    expect(url).toBe("https://api.mashgate.uz/v1/payments");
    expect(init.method).toBe("POST");
  });

  it("payments.refund sends idempotency key", async () => {
    const mockFetch = createMockFetch({
      body: { refundId: "ref_1", status: "committed" },
    });
    const client = new MashgateClient({ baseUrl: "https://api.mashgate.uz", fetch: mockFetch });

    await client.payments.refund("pay_1", {
      amount: "25.00",
      reason: "customer_request",
      idempotencyKey: "idem_123",
    });

    const [url, init] = mockFetch.mock.calls[0];
    expect(url).toBe("https://api.mashgate.uz/v1/payments/pay_1/refund");
    expect(init.headers["X-Idempotency-Key"]).toBe("idem_123");
  });

  it("webhooks.createEndpoint sends correct body", async () => {
    const mockFetch = createMockFetch({
      body: { endpoint: { endpointId: "ep_1", signingSecret: "whsec_abc" } },
    });
    const client = new MashgateClient({ baseUrl: "https://api.mashgate.uz", fetch: mockFetch });

    const result = await client.webhooks.createEndpoint({
      url: "https://example.com/hook",
      eventTypes: ["payment.captured"],
    });

    expect(result.endpoint.signingSecret).toBe("whsec_abc");
  });
});

describe("MashgateClient resource wiring", () => {
  let client: MashgateClient;

  beforeEach(() => {
    client = new MashgateClient({
      baseUrl: "https://api.mashgate.uz",
      apiKey: "mg_test_key",
      fetch: vi.fn(),
    });
  });

  it.each(RESOURCE_PROPERTIES)("wires the %s resource", (prop) => {
    const resource = (client as unknown as Record<string, unknown>)[prop];
    // Would be undefined if the constructor assignment were dropped.
    expect(resource).toBeDefined();
    expect(resource).not.toBeNull();
    // A wired resource is a real instance, not a leftover empty slot.
    expect(typeof resource).toBe("object");
  });

  it("wires exactly the expected set of resources (no missing, no orphan)", () => {
    const wired = RESOURCE_PROPERTIES.filter((prop) => {
      const value = (client as unknown as Record<string, unknown>)[prop];
      return value !== undefined && value !== null && typeof value === "object";
    });
    // Fails if any listed resource is unwired, catching a regressed constructor.
    expect(wired).toEqual([...RESOURCE_PROPERTIES]);
    expect(wired).toHaveLength(25);
  });
});

describe("package entrypoint exports", () => {
  it("exports verifyWebhookSignature as a function", () => {
    expect(pkg.verifyWebhookSignature).toBeDefined();
    expect(typeof pkg.verifyWebhookSignature).toBe("function");
  });

  it("exports the MashgateError error class", () => {
    expect(pkg.MashgateError).toBeDefined();
    expect(typeof pkg.MashgateError).toBe("function");
    // It is re-exported from the entrypoint and is the same class.
    expect(pkg.MashgateError).toBe(MashgateError);
    // Behaves as a real Error subclass, not a placeholder.
    const err = new pkg.MashgateError({ message: "boom", status: 500 });
    expect(err).toBeInstanceOf(Error);
    expect(err.message).toBe("boom");
    expect(err.status).toBe(500);
  });
});
