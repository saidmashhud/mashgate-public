/**
 * Verify a HookLine webhook signature using HMAC-SHA256 with timing-safe
 * comparison. Works in both Node.js and browser (WebCrypto API) environments.
 *
 * HookLine signs HMAC-SHA256 over `{timestamp}.{body}` and sends the signature
 * as `x-hl-signature: v1=<hex>` plus an `x-hl-timestamp` (Unix epoch ms). To
 * mitigate replay, a signature whose timestamp is more than `maxAgeMs` from now
 * (in either direction) is rejected; pass `maxAgeMs = 0` to disable the window.
 * This matches the Go and Python SDK verifiers.
 *
 * @param payload   raw request body (string or bytes), NOT re-serialized JSON
 * @param signatureHeader `x-hl-signature` header value (`v1=<hex>`)
 * @param secret    the endpoint signing secret
 * @param timestamp `x-hl-timestamp` header value (Unix epoch milliseconds)
 * @param maxAgeMs  replay window in ms (default 300_000 = 5 min; 0 disables it)
 */
export async function verifyWebhookSignature(
  payload: string | Uint8Array,
  signatureHeader: string,
  secret: string,
  timestamp: string,
  maxAgeMs: number = 300_000,
): Promise<boolean> {
  if (!signatureHeader?.startsWith("v1=") || !timestamp) {
    return false;
  }

  // Replay-window check (fail fast, before the HMAC). The timestamp is Unix
  // epoch milliseconds; reject anything outside ±maxAgeMs of now.
  if (maxAgeMs > 0) {
    const tsMs = Number(timestamp);
    if (!Number.isFinite(tsMs) || Math.abs(Date.now() - tsMs) > maxAgeMs) {
      return false;
    }
  }

  const encoder = new TextEncoder();
  const payloadBytes = typeof payload === "string" ? encoder.encode(payload) : payload;
  const prefixBytes = encoder.encode(`${timestamp}.`);
  const signingInput = new Uint8Array(prefixBytes.length + payloadBytes.length);
  signingInput.set(prefixBytes, 0);
  signingInput.set(payloadBytes, prefixBytes.length);
  const secretBytes = encoder.encode(secret);

  const key = await crypto.subtle.importKey(
    "raw",
    secretBytes,
    { name: "HMAC", hash: "SHA-256" },
    false,
    ["sign"],
  );

  const signatureBuffer = await crypto.subtle.sign("HMAC", key, signingInput);
  const computed = bufferToHex(signatureBuffer);
  const provided = signatureHeader.slice(3);

  return timingSafeEqual(computed, provided);
}

function bufferToHex(buffer: ArrayBuffer): string {
  const bytes = new Uint8Array(buffer);
  let hex = "";
  for (let i = 0; i < bytes.length; i++) {
    hex += bytes[i].toString(16).padStart(2, "0");
  }
  return hex;
}

function timingSafeEqual(a: string, b: string): boolean {
  if (a.length !== b.length) return false;

  let result = 0;
  for (let i = 0; i < a.length; i++) {
    result |= a.charCodeAt(i) ^ b.charCodeAt(i);
  }
  return result === 0;
}
