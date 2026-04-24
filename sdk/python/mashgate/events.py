"""Webhook event envelope types and helpers.

Envelope v1 (ADR-0013 §4, contracts/events/_envelope.v1.json) is the
canonical shape Mashgate POSTs to webhook endpoints. Legacy emitters write
a flat body; the helpers below accept either form transparently.
"""

from __future__ import annotations

from typing import Any, Dict, Optional, TypedDict


class TraceContext(TypedDict, total=False):
    """W3C Trace Context carried under the envelope's ``_trace`` field."""
    traceparent: str
    tracestate: str


class WebhookEvent(TypedDict, total=False):
    """Shape of the JSON body posted to a webhook endpoint.

    Envelope v1 canonical fields (``id``, ``topic``, ``created_at``,
    ``payload``, ``source``, ``_trace``) and legacy fields (``event_type``,
    ``event_version``, ``correlation_id``, ``aggregate_id``, ``data``) may
    both be present depending on the producer's migration status. Use
    :func:`event_payload` and :func:`event_key` to read the event
    format-agnostically.
    """
    # Envelope v1
    id: str
    topic: str
    created_at: int
    payload: Any
    source: str
    _trace: TraceContext

    # Present in both legacy and envelope v1 emissions
    event_id: str
    tenant_id: str
    occurred_at: int
    event_type: str
    event_version: int

    # Legacy (pre-envelope) — `data` is the flat payload alias of `payload`.
    correlation_id: str
    aggregate_id: str
    data: Any


def event_payload(event: Dict[str, Any]) -> Any:
    """Return the business payload regardless of envelope version.

    Prefers ``payload`` (envelope v1), falls back to ``data`` (legacy),
    then an empty dict.
    """
    if "payload" in event and event["payload"] is not None:
        return event["payload"]
    if "data" in event and event["data"] is not None:
        return event["data"]
    return {}


def event_key(event: Dict[str, Any]) -> str:
    """Return the routing key: envelope ``topic`` if present else legacy ``event_type``."""
    return event.get("topic") or event.get("event_type") or ""


def event_id(event: Dict[str, Any]) -> Optional[str]:
    """Return the canonical event id (``id`` preferred over ``event_id``)."""
    return event.get("id") or event.get("event_id")


# Envelope v1 topic constants (ADR-0013 §4 `<product>.<resource>.<verb>`).
# Prefer these over the legacy dotted event_type strings.
class WebhookTopic:
    PAYMENT_CREATED             = "payments.payment.created"
    PAYMENT_COMPLETED           = "payments.payment.completed"
    PAYMENT_FAILED              = "payments.payment.failed"
    PAYMENT_AUTHORIZED          = "payments.payment.authorized"
    PAYMENT_VOIDED              = "payments.payment.voided"
    REFUND_CREATED              = "payments.refund.created"
    REFUND_COMPLETED            = "payments.refund.completed"
    REFUND_FAILED               = "payments.refund.failed"
    CHECKOUT_SESSION_CREATED    = "payments.checkout_session.created"
    CHECKOUT_SESSION_COMPLETED  = "payments.checkout_session.completed"
    USER_REGISTERED             = "iam.user.registered"
    NOTIFICATION_SENT           = "notifications.notification.sent"


__all__ = [
    "TraceContext",
    "WebhookEvent",
    "WebhookTopic",
    "event_payload",
    "event_key",
    "event_id",
]
