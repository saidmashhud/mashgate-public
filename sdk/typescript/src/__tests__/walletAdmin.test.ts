import { describe, it, expect, vi, beforeEach } from "vitest";
import { MashgateClient } from "../client.js";
import {
  Currency,
  Network,
  Mint,
  WalletStatus,
} from "../resources/walletAdmin.js";

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

describe("WalletAdminResource", () => {
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

  it("createChain POSTs /v1/wallets/chain with body and returns mnemonic", async () => {
    mockFetch = mockFetchReturning({
      wallet: { wallet_id: "w-1", currency: "USDC" },
      mnemonic: "abandon ability ...",
    });
    client = new MashgateClient({
      baseUrl: "https://api.mashgate.uz",
      apiKey: "mg_test_key",
      fetch: mockFetch,
    });
    const out = await client.walletAdmin.createChain({
      subject_id: "u-1",
      subject_type: "user",
      currency: Currency.USDC,
      network: Network.Solana,
    });
    expect(out.mnemonic).toBe("abandon ability ...");
    expect(out.wallet.wallet_id).toBe("w-1");

    const { url, init } = lastCall(mockFetch);
    expect(url).toBe("https://api.mashgate.uz/v1/wallets/chain");
    expect(init.method).toBe("POST");
    const body = JSON.parse(init.body as string);
    expect(body.currency).toBe("USDC"); // typed alias serialises as plain string
    expect(body.network).toBe("SOLANA");
  });

  it("depositAddress passes mint when provided", async () => {
    await client.walletAdmin.depositAddress(
      "w-1",
      Network.Solana,
      Mint.USDCSolanaMainnet,
    );
    const { url } = lastCall(mockFetch);
    expect(url).toContain("/v1/wallets/w-1/deposit-address");
    expect(url).toContain("network=SOLANA");
    expect(url).toContain("mint=EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v");
  });

  it("depositAddress omits mint when empty (native asset)", async () => {
    await client.walletAdmin.depositAddress("w-1", Network.Solana, "");
    const { url } = lastCall(mockFetch);
    expect(url).toContain("network=SOLANA");
    expect(url).not.toContain("mint=");
  });

  it("withdraw includes mint in body", async () => {
    await client.walletAdmin.withdraw("w-1", {
      amount: "10.50",
      destination_type: "crypto_address",
      destination_id: "DestSolanaAddr",
      network: Network.Solana,
      mint: Mint.USDCSolanaMainnet,
    });
    const { init } = lastCall(mockFetch);
    const body = JSON.parse(init.body as string);
    expect(body.mint).toBe("EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v");
    expect(body.network).toBe("SOLANA");
    // Critically — description must NOT carry the legacy mint=...; hack.
    expect(body.description).toBeUndefined();
  });

  it("freeze and unfreeze hit the right endpoints", async () => {
    mockFetch = mockFetchReturning({
      wallet_id: "w-1",
      status: WalletStatus.Frozen,
    });
    client = new MashgateClient({
      baseUrl: "https://api.mashgate.uz",
      apiKey: "mg_test_key",
      fetch: mockFetch,
    });
    const w = await client.walletAdmin.freeze("w-1", "fraud-investigation");
    expect(w.status).toBe(WalletStatus.Frozen);
    let { url, init } = lastCall(mockFetch);
    expect(url).toContain("/v1/wallets/w-1/freeze");
    expect(init.method).toBe("POST");
    const fbody = JSON.parse(init.body as string);
    expect(fbody.freeze_reason).toBe("fraud-investigation");

    mockFetch = mockFetchReturning({
      wallet_id: "w-1",
      status: WalletStatus.Active,
    });
    client = new MashgateClient({
      baseUrl: "https://api.mashgate.uz",
      apiKey: "mg_test_key",
      fetch: mockFetch,
    });
    const u = await client.walletAdmin.unfreeze("w-1", "case-resolved");
    expect(u.status).toBe(WalletStatus.Active);
    ({ url } = lastCall(mockFetch));
    expect(url).toContain("/v1/wallets/w-1/unfreeze");
  });

  it("getTransaction GETs /v1/wallets/{id}/transactions/{tx}", async () => {
    mockFetch = mockFetchReturning({ transaction_id: "tx-99" });
    client = new MashgateClient({
      baseUrl: "https://api.mashgate.uz",
      apiKey: "mg_test_key",
      fetch: mockFetch,
    });
    const tx = await client.walletAdmin.getTransaction("w-1", "tx-99");
    expect(tx.transaction_id).toBe("tx-99");
    const { url, init } = lastCall(mockFetch);
    expect(url).toBe("https://api.mashgate.uz/v1/wallets/w-1/transactions/tx-99");
    expect(init.method).toBe("GET");
  });

  it("list passes cursor and limit as query params", async () => {
    mockFetch = mockFetchReturning({
      wallets: [{ wallet_id: "w-1" }],
      next_cursor: "opaque-token",
    });
    client = new MashgateClient({
      baseUrl: "https://api.mashgate.uz",
      apiKey: "mg_test_key",
      fetch: mockFetch,
    });
    const resp = await client.walletAdmin.list({
      subject_id: "user-1",
      limit: 25,
      cursor: "prev-token",
    });
    expect(resp.next_cursor).toBe("opaque-token");
    const { url } = lastCall(mockFetch);
    expect(url).toContain("subject_id=user-1");
    expect(url).toContain("limit=25");
    expect(url).toContain("cursor=prev-token");
  });

  it("typed constants have the expected wire values", () => {
    // Server-side parsers expect plain uppercase enum strings.
    expect(Currency.USDC).toBe("USDC");
    expect(Currency.UZS).toBe("UZS");
    expect(Network.Solana).toBe("SOLANA");
    expect(Network.Ethereum).toBe("ETHEREUM");
    expect(Mint.USDCSolanaMainnet).toBe(
      "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
    );
    expect(WalletStatus.Frozen).toBe("WALLET_STATUS_FROZEN");
  });
});
