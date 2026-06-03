"""Mail resource (mgMail) — Mashgate core primitive (ADR-0019).

Mirrors ``mail.v1.MailService`` from
``mashgate/contracts/proto/v1/mail.proto``, exposed over the gateway as
REST via google.api.http transcoding.

Auth: pass an end-user JWT for self-service mailbox operations
(``mail:read`` / ``mail:write`` scope), or admin/service-account
credentials for tenant-wide operations (``mail:admin`` scope).

Events (subscribe via webhooks): ``mail.received`` / ``mail.sent`` /
``mail.delivered`` / ``mail.bounced`` — see ``contracts/events/mail.*.json``.
"""

from __future__ import annotations

from enum import Enum
from typing import TYPE_CHECKING, Any

if TYPE_CHECKING:
    from mashgate.client import MashgateClient


# ── Typed string enums from mail.proto ──────────────────────────────────
#
# We inherit from ``str`` so members serialise transparently to JSON and so
# callers may pass either ``MessageFolder.INBOX`` or the plain string
# ``"MESSAGE_FOLDER_INBOX"`` — resource methods cast to ``str(...)`` at the
# wire boundary.


class MessageFolder(str, Enum):
    INBOX = "MESSAGE_FOLDER_INBOX"
    SENT = "MESSAGE_FOLDER_SENT"
    DRAFTS = "MESSAGE_FOLDER_DRAFTS"
    SPAM = "MESSAGE_FOLDER_SPAM"
    TRASH = "MESSAGE_FOLDER_TRASH"


class MailboxStatus(str, Enum):
    ACTIVE = "MAILBOX_STATUS_ACTIVE"
    FROZEN = "MAILBOX_STATUS_FROZEN"
    CLOSED = "MAILBOX_STATUS_CLOSED"


class DomainStatus(str, Enum):
    PENDING = "DOMAIN_STATUS_PENDING"
    ACTIVE = "DOMAIN_STATUS_ACTIVE"
    SUSPENDED = "DOMAIN_STATUS_SUSPENDED"


class SendStatus(str, Enum):
    QUEUED = "SEND_STATUS_QUEUED"
    SENT = "SEND_STATUS_SENT"
    DELIVERED = "SEND_STATUS_DELIVERED"
    FAILED = "SEND_STATUS_FAILED"


def _opt_str(v: Any) -> str | None:
    """Coerce enum / None / str to plain string. Returns None for None."""
    if v is None:
        return None
    return str(v.value) if isinstance(v, Enum) else str(v)


