/**
 * Verify a webhook signature using HMAC-SHA256 with timing-safe comparison.
 * Works in both Node.js and browser (WebCrypto API) environments.
 */
export async function verifyWebhookSignature(
  payload: string | Uint8Array,
  signatureHeader: string,
  secret: string,
  timestamp: string,
): Promise<boolean> {
  if (!signatureHeader?.startsWith("v1=") || !timestamp) {
    return false;
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
