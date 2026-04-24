"""Webhook signature verification utility."""

from __future__ import annotations

import hashlib
import hmac


def verify_webhook_signature(
    payload: str | bytes,
    signature: str,
    secret: str,
) -> bool:
    """Verify an incoming webhook signature (HMAC-SHA256).

    Args:
        payload: The raw request body.
        signature: The ``X-Mashgate-Signature`` header value.
        secret: The webhook endpoint's signing secret.

    Returns:
        ``True`` if the signature is valid.
    """
    if isinstance(payload, str):
        payload = payload.encode()
    expected = hmac.new(secret.encode(), payload, hashlib.sha256).hexdigest()
    return hmac.compare_digest(expected, signature)