class MailResource:
    """Client for the full ``mail.v1.MailService``.

    User-facing methods (``mail:read`` / ``mail:write``) operate on the
    authenticated subject's mailbox; admin methods (``mail:admin``) operate
    tenant-wide.
    """

    def __init__(self, client: MashgateClient) -> None:
        self._c = client

    # ── User-facing (mail:read / mail:write) ──────────────────────────────

    def get_my_mailbox(self) -> dict[str, Any]:
        """Return the mailbox for the authenticated subject (from JWT)."""
        return self._c.request("GET", "/v1/mail/mailboxes/me")

    def list_messages(
        self,
        *,
        folder: MessageFolder | str | None = None,
        limit: int | None = None,
        cursor: str | None = None,
        mailbox_id_override: str | None = None,
    ) -> dict[str, Any]:
        """List message previews from a folder (cursor pagination).

        :param mailbox_id_override: ``mail:admin`` only — read another
            subject's mailbox. Omit to read your own.
        """
        return self._c.request(
            "GET",
            "/v1/mail/messages",
            query={
                "folder": _opt_str(folder),
                "limit": limit,
                "cursor": cursor,
                "mailbox_id_override": mailbox_id_override,
            },
        )

    def get_message(self, message_id: str) -> dict[str, Any]:
        """Return a single message with body, headers and attachment links."""
        return self._c.request("GET", f"/v1/mail/messages/{message_id}")

    def send_message(
        self,
        *,
        to: list[str],
        subject: str,
        cc: list[str] | None = None,
        bcc: list[str] | None = None,
        body_text: str | None = None,
        body_html: str | None = None,
        headers: dict[str, str] | None = None,
        attachment_ids: list[str] | None = None,
        in_reply_to: str | None = None,
        idempotency_key: str | None = None,
    ) -> dict[str, Any]:
        """Queue a message for delivery.

        Idempotent via ``idempotency_key`` — the same key on the same tenant
        returns the original ``message_id``. SMTP delivery is asynchronous;
        status arrives later via ``mail.sent`` / ``mail.delivered`` /
        ``mail.bounced`` events.

        :param attachment_ids: IDs of files pre-uploaded via Mashgate
            storage-service.
        :param in_reply_to: Message-ID of the parent message for threading.
        """
        body: dict[str, Any] = {"to": to, "subject": subject}
        if cc is not None:
            body["cc"] = cc
        if bcc is not None:
            body["bcc"] = bcc
        if body_text is not None:
            body["body_text"] = body_text
        if body_html is not None:
            body["body_html"] = body_html
        if headers is not None:
            body["headers"] = headers
        if attachment_ids is not None:
            body["attachment_ids"] = attachment_ids
        if in_reply_to is not None:
            body["in_reply_to"] = in_reply_to
        if idempotency_key is not None:
            body["idempotency_key"] = idempotency_key
        return self._c.request("POST", "/v1/mail/messages", body=body)

    def update_message(
        self,
        message_id: str,
        *,
        read: bool | None = None,
        folder: MessageFolder | str | None = None,
        labels: list[str] | None = None,
    ) -> dict[str, Any]:
        """Patch mutable flags (read/folder/labels) on a message.

        Pass only the fields you want to change.
        """
        body: dict[str, Any] = {}
        if read is not None:
            body["read"] = read
        if folder is not None:
            body["folder"] = _opt_str(folder)
        if labels is not None:
            body["labels"] = labels
        return self._c.request("PATCH", f"/v1/mail/messages/{message_id}", body=body)

    def delete_message(self, message_id: str, *, hard_delete: bool = False) -> Any:
        """Soft-delete a message — moves it to TRASH.

        If the message is already in TRASH and ``hard_delete=True``, it is
        permanently removed.
        """
        return self._c.request(
            "DELETE",
            f"/v1/mail/messages/{message_id}",
            query={"hard_delete": hard_delete},
        )

    # ── Admin-facing (mail:admin) ─────────────────────────────────────────

    def list_mailboxes(
        self,
        *,
        status: MailboxStatus | str | None = None,
        limit: int | None = None,
        cursor: str | None = None,
    ) -> dict[str, Any]:
        """List all mailboxes in the tenant (admin)."""
        return self._c.request(
            "GET",
            "/v1/mail/mailboxes",
            query={
                "status": _opt_str(status),
                "limit": limit,
                "cursor": cursor,
            },
        )

    def create_mailbox(
        self,
        *,
        subject_id: str,
        email: str,
        display_name: str | None = None,
        quota_bytes: int | None = None,
        idempotency_key: str | None = None,
    ) -> dict[str, Any]:
        """Create a mailbox (admin).

        ``email`` must be in one of the tenant's active domains.

        :param quota_bytes: ``0`` (or omitted) = the tenant default quota.
        """
        body: dict[str, Any] = {"subject_id": subject_id, "email": email}
        if display_name is not None:
            body["display_name"] = display_name
        if quota_bytes is not None:
            body["quota_bytes"] = quota_bytes
        if idempotency_key is not None:
            body["idempotency_key"] = idempotency_key
        return self._c.request("POST", "/v1/mail/mailboxes", body=body)

    def list_domains(
        self, *, status: DomainStatus | str | None = None
    ) -> dict[str, Any]:
        """List the tenant's mail domains."""
        return self._c.request(
            "GET", "/v1/mail/domains", query={"status": _opt_str(status)}
        )

    def create_domain(self, *, name: str) -> dict[str, Any]:
        """Register a domain (admin).

        Returns a ``Domain`` in ``pending`` status with a generated
        ``dkim_public_key`` — the caller must publish the DKIM/SPF/DMARC TXT
        records in DNS, then call :meth:`verify_domain`.

        :param name: FQDN, e.g. ``"mail.entry-i.com"``.
        """
        return self._c.request("POST", "/v1/mail/domains", body={"name": name})

    def verify_domain(self, domain_id: str) -> dict[str, Any]:
        """Re-check DNS records for the domain.

        On success the status flips to ``active`` and the domain may
        send/receive mail. On failure the returned ``Domain`` has
        ``verification_errors`` populated.
        """
        return self._c.request(
            "POST", f"/v1/mail/domains/{domain_id}/verify", body={}
        )

    def rotate_dkim(self, domain_id: str, *, key_bits: int = 2048) -> dict[str, Any]:
        """Generate a new DKIM key for the domain.

        The previous selector remains valid until the next rotation
        (~30 days) for graceful DNS propagation at recipients.

        :param key_bits: ``1024`` / ``2048``. Defaults to ``2048``.
        """
        return self._c.request(
            "POST",
            f"/v1/mail/domains/{domain_id}/dkim/rotate",
            body={"key_bits": key_bits},
        )
