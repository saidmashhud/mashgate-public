"""Tenant-scoped compliance alert lifecycle."""

from __future__ import annotations

from enum import Enum
from typing import TYPE_CHECKING, Any

if TYPE_CHECKING:
    from mashgate.client import MashgateClient


class AlertStatus(str, Enum):
    OPEN = "ALERT_STATUS_OPEN"
    UNDER_REVIEW = "ALERT_STATUS_UNDER_REVIEW"
    RESOLVED = "ALERT_STATUS_RESOLVED"
    ESCALATED = "ALERT_STATUS_ESCALATED"
    SAR_FILED = "ALERT_STATUS_SAR_FILED"
    CLOSED = "ALERT_STATUS_CLOSED"


class AlertSeverity(str, Enum):
    LOW = "ALERT_SEVERITY_LOW"
    MEDIUM = "ALERT_SEVERITY_MEDIUM"
    HIGH = "ALERT_SEVERITY_HIGH"
    CRITICAL = "ALERT_SEVERITY_CRITICAL"


class AlertCategory(str, Enum):
    AML = "ALERT_CATEGORY_AML"
    SANCTIONS = "ALERT_CATEGORY_SANCTIONS"
    PEP = "ALERT_CATEGORY_PEP"
    FRAUD = "ALERT_CATEGORY_FRAUD"
    TRANSACTION = "ALERT_CATEGORY_TRANSACTION"
    KYC = "ALERT_CATEGORY_KYC"
    WATCHLIST = "ALERT_CATEGORY_WATCHLIST"


class ComplianceResource:
    def __init__(self, client: MashgateClient) -> None:
        self._c = client

    def raise_alert(
        self,
        *,
        subject_id: str,
        subject_type: str,
        category: AlertCategory | str,
        severity: AlertSeverity | str,
        source: str,
        source_ref: str,
        description: str,
        idempotency_key: str,
        evidence: list[dict[str, Any]] | None = None,
    ) -> dict[str, Any]:
        body: dict[str, Any] = {
            "tenantId": self._c.require_tenant_id(),
            "subjectId": subject_id,
            "subjectType": subject_type,
            "category": category,
            "severity": severity,
            "source": source,
            "sourceRef": source_ref,
            "description": description,
        }
        if evidence is not None:
            body["evidence"] = evidence
        return self._c.request(
            "POST",
            "/v1/compliance/alerts",
            body=body,
            extra_headers={"Idempotency-Key": idempotency_key},
        )

    def get(self, alert_id: str) -> dict[str, Any]:
        return self._c.request(
            "GET",
            f"/v1/compliance/alerts/{alert_id}",
            query={"tenant_id": self._c.require_tenant_id()},
        )

    def list(
        self,
        *,
        subject_id: str | None = None,
        status: AlertStatus | str | None = None,
        severity: AlertSeverity | str | None = None,
        limit: int | None = None,
        cursor: str | None = None,
    ) -> dict[str, Any]:
        return self._c.request(
            "GET",
            "/v1/compliance/alerts",
            query={
                "tenant_id": self._c.require_tenant_id(),
                "subject_id": subject_id,
                "status": status,
                "severity": severity,
                "limit": limit,
                "cursor": cursor,
            },
        )

    def resolve(self, alert_id: str, resolve_note: str) -> dict[str, Any]:
        return self._c.request(
            "POST",
            f"/v1/compliance/alerts/{alert_id}/resolve",
            body={
                "tenant_id": self._c.require_tenant_id(),
                "alert_id": alert_id,
                "resolve_note": resolve_note,
            },
        )

    def escalate(
        self, alert_id: str, *, escalated_to: str, escalate_note: str
    ) -> dict[str, Any]:
        return self._c.request(
            "POST",
            f"/v1/compliance/alerts/{alert_id}/escalate",
            body={
                "tenant_id": self._c.require_tenant_id(),
                "alert_id": alert_id,
                "escalated_to": escalated_to,
                "escalate_note": escalate_note,
            },
        )

    def has_open_alerts(self, subject_id: str) -> bool:
        response = self.list(subject_id=subject_id, status=AlertStatus.OPEN, limit=1)
        return bool(response.get("alerts"))
