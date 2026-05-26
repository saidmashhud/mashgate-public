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

  it("importChain POSTs /v1/wallets/chain/import and surfaces was_existing", async () => {
    mockFetch = mockFetchReturning({
      wallet: { wallet_id: "w-1", currency: "USDC", status: "WALLET_STATUS_ACTIVE" },
      was_existing: false,
      recovered_at: "2026-05-19T10:15:00Z",
    });
    client = new MashgateClient({
      baseUrl: "https://api.mashgate.uz",
      apiKey: "mg_test_key",
      fetch: mockFetch,
    });
    const out = await client.walletAdmin.importChain({
      subject_id: "user-1",
      subject_type: "user",
      currency: Currency.USDC,
      network: Network.Solana,
      mnemonic:
        "abandon ability able about above absent absorb abstract absurd abuse access accident",
      idempotency_key: "idem-import-1",
    });
    expect(out.wallet.wallet_id).toBe("w-1");
    expect(out.was_existing).toBe(false);
    expect(out.recovered_at).toBe("2026-05-19T10:15:00Z");

    const { url, init } = lastCall(mockFetch);
    expect(url).toBe("https://api.mashgate.uz/v1/wallets/chain/import");
    expect(init.method).toBe("POST");
    const body = JSON.parse(String(init.body));
    expect(body.mnemonic).toBe(
      "abandon ability able about above absent absorb abstract absurd abuse access accident",
    );
    expect(body.subject_id).toBe("user-1");
  });

  it("importChain returns was_existing=true on recovery", async () => {
    mockFetch = mockFetchReturning({
      wallet: { wallet_id: "w-existing", currency: "USDC" },
      was_existing: true,
      recovered_at: "2026-04-01T12:00:00Z",
    });
    client = new MashgateClient({
      baseUrl: "https://api.mashgate.uz",
      apiKey: "mg_test_key",
      fetch: mockFetch,
    });
    const out = await client.walletAdmin.importChain({
      subject_id: "user-1",
      subject_type: "user",
      currency: Currency.USDC,
      network: Network.Solana,
      mnemonic:
        "abandon ability able about above absent absorb abstract absurd abuse access accident",
    });
    expect(out.was_existing).toBe(true);
    expect(out.wallet.wallet_id).toBe("w-existing");
  });

  it("transfer POSTs /v1/wallets/{from}/transfer and returns both legs", async () => {
    mockFetch = mockFetchReturning({
      transfer_id: "xfer-uuid",
      debit: {
        transaction_id: "tx-debit",
        wallet_id: "w-from",
        type: "TRANSACTION_TYPE_DEBIT",
        amount: "25.50",
        currency: "USDC",
        balance_after: "74.50",
      },
      credit: {
        transaction_id: "tx-credit",
        wallet_id: "w-to",
        type: "TRANSACTION_TYPE_CREDIT",
        amount: "25.50",
        currency: "USDC",
        balance_after: "125.50",
      },
    });
    client = new MashgateClient({
      baseUrl: "https://api.mashgate.uz",
      apiKey: "mg_test_key",
      fetch: mockFetch,
    });
    const resp = await client.walletAdmin.transfer("w-from", {
      to_wallet_id: "w-to",
      amount: "25.50",
      reason: "TRANSACTION_REASON_SETTLEMENT",
      description: "monthly close",
      idempotency_key: "idem-xfer-1",
      merchant_id: "m-1",
      note: "Q2 settlement",
    });
    expect(resp.transfer_id).toBe("xfer-uuid");
    expect(resp.debit.wallet_id).toBe("w-from");
    expect(resp.credit.wallet_id).toBe("w-to");

    const { url, init } = lastCall(mockFetch);
    expect(url).toBe("https://api.mashgate.uz/v1/wallets/w-from/transfer");
    expect(init.method).toBe("POST");
    const body = JSON.parse(String(init.body));
    expect(body.to_wallet_id).toBe("w-to");
    expect(body.amount).toBe("25.50");
    expect(body.note).toBe("Q2 settlement");
    expect(body.idempotency_key).toBe("idem-xfer-1");
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
