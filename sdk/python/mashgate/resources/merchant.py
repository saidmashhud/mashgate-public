"""Tenant-scoped merchant acceptance lifecycle."""

from __future__ import annotations

from enum import Enum
from typing import TYPE_CHECKING, Any

if TYPE_CHECKING:
    from mashgate.client import MashgateClient


class MerchantStatus(str, Enum):
    PENDING = "MERCHANT_STATUS_PENDING"
    KYC_REQUIRED = "MERCHANT_STATUS_KYC_REQUIRED"
    UNDER_REVIEW = "MERCHANT_STATUS_UNDER_REVIEW"
    ACCEPTED = "MERCHANT_STATUS_ACCEPTED"
    REJECTED = "MERCHANT_STATUS_REJECTED"
    SUSPENDED = "MERCHANT_STATUS_SUSPENDED"
    OFFBOARDED = "MERCHANT_STATUS_OFFBOARDED"


class MerchantType(str, Enum):
    INDIVIDUAL = "MERCHANT_TYPE_INDIVIDUAL"
    BUSINESS = "MERCHANT_TYPE_BUSINESS"
    DAO = "MERCHANT_TYPE_DAO"


class MerchantResource:
    def __init__(self, client: MashgateClient) -> None:
        self._c = client

    def onboard(
        self,
        *,
        subject_id: str,
        merchant_type: MerchantType | str,
        display_name: str,
        legal_name: str,
        registration_number: str,
        country_code: str,
        config: dict[str, Any],
        idempotency_key: str,
    ) -> dict[str, Any]:
        return self._c.request(
            "POST",
            "/v1/merchants",
            body={
                "tenantId": self._c.require_tenant_id(),
                "subjectId": subject_id,
                "merchantType": merchant_type,
                "displayName": display_name,
                "legalName": legal_name,
                "registrationNumber": registration_number,
                "countryCode": country_code,
                "config": config,
            },
            extra_headers={"Idempotency-Key": idempotency_key},
        )

    def get(self, merchant_id: str) -> dict[str, Any]:
        return self._c.request(
            "GET",
            f"/v1/merchants/{merchant_id}",
            query={"tenant_id": self._c.require_tenant_id()},
        )

    def list(
        self,
        *,
        status: MerchantStatus | str | None = None,
        limit: int | None = None,
        cursor: str | None = None,
    ) -> dict[str, Any]:
        return self._c.request(
            "GET",
            "/v1/merchants",
            query={
                "tenant_id": self._c.require_tenant_id(),
                "status": status,
                "limit": limit,
                "cursor": cursor,
            },
        )

    def accept(self, merchant_id: str, note: str) -> dict[str, Any]:
        return self._transition(merchant_id, "accept", {"note": note})

    def reject(self, merchant_id: str, rejection_reason: str) -> dict[str, Any]:
        return self._transition(
            merchant_id, "reject", {"rejection_reason": rejection_reason}
        )

    def suspend(self, merchant_id: str, suspend_reason: str) -> dict[str, Any]:
        return self._transition(
            merchant_id, "suspend", {"suspend_reason": suspend_reason}
        )

    def reinstate(self, merchant_id: str, note: str) -> dict[str, Any]:
        return self._transition(merchant_id, "reinstate", {"note": note})

    def _transition(
        self, merchant_id: str, action: str, fields: dict[str, str]
    ) -> dict[str, Any]:
        return self._c.request(
            "POST",
            f"/v1/merchants/{merchant_id}/{action}",
            body={
                "tenant_id": self._c.require_tenant_id(),
                "merchant_id": merchant_id,
                **fields,
            },
        )
