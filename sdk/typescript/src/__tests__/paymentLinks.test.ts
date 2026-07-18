import { describe, expect, it, vi } from "vitest";
import { MashgateClient } from "../client.js";

describe("PaymentLinksResource", () => {
  it("uses merchantId as structural ownership and preserves description", async () => {
    const fetch = vi
      .fn()
      .mockResolvedValueOnce({
        ok: true,
        status: 201,
        statusText: "Created",
        headers: new Headers(),
        json: async () => ({ id: "link-1", merchantId: "merchant-1", description: "invoice 42" }),
      })
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        statusText: "OK",
        headers: new Headers(),
        json: async () => [],
      });
    const client = new MashgateClient({ baseUrl: "https://example.test", fetch });

    await client.paymentLinks.create({
      tenantId: "tenant-1",
      merchantId: "merchant-1",
      amount: 42,
      currency: "USDT",
      description: "invoice 42",
    });
    const [, createInit] = fetch.mock.calls[0];
    expect(JSON.parse(createInit.body)).toMatchObject({
      merchantId: "merchant-1",
      description: "invoice 42",
    });

    await client.paymentLinks.list("tenant-1", "merchant-1");
    expect(fetch.mock.calls[1][0]).toContain("merchantId=merchant-1");
  });
});
