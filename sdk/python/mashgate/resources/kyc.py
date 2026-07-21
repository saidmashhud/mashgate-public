"""Tenant-scoped KYC capability."""

from __future__ import annotations

from enum import Enum
from typing import TYPE_CHECKING, Any

if TYPE_CHECKING:
    from mashgate.client import MashgateClient


class KycStatus(str, Enum):
    UNSPECIFIED = "KYC_STATUS_UNSPECIFIED"
    PENDING = "KYC_STATUS_PENDING"
    IN_REVIEW = "KYC_STATUS_IN_REVIEW"
    PASSED = "KYC_STATUS_PASSED"
    FAILED = "KYC_STATUS_FAILED"
    EXPIRED = "KYC_STATUS_EXPIRED"
    OVERRIDDEN = "KYC_STATUS_OVERRIDDEN"


class KycSubjectType(str, Enum):
    INDIVIDUAL = "KYC_SUBJECT_INDIVIDUAL"
    BUSINESS = "KYC_SUBJECT_BUSINESS"


class KycCheckType(str, Enum):
    IDENTITY = "KYC_CHECK_IDENTITY"
    AML = "KYC_CHECK_AML"
    SANCTIONS = "KYC_CHECK_SANCTIONS"
    PEP = "KYC_CHECK_PEP"
    FULL = "KYC_CHECK_FULL"


class KYCResource:
    def __init__(self, client: MashgateClient) -> None:
        self._c = client

    def request_check(
        self,
        *,
        subject_id: str,
        subject_type: KycSubjectType | str,
        check_type: KycCheckType | str,
        idempotency_key: str,
        provider: str | None = None,
        metadata: dict[str, str] | None = None,
    ) -> dict[str, Any]:
        body: dict[str, Any] = {
            "tenantId": self._c.require_tenant_id(),
            "subjectId": subject_id,
            "subjectType": subject_type,
            "checkType": check_type,
        }
        if provider is not None:
            body["provider"] = provider
        if metadata is not None:
            body["metadata"] = metadata
        return self._c.request(
            "POST",
            "/v1/kyc/checks",
            body=body,
            extra_headers={"Idempotency-Key": idempotency_key},
        )

    def get(self, check_id: str) -> dict[str, Any]:
        return self._c.request(
            "GET",
            f"/v1/kyc/checks/{check_id}",
            query={"tenant_id": self._c.require_tenant_id()},
        )

    def list(
        self,
        *,
        subject_id: str | None = None,
        status: KycStatus | str | None = None,
        limit: int | None = None,
        cursor: str | None = None,
    ) -> dict[str, Any]:
        return self._c.request(
            "GET",
            "/v1/kyc/checks",
            query={
                "tenant_id": self._c.require_tenant_id(),
                "subject_id": subject_id,
                "status": status,
                "limit": limit,
                "cursor": cursor,
            },
        )

    def override(
        self,
        check_id: str,
        *,
        status: KycStatus | str,
        override_note: str,
    ) -> dict[str, Any]:
        return self._c.request(
            "POST",
            f"/v1/kyc/checks/{check_id}/override",
            body={
                "tenantId": self._c.require_tenant_id(),
                "checkId": check_id,
                "status": status,
                "overrideNote": override_note,
            },
        )
