#!/usr/bin/env python3
"""Mashgate Python SDK — quickstart.

Runnable end-to-end example covering the three things every integration needs:

  1. Initialize the client from environment variables.
  2. Create a hosted checkout session (and a single payment) and print the
     redirect URL + ids you hand back to your frontend.
  3. Receive webhooks and verify their signature with the SDK helper.

Run the checkout/payment demo:

    export MASHGATE_BASE_URL="https://api.mashgate.uz"   # or http://localhost:9661 for local dev
    export MASHGATE_API_KEY="mg_test_..."
    python quickstart.py

Run the webhook receiver instead (Flask if installed, else stdlib):

    export MASHGATE_WEBHOOK_SECRET="whsec_..."
    python quickstart.py serve

Install deps with:  pip install -r requirements.txt
"""

from __future__ import annotations

import os
import sys
import uuid

from mashgate import MashgateClient, MashgateError, verify_webhook_signature
from mashgate.events import WebhookTopic, event_id, event_key, event_payload


# ──────────────────────────────────────────────────────────────────────────
# 1. Client init from env
# ──────────────────────────────────────────────────────────────────────────
# Keys are environment-scoped: mg_test_... for sandbox, mg_live_... for prod.
# Keep the key out of source — read it from the environment.
#
# MASHGATE_BASE_URL is the canonical name; MASHGATE_API_URL is accepted as an
# alias so this example works against either convention.

def make_client() -> MashgateClient:
    base_url = os.environ.get("MASHGATE_BASE_URL") or os.environ.get("MASHGATE_API_URL")
    api_key = os.environ.get("MASHGATE_API_KEY")
    if not base_url or not api_key:
        sys.exit(
            "Set MASHGATE_BASE_URL (or MASHGATE_API_URL) and MASHGATE_API_KEY first.\n"
            '  export MASHGATE_BASE_URL="https://api.mashgate.uz"\n'
            '  export MASHGATE_API_KEY="mg_test_..."'
        )
    # The constructor is keyword-only. It also accepts access_token (for
    # end-user auth), timeout (seconds, default 30.0), and headers.
    return MashgateClient(base_url=base_url, api_key=api_key)


# ──────────────────────────────────────────────────────────────────────────
# 2a. Hosted checkout session
# ──────────────────────────────────────────────────────────────────────────
# create_session returns parsed JSON as a dict. Redirect the customer to
# session["checkoutUrl"]; persist session["sessionId"] to reconcile later.

def create_checkout_session(mg: MashgateClient) -> dict:
    session = mg.checkout.create_session(
        currency="UZS",
        success_url="https://example.com/success?session={sessionId}",
        cancel_url="https://example.com/cancel",
        line_items=[
            {
                "name": "Pro plan",
                "quantity": 1,
                "unitPrice": {"amount": "150000.00", "currency": "UZS"},
            },
        ],
        expires_in_minutes=30,
        metadata={"order_id": "order-001"},
    )
    print("checkout session created")
    print("  sessionId   :", session.get("sessionId"))
    print("  status      :", session.get("status"))
    print("  totalAmount :", session.get("totalAmount"))
    print("  expiresAt   :", session.get("expiresAt"))
    print("  -> redirect customer to:", session.get("checkoutUrl"))
    return session


# ──────────────────────────────────────────────────────────────────────────
# 2b. Single payment
# ──────────────────────────────────────────────────────────────────────────
# Amounts are decimal strings. Pass a unique idempotency_key so retries never
# double-charge — reuse the same key when retrying the same logical request.

def create_payment(mg: MashgateClient) -> dict:
    payment = mg.payments.create(
        amount="150000.00",
        currency="UZS",
        order_id="order-001",
        idempotency_key=str(uuid.uuid4()),
    )
    print("payment created")
    print("  paymentId :", payment.get("paymentId"))
    print("  status    :", payment.get("status"))
    return payment


def run_demo() -> None:
    # The client is a context manager: it closes its connection pool on exit.
    with make_client() as mg:
        try:
            create_checkout_session(mg)
            create_payment(mg)
        except MashgateError as e:
            # Non-2xx responses raise MashgateError carrying status + code.
            print(f"Mashgate API error: status={e.status} code={e.code} message={e}")
            sys.exit(1)


