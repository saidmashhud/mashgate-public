"""Notify resource (mgNotify)."""

from __future__ import annotations

from typing import TYPE_CHECKING, Any

if TYPE_CHECKING:
    from mashgate.client import MashgateClient


class NotifyResource:
    def __init__(self, client: MashgateClient) -> None:
        self._c = client

    def send_sms(
        self,
        *,
        tenant_id: str,
        to: str,
        text: str,
        provider: str | None = None,
        template_id: str | None = None,
        template_vars: dict[str, str] | None = None,
    ) -> dict[str, Any]:
        body: dict[str, Any] = {"tenantId": tenant_id, "to": to, "text": text}
        if provider is not None:
            body["provider"] = provider
        if template_id is not None:
            body["templateId"] = template_id
        if template_vars is not None:
            body["templateVars"] = template_vars
        return self._c.request("POST", "/v1/notify/sms", body=body)

    def send_email(
        self,
        *,
        tenant_id: str,
        to: str,
        subject: str,
        body_html: str | None = None,
        body_text: str | None = None,
        template_id: str | None = None,
        template_vars: dict[str, str] | None = None,
    ) -> dict[str, Any]:
        body: dict[str, Any] = {"tenantId": tenant_id, "to": to, "subject": subject}
        if body_html is not None:
            body["bodyHtml"] = body_html
        if body_text is not None:
            body["bodyText"] = body_text
        if template_id is not None:
            body["templateId"] = template_id
        if template_vars is not None:
            body["templateVars"] = template_vars
        return self._c.request("POST", "/v1/notify/email", body=body)

    def create_template(
        self,
        *,
        tenant_id: str,
        name: str,
        channel: str,
        body_template: str,
        subject_template: str | None = None,
    ) -> dict[str, Any]:
        payload: dict[str, Any] = {
            "tenantId": tenant_id,
            "name": name,
            "channel": channel,
            "bodyTemplate": body_template,
        }
        if subject_template is not None:
            payload["subjectTemplate"] = subject_template
        return self._c.request("POST", "/v1/notify/templates", body=payload)

    def list_templates(self, tenant_id: str) -> dict[str, Any]:
        return self._c.request("GET", "/v1/notify/templates", query={"tenantId": tenant_id})

    def list_logs(
        self,
        *,
        tenant_id: str,
        page: int | None = None,
    ) -> dict[str, Any]:
        return self._c.request(
            "GET", "/v1/notify/logs", query={"tenantId": tenant_id, "page": page}
        )
