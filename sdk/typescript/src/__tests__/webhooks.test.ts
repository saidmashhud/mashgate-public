import { describe, it, expect } from "vitest";
import { createHmac } from "node:crypto";
import { verifyWebhookSignature } from "../webhooks.js";

const SECRET = "whsec_test_3f9a2c";
const TIMESTAMP = "1717000000";
const BODY = JSON.stringify({ topic: "payment.captured", id: "evt_123" });

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
  it("returns true for a correctly signed payload", async () => {
    const header = sign(SECRET, TIMESTAMP, BODY);
    await expect(verifyWebhookSignature(BODY, header, SECRET, TIMESTAMP)).resolves.toBe(true);
  });

  it("returns false when the body is tampered after signing", async () => {
    const header = sign(SECRET, TIMESTAMP, BODY);
    const tampered = BODY.replace("evt_123", "evt_999");
    expect(tampered).not.toBe(BODY);
    await expect(
      verifyWebhookSignature(tampered, header, SECRET, TIMESTAMP),
    ).resolves.toBe(false);
  });

  it("returns false when verified with the wrong secret", async () => {
    const header = sign(SECRET, TIMESTAMP, BODY);
    await expect(
      verifyWebhookSignature(BODY, header, "whsec_wrong_secret", TIMESTAMP),
    ).resolves.toBe(false);
  });

  it("returns false when the signature is bound to a different timestamp", async () => {
    // Signature is valid for TIMESTAMP, but the verifier is told a different one.
    const header = sign(SECRET, TIMESTAMP, BODY);
    await expect(
      verifyWebhookSignature(BODY, header, SECRET, "1717000999"),
    ).resolves.toBe(false);
  });

  it("returns false when the signature lacks the 'v1=' prefix", async () => {
    const header = sign(SECRET, TIMESTAMP, BODY);
    const noPrefix = header.slice(3); // strip "v1="
    expect(noPrefix.startsWith("v1=")).toBe(false);
    await expect(
      verifyWebhookSignature(BODY, noPrefix, SECRET, TIMESTAMP),
    ).resolves.toBe(false);
  });

  it("returns false for a different scheme prefix (e.g. 'v0=')", async () => {
    const header = sign(SECRET, TIMESTAMP, BODY);
    const v0 = `v0=${header.slice(3)}`;
    await expect(
      verifyWebhookSignature(BODY, v0, SECRET, TIMESTAMP),
    ).resolves.toBe(false);
  });

  it("returns false when the timestamp is empty (falsy)", async () => {
    const header = sign(SECRET, "", BODY);
    await expect(verifyWebhookSignature(BODY, header, SECRET, "")).resolves.toBe(false);
  });

  it("returns false for a well-formed v1= header with a wrong hex digest", async () => {
    const valid = sign(SECRET, TIMESTAMP, BODY).slice(3);
    // Flip the first hex nibble so length matches but the digest differs.
    const flipped = (valid[0] === "0" ? "1" : "0") + valid.slice(1);
    expect(flipped).not.toBe(valid);
    await expect(
      verifyWebhookSignature(BODY, `v1=${flipped}`, SECRET, TIMESTAMP),
    ).resolves.toBe(false);
  });

  it("verifies a Uint8Array body identically to its string form", async () => {
    const bytes = new TextEncoder().encode(BODY);
    const header = sign(SECRET, TIMESTAMP, bytes);
    await expect(
      verifyWebhookSignature(bytes, header, SECRET, TIMESTAMP),
    ).resolves.toBe(true);
    // A signature produced over the string form must also accept the bytes,
    // proving the verifier treats both representations as the same payload.
    await expect(
      verifyWebhookSignature(bytes, sign(SECRET, TIMESTAMP, BODY), SECRET, TIMESTAMP),
    ).resolves.toBe(true);
  });

  it("returns false for a tampered Uint8Array body", async () => {
    const bytes = new TextEncoder().encode(BODY);
    const header = sign(SECRET, TIMESTAMP, bytes);
    const tampered = new TextEncoder().encode(BODY.replace("payment", "PAYMENT"));
    await expect(
      verifyWebhookSignature(tampered, header, SECRET, TIMESTAMP),
    ).resolves.toBe(false);
  });

  // NOTE: src/webhooks.ts does NOT enforce a replay window — it only rejects an
  // empty/falsy timestamp (see line 11). The "within 5 minutes" wording in the
  // JSDoc of resources/webhooks.ts is aspirational and not implemented. We assert
  // the ACTUAL behavior: an old timestamp whose signature matches still verifies.
  it("does not enforce a replay window: an old-but-correctly-signed timestamp still verifies", async () => {
    const oldTs = "1000000000"; // 2001-09-09, far outside any 5-minute window
    const header = sign(SECRET, oldTs, BODY);
    await expect(
      verifyWebhookSignature(BODY, header, SECRET, oldTs),
    ).resolves.toBe(true);
  });
});
