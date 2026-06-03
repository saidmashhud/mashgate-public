import { describe, it, expect } from "vitest";
import { createHmac } from "node:crypto";
import { verifyWebhookSignature } from "../webhooks.js";

const SECRET = "whsec_test_3f9a2c";
const BODY = JSON.stringify({ topic: "payment.captured", id: "evt_123" });

// Current Unix-ms timestamp — valid cases must sit inside the replay window.
function nowTs(): string {
  return String(Date.now());
}

/**
 * Construct a real `x-hl-signature` header value the way a producer would:
 * HMAC-SHA256 over `{timestamp}.{body}`, hex-encoded, prefixed with `v1=`.
 * Computed with node crypto independently of the verifier's WebCrypto path,
 * so a regression in the verifier (wrong signing input, wrong algorithm,
 * wrong encoding) makes the assertions fail rather than pass tautologically.
 */
function sign(secret: string, timestamp: string, body: string | Uint8Array): string {
  const hmac = createHmac("sha256", secret);
  hmac.update(`${timestamp}.`);
  hmac.update(typeof body === "string" ? Buffer.from(body, "utf-8") : Buffer.from(body));
  return `v1=${hmac.digest("hex")}`;
}

describe("verifyWebhookSignature", () => {
  it("returns true for a correctly signed payload within the window", async () => {
    const ts = nowTs();
    const header = sign(SECRET, ts, BODY);
    await expect(verifyWebhookSignature(BODY, header, SECRET, ts)).resolves.toBe(true);
  });

  it("returns false when the body is tampered after signing", async () => {
    const ts = nowTs();
    const header = sign(SECRET, ts, BODY);
    const tampered = BODY.replace("evt_123", "evt_999");
    expect(tampered).not.toBe(BODY);
    await expect(
      verifyWebhookSignature(tampered, header, SECRET, ts),
    ).resolves.toBe(false);
  });

  it("returns false when verified with the wrong secret", async () => {
    const ts = nowTs();
    const header = sign(SECRET, ts, BODY);
    await expect(
      verifyWebhookSignature(BODY, header, "whsec_wrong_secret", ts),
    ).resolves.toBe(false);
  });

  it("returns false when the signature is bound to a different timestamp", async () => {
    // Both timestamps sit inside the window, so the rejection is due to the
    // signature being bound to a different timestamp — not the replay check.
    const signedTs = nowTs();
    const header = sign(SECRET, signedTs, BODY);
    const otherTs = String(Number(signedTs) + 5_000);
    await expect(
      verifyWebhookSignature(BODY, header, SECRET, otherTs),
    ).resolves.toBe(false);
  });

  it("returns false when the signature lacks the 'v1=' prefix", async () => {
    const ts = nowTs();
    const header = sign(SECRET, ts, BODY);
    const noPrefix = header.slice(3); // strip "v1="
    expect(noPrefix.startsWith("v1=")).toBe(false);
    await expect(
      verifyWebhookSignature(BODY, noPrefix, SECRET, ts),
    ).resolves.toBe(false);
  });

  it("returns false for a different scheme prefix (e.g. 'v0=')", async () => {
    const ts = nowTs();
    const header = sign(SECRET, ts, BODY);
    const v0 = `v0=${header.slice(3)}`;
    await expect(
      verifyWebhookSignature(BODY, v0, SECRET, ts),
    ).resolves.toBe(false);
  });

  it("returns false when the timestamp is empty (falsy)", async () => {
    const header = sign(SECRET, "", BODY);
    await expect(verifyWebhookSignature(BODY, header, SECRET, "")).resolves.toBe(false);
  });

  it("returns false for a non-numeric timestamp", async () => {
    const ts = "not-a-number";
    const header = sign(SECRET, ts, BODY);
    await expect(verifyWebhookSignature(BODY, header, SECRET, ts)).resolves.toBe(false);
  });

  it("returns false for a well-formed v1= header with a wrong hex digest", async () => {
    const ts = nowTs();
    const valid = sign(SECRET, ts, BODY).slice(3);
    // Flip the first hex nibble so length matches but the digest differs.
    const flipped = (valid[0] === "0" ? "1" : "0") + valid.slice(1);
    expect(flipped).not.toBe(valid);
    await expect(
      verifyWebhookSignature(BODY, `v1=${flipped}`, SECRET, ts),
    ).resolves.toBe(false);
  });

  it("verifies a Uint8Array body identically to its string form", async () => {
    const ts = nowTs();
    const bytes = new TextEncoder().encode(BODY);
    const header = sign(SECRET, ts, bytes);
    await expect(
      verifyWebhookSignature(bytes, header, SECRET, ts),
    ).resolves.toBe(true);
    // A signature produced over the string form must also accept the bytes,
    // proving the verifier treats both representations as the same payload.
    await expect(
      verifyWebhookSignature(bytes, sign(SECRET, ts, BODY), SECRET, ts),
    ).resolves.toBe(true);
  });

  it("returns false for a tampered Uint8Array body", async () => {
    const ts = nowTs();
    const bytes = new TextEncoder().encode(BODY);
    const header = sign(SECRET, ts, bytes);
    const tampered = new TextEncoder().encode(BODY.replace("payment", "PAYMENT"));
    await expect(
      verifyWebhookSignature(tampered, header, SECRET, ts),
    ).resolves.toBe(false);
  });

  // ── Replay window (aligned with the Go/Python verifiers) ──────────────────

  it("rejects an old-but-correctly-signed timestamp (replay protection)", async () => {
    const oldTs = String(Date.now() - 600_000); // 10 minutes old, outside ±5 min
    const header = sign(SECRET, oldTs, BODY);
    await expect(
      verifyWebhookSignature(BODY, header, SECRET, oldTs),
    ).resolves.toBe(false);
  });

  it("rejects a future timestamp beyond the window", async () => {
    const futureTs = String(Date.now() + 600_000); // 10 minutes ahead
    const header = sign(SECRET, futureTs, BODY);
    await expect(
      verifyWebhookSignature(BODY, header, SECRET, futureTs),
    ).resolves.toBe(false);
  });

  it("accepts an old timestamp when maxAgeMs=0 disables the window", async () => {
    const oldTs = String(Date.now() - 86_400_000); // one day old
    const header = sign(SECRET, oldTs, BODY);
    // Disabled window → accepted (signature is still correct).
    await expect(
      verifyWebhookSignature(BODY, header, SECRET, oldTs, 0),
    ).resolves.toBe(true);
    // Default window → rejected.
    await expect(
      verifyWebhookSignature(BODY, header, SECRET, oldTs),
    ).resolves.toBe(false);
  });
});
