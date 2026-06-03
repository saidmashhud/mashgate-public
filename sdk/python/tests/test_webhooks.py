"""Tests for the HookLine webhook signature verifier.

The verifier is pure stdlib (HMAC-SHA256, timing-safe compare, millisecond
replay window) — no HTTP, so these are plain ``unittest`` cases with a local
helper that signs ``f"{timestamp}.{body}"`` exactly the way the server does.
A signature is ``v1=<hex>`` and the replay window is +/- ``max_age_ms``.
"""

from __future__ import annotations

import hashlib
import hmac
import time
import unittest

from mashgate import verify_webhook_signature


SECRET = "whsec_test_secret"
BODY = '{"event":"payment.succeeded","id":"evt_123"}'


def _now_ms() -> int:
    return int(time.time() * 1000)


def _sign(body: str, secret: str, ts: int | str) -> str:
    """Produce a valid ``v1=<hex>`` signature for ``{ts}.{body}``."""
    mac = hmac.new(
        secret.encode(), f"{ts}.{body}".encode(), hashlib.sha256
    ).hexdigest()
    return "v1=" + mac


class VerifyWebhookSignatureTest(unittest.TestCase):
    def test_valid_signature_within_window_returns_true(self):
        ts = _now_ms()
        sig = _sign(BODY, SECRET, ts)
        self.assertTrue(verify_webhook_signature(BODY, sig, SECRET, timestamp=ts))

    def test_tampered_body_returns_false(self):
        ts = _now_ms()
        sig = _sign(BODY, SECRET, ts)
        tampered = BODY.replace("succeeded", "failed")
        self.assertNotEqual(tampered, BODY)
        self.assertFalse(
            verify_webhook_signature(tampered, sig, SECRET, timestamp=ts)
        )

    def test_wrong_secret_returns_false(self):
        ts = _now_ms()
        sig = _sign(BODY, SECRET, ts)
        self.assertFalse(
            verify_webhook_signature(BODY, sig, "whsec_wrong", timestamp=ts)
        )

    def test_missing_v1_prefix_returns_false(self):
        ts = _now_ms()
        # Correct hex digest, but without the required "v1=" scheme prefix.
        raw_hex = _sign(BODY, SECRET, ts)[len("v1=") :]
        self.assertFalse(
            verify_webhook_signature(BODY, raw_hex, SECRET, timestamp=ts)
        )

    def test_empty_signature_returns_false(self):
        ts = _now_ms()
        self.assertFalse(verify_webhook_signature(BODY, "", SECRET, timestamp=ts))

    def test_none_signature_returns_false(self):
        ts = _now_ms()
        self.assertFalse(
            verify_webhook_signature(BODY, None, SECRET, timestamp=ts)
        )

    def test_timestamp_none_returns_false(self):
        # A perfectly valid signature is still rejected without a timestamp,
        # because the server signs over "{timestamp}.{body}".
        sig = _sign(BODY, SECRET, _now_ms())
        self.assertFalse(
            verify_webhook_signature(BODY, sig, SECRET, timestamp=None)
        )

    def test_non_numeric_timestamp_returns_false(self):
        ts = "not-a-number"
        sig = _sign(BODY, SECRET, ts)
        self.assertFalse(
            verify_webhook_signature(BODY, sig, SECRET, timestamp=ts)
        )

    def test_timestamp_older_than_window_returns_false(self):
        # Correctly signed, but 10 minutes old — outside the 5-minute window.
        ts = _now_ms() - 600_000
        sig = _sign(BODY, SECRET, ts)
        self.assertFalse(
            verify_webhook_signature(BODY, sig, SECRET, timestamp=ts)
        )

    def test_timestamp_in_future_beyond_window_returns_false(self):
        # Correctly signed, but 10 minutes in the future — outside the window.
        ts = _now_ms() + 600_000
        sig = _sign(BODY, SECRET, ts)
        self.assertFalse(
            verify_webhook_signature(BODY, sig, SECRET, timestamp=ts)
        )

    def test_max_age_zero_disables_window(self):
        # An ancient but correctly-signed timestamp passes when the window is off.
        ts = _now_ms() - 86_400_000  # one day old
        sig = _sign(BODY, SECRET, ts)
        self.assertTrue(
            verify_webhook_signature(BODY, sig, SECRET, timestamp=ts, max_age_ms=0)
        )
        # ...and the same call with the default window would reject it.
        self.assertFalse(
            verify_webhook_signature(BODY, sig, SECRET, timestamp=ts)
        )

    def test_str_and_bytes_payload_both_verify(self):
        ts = _now_ms()
        sig = _sign(BODY, SECRET, ts)
        self.assertTrue(
            verify_webhook_signature(BODY, sig, SECRET, timestamp=ts)
        )
        self.assertTrue(
            verify_webhook_signature(BODY.encode(), sig, SECRET, timestamp=ts)
        )

    def test_flipped_single_hex_char_returns_false(self):
        ts = _now_ms()
        sig = _sign(BODY, SECRET, ts)
        # Flip exactly one hex digit of an otherwise-valid signature.
        last = sig[-1]
        flipped = "0" if last != "0" else "1"
        bad = sig[:-1] + flipped
        self.assertNotEqual(bad, sig)
        self.assertFalse(
            verify_webhook_signature(BODY, bad, SECRET, timestamp=ts)
        )


if __name__ == "__main__":
    unittest.main()
