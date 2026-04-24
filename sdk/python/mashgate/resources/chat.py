"""Chat resource (mgChat)."""

from __future__ import annotations

from typing import TYPE_CHECKING, Any

if TYPE_CHECKING:
    from mashgate.client import MashgateClient


class ChatResource:
    def __init__(self, client: MashgateClient) -> None:
        self._c = client

    def create_channel(
        self,
        *,
        tenant_id: str,
        name: str,
        members: list[str] | None = None,
        metadata: dict[str, Any] | None = None,
    ) -> dict[str, Any]:
        body: dict[str, Any] = {"tenantId": tenant_id, "name": name}
        if members is not None:
            body["members"] = members
        if metadata is not None:
            body["metadata"] = metadata
        return self._c.request("POST", "/v1/chat/channels", body=body)

    def list_channels(self, tenant_id: str) -> dict[str, Any]:
        return self._c.request("GET", "/v1/chat/channels", query={"tenantId": tenant_id})

    def send_message(
        self,
        channel_id: str,
        *,
        tenant_id: str,
        sender_id: str,
        text: str,
        metadata: dict[str, Any] | None = None,
    ) -> dict[str, Any]:
        body: dict[str, Any] = {"tenantId": tenant_id, "senderId": sender_id, "text": text}
        if metadata is not None:
            body["metadata"] = metadata
        return self._c.request("POST", f"/v1/chat/channels/{channel_id}/messages", body=body)

    def list_messages(
        self,
        channel_id: str,
        *,
        tenant_id: str,
        before: str | None = None,
        limit: int | None = None,
    ) -> dict[str, Any]:
        return self._c.request(
            "GET",
            f"/v1/chat/channels/{channel_id}/messages",
            query={"tenantId": tenant_id, "before": before, "limit": limit},
        )

    def get_members(self, channel_id: str, *, tenant_id: str) -> dict[str, Any]:
        return self._c.request(
            "GET",
            f"/v1/chat/channels/{channel_id}/members",
            query={"tenantId": tenant_id},
        )

    def delete_message(
        self, channel_id: str, message_id: str, *, tenant_id: str
    ) -> None:
        self._c.request(
            "DELETE",
            f"/v1/chat/channels/{channel_id}/messages/{message_id}",
            query={"tenantId": tenant_id},
        )