# ──────────────────────────────────────────────────────────────────────────
# 3. Webhook receiver — verify the signature, then act on the event
# ──────────────────────────────────────────────────────────────────────────
# Mashgate signs every webhook with HMAC-SHA256 over the RAW request body and
# sends the hex digest in the X-Mashgate-Signature header. Always verify
# against the unparsed bytes — re-serialized JSON will not match.
#
# NOTE: the real SDK helper signature is
#   verify_webhook_signature(payload, signature, secret) -> bool
# The first positional arg is the raw payload (str or bytes).

SIGNATURE_HEADER = "X-Mashgate-Signature"


def handle_event(event: dict) -> None:
    """Dispatch a verified event. Use the envelope-agnostic helpers so this
    works for both envelope-v1 and legacy emissions."""
    topic = event_key(event)        # "payments.checkout_session.completed", etc.
    payload = event_payload(event)  # business payload (envelope `payload` or legacy `data`)
    print(f"verified event {event_id(event)} topic={topic}")

    if topic == WebhookTopic.PAYMENT_COMPLETED:
        print("  payment completed:", payload.get("paymentId"))
    elif topic == WebhookTopic.CHECKOUT_SESSION_COMPLETED:
        print("  checkout completed:", payload.get("sessionId"))
    else:
        print("  (ignored) topic:", topic)


def serve_flask(secret: str, port: int) -> bool:
    """Minimal Flask receiver. Returns False if Flask isn't installed."""
    try:
        from flask import Flask, Response, request
    except ImportError:
        return False

    app = Flask(__name__)

    @app.post("/webhooks/mashgate")
    def webhook():  # noqa: ANN202
        raw = request.get_data()  # raw bytes — do not use request.json here
        sig = request.headers.get(SIGNATURE_HEADER, "")
        # verify_webhook_signature(payload, signature, secret)
        if not verify_webhook_signature(raw, sig, secret):
            return Response("invalid signature", status=401)
        handle_event(request.get_json(force=True))
        return Response(status=204)  # ack fast; do heavy work async

    print(f"Flask webhook receiver on http://0.0.0.0:{port}/webhooks/mashgate")
    app.run(host="0.0.0.0", port=port)
    return True


def serve_stdlib(secret: str, port: int) -> None:
    """Zero-dependency fallback receiver using http.server."""
    import json
    from http.server import BaseHTTPRequestHandler, HTTPServer

    class Handler(BaseHTTPRequestHandler):
        def do_POST(self) -> None:  # noqa: N802
            if self.path != "/webhooks/mashgate":
                self.send_response(404)
                self.end_headers()
                return
            length = int(self.headers.get("Content-Length", 0))
            raw = self.rfile.read(length)  # raw bytes for signature verification
            sig = self.headers.get(SIGNATURE_HEADER, "")
            # verify_webhook_signature(payload, signature, secret)
            if not verify_webhook_signature(raw, sig, secret):
                self.send_response(401)
                self.end_headers()
                self.wfile.write(b"invalid signature")
                return
            handle_event(json.loads(raw))
            self.send_response(204)  # ack fast; do heavy work async
            self.end_headers()

        def log_message(self, *_args) -> None:  # quieter logs
            pass

    print(f"stdlib webhook receiver on http://0.0.0.0:{port}/webhooks/mashgate")
    HTTPServer(("0.0.0.0", port), Handler).serve_forever()


def run_server() -> None:
    secret = os.environ.get("MASHGATE_WEBHOOK_SECRET")
    if not secret:
        sys.exit('Set MASHGATE_WEBHOOK_SECRET first, e.g. export MASHGATE_WEBHOOK_SECRET="whsec_..."')
    port = int(os.environ.get("PORT", "8080"))
    if not serve_flask(secret, port):  # prefer Flask, fall back to stdlib
        serve_stdlib(secret, port)


if __name__ == "__main__":
    if len(sys.argv) > 1 and sys.argv[1] == "serve":
        run_server()
    else:
        run_demo()
