"""Storage resource (mgStorage)."""

from __future__ import annotations

from typing import TYPE_CHECKING, Any

if TYPE_CHECKING:
    from mashgate.client import MashgateClient


class StorageResource:
    def __init__(self, client: MashgateClient) -> None:
        self._c = client

    def generate_upload_url(
        self,
        *,
        tenant_id: str,
        filename: str,
        content_type: str | None = None,
        max_size_bytes: int | None = None,
    ) -> dict[str, Any]:
        body: dict[str, Any] = {"tenantId": tenant_id, "filename": filename}
        if content_type is not None:
            body["contentType"] = content_type
        if max_size_bytes is not None:
            body["maxSizeBytes"] = max_size_bytes
        return self._c.request("POST", "/v1/storage/upload-url", body=body)

    def list_files(self, tenant_id: str) -> dict[str, Any]:
        return self._c.request("GET", "/v1/storage/files", query={"tenantId": tenant_id})

    def delete_file(self, file_id: str, *, tenant_id: str) -> None:
        self._c.request(
            "DELETE", f"/v1/storage/files/{file_id}", query={"tenantId": tenant_id}
        )

    def get_download_url(self, file_id: str, *, tenant_id: str) -> dict[str, Any]:
        return self._c.request(
            "GET", f"/v1/storage/files/{file_id}/download", query={"tenantId": tenant_id}
        )
