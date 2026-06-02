"""Webhook signature verification utility."""

from __future__ import annotations

import hashlib
import hmac
import time

# HookLine replay window: 5 minutes, matching the server (300000 ms).
SIGNATURE_MAX_AGE_MS = 300_000


def verify_webhook_signature(
    payload: str | bytes,
    signature: str,
    secret: str,
    timestamp: str | int | None = None,
    max_age_ms: int = SIGNATURE_MAX_AGE_MS,
) -> bool:
    """Verify a HookLine webhook signature (HMAC-SHA256, timing-safe).

    HookLine signs ``HMAC_SHA256(secret, f"{timestamp}.{body}")`` and sends:

    * ``x-hl-signature``: ``v1=<hex>``
    * ``x-hl-timestamp``: Unix epoch **milliseconds**

    This mirrors the TypeScript/Go SDK verifiers exactly — the signed input is
    ``{timestamp}.{body}`` (NOT the raw body), and the header carries a ``v1=``
    prefix that must be stripped before comparison.

    Args:
        payload: The raw request body.
        signature: The ``x-hl-signature`` header value (``v1=<hex>``).
        secret: The webhook endpoint's signing secret.
        timestamp: The ``x-hl-timestamp`` header value (Unix milliseconds).
            Required — the server signs over ``{timestamp}.{body}``.
        max_age_ms: Reject signatures whose timestamp is older/newer than this
            (replay protection). Pass ``0`` to skip the window check.

    Returns:
        ``True`` iff the ``v1=`` signature matches and, when ``max_age_ms`` is
        set, the timestamp is within the replay window.
    """
    if not signature or not signature.startswith("v1=") or timestamp is None:
        return False

    ts = str(timestamp)

    # Replay window — timestamp is in milliseconds, matching the server.
    if max_age_ms:
        try:
            ts_ms = int(ts)
        except ValueError:
            return False
        now_ms = int(time.time() * 1000)
        if abs(now_ms - ts_ms) > max_age_ms:
            return False

    body = bytes(payload) if isinstance(payload, (bytes, bytearray)) else payload.encode()
    signing_input = ts.encode() + b"." + body

    expected = hmac.new(secret.encode(), signing_input, hashlib.sha256).hexdigest()
    provided = signature[3:]  # strip the "v1=" prefix
    return hmac.compare_digest(expected, provided)
