"""IAM resource (mgID).

Mirrors ``iam.v1.IamService`` exposed over the gateway as REST.

Surface unions the Go SDK (``sdk/go/iam.go``) and the TypeScript SDK
(``sdk/typescript/src/resources/iam.ts``):

- Tenant identity + lifecycle (create / get / list / update / suspend /
  delete / provisioning-status / quotas / bulk actions).
- Roles, groups, policies (RBAC + ABAC).
- API keys (create / list / rotate / revoke).
- Access evaluation (``evaluate`` ABAC, ``check`` simple permission probe).
- Audit events and application scopes.

All methods return the parsed JSON ``dict`` from the gateway. Callers that
want a specific sub-field (e.g. ``["roles"]``) pull it from the returned
dict; we keep the full envelope so callers also see pagination / counts.
"""

from __future__ import annotations

from typing import TYPE_CHECKING, Any

if TYPE_CHECKING:
    from mashgate.client import MashgateClient


class IamResource:
    """Client for the full ``iam.v1.IamService`` (mgID)."""

    def __init__(self, client: MashgateClient) -> None:
        self._c = client

    # ── Permissions ───────────────────────────────────────────────────

    def list_permissions(self, tenant_id: str) -> dict[str, Any]:
        return self._c.request(
            "GET", "/v1/iam/permissions", query={"tenantId": tenant_id}
        )

    # ── Roles ─────────────────────────────────────────────────────────

    def list_roles(self, tenant_id: str) -> dict[str, Any]:
        return self._c.request("GET", "/v1/iam/roles", query={"tenantId": tenant_id})

    def upsert_role(
        self,
        *,
        tenant_id: str,
        code: str,
        name: str,
        permissions: list[str],
        role_id: str | None = None,
        metadata: dict[str, str] | None = None,
    ) -> dict[str, Any]:
        body: dict[str, Any] = {
            "tenantId": tenant_id,
            "code": code,
            "name": name,
            "permissions": permissions,
        }
        if role_id is not None:
            body["roleId"] = role_id
        if metadata is not None:
            body["metadata"] = metadata
        return self._c.request("POST", "/v1/iam/roles", body=body)

    def delete_role(self, tenant_id: str, role_id: str) -> dict[str, Any]:
        return self._c.request(
            "DELETE", f"/v1/iam/roles/{role_id}", query={"tenantId": tenant_id}
        )

    def assign_role(
        self, tenant_id: str, principal_id: str, role_id: str
    ) -> dict[str, Any]:
        return self._c.request(
            "POST",
            "/v1/iam/roles/assign",
            body={
                "tenantId": tenant_id,
                "principalId": principal_id,
                "roleId": role_id,
            },
        )

    def revoke_role(
        self, tenant_id: str, principal_id: str, role_id: str
    ) -> dict[str, Any]:
        return self._c.request(
            "POST",
            "/v1/iam/roles/revoke",
            body={
                "tenantId": tenant_id,
                "principalId": principal_id,
                "roleId": role_id,
            },
        )

    # ── Groups ────────────────────────────────────────────────────────

    def list_groups(self, tenant_id: str) -> dict[str, Any]:
        return self._c.request("GET", "/v1/iam/groups", query={"tenantId": tenant_id})

    def upsert_group(
        self,
        *,
        tenant_id: str,
        code: str,
        name: str,
        group_id: str | None = None,
        metadata: dict[str, str] | None = None,
    ) -> dict[str, Any]:
        body: dict[str, Any] = {"tenantId": tenant_id, "code": code, "name": name}
        if group_id is not None:
            body["groupId"] = group_id
        if metadata is not None:
            body["metadata"] = metadata
        return self._c.request("POST", "/v1/iam/groups", body=body)

    def delete_group(self, tenant_id: str, group_id: str) -> dict[str, Any]:
        return self._c.request(
            "DELETE", f"/v1/iam/groups/{group_id}", query={"tenantId": tenant_id}
        )

    def add_group_member(
        self, tenant_id: str, group_id: str, principal_id: str
    ) -> dict[str, Any]:
        return self._c.request(
            "POST",
            f"/v1/iam/groups/{group_id}/members",
            body={"tenantId": tenant_id, "principalId": principal_id},
        )

    def remove_group_member(
        self, tenant_id: str, group_id: str, principal_id: str
    ) -> dict[str, Any]:
        return self._c.request(
            "DELETE",
            f"/v1/iam/groups/{group_id}/members/{principal_id}",
            query={"tenantId": tenant_id},
        )

    # ── Policies ──────────────────────────────────────────────────────

    def upsert_policy(
        self,
        *,
        tenant_id: str,
        code: str,
        conditions_json: str,
        policy_id: str | None = None,
        description: str | None = None,
        metadata: dict[str, str] | None = None,
    ) -> dict[str, Any]:
        body: dict[str, Any] = {
            "tenantId": tenant_id,
            "code": code,
            "conditionsJson": conditions_json,
        }
        if policy_id is not None:
            body["policyId"] = policy_id
        if description is not None:
            body["description"] = description
        if metadata is not None:
            body["metadata"] = metadata
        return self._c.request("POST", "/v1/iam/policies", body=body)

    def delete_policy(self, tenant_id: str, policy_id: str) -> dict[str, Any]:
        return self._c.request(
            "DELETE", f"/v1/iam/policies/{policy_id}", query={"tenantId": tenant_id}
        )

    def bind_policy(
        self,
        *,
        tenant_id: str,
        policy_id: str,
        binding_type: str,
        binding_id: str,
    ) -> dict[str, Any]:
        return self._c.request(
            "POST",
            "/v1/iam/policies/bind",
            body={
                "tenantId": tenant_id,
                "policyId": policy_id,
                "bindingType": binding_type,
                "bindingId": binding_id,
            },
        )

    def unbind_policy(
        self,
        *,
        tenant_id: str,
        policy_id: str,
        binding_type: str,
        binding_id: str,
    ) -> dict[str, Any]:
        return self._c.request(
            "POST",
            "/v1/iam/policies/unbind",
            body={
                "tenantId": tenant_id,
                "policyId": policy_id,
                "bindingType": binding_type,
                "bindingId": binding_id,
            },
        )

    # ── API keys ──────────────────────────────────────────────────────

    def list_api_keys(self, tenant_id: str) -> dict[str, Any]:
        return self._c.request(
            "GET", "/v1/iam/api-keys", query={"tenantId": tenant_id}
        )

    def create_api_key(
        self,
        *,
        tenant_id: str,
        client_id: str,
        name: str,
        scopes: list[str] | None = None,
        rpm: int | None = None,
        burst: int | None = None,
        expires_at: int | None = None,
        app_id: str | None = None,
    ) -> dict[str, Any]:
        """Create a new API key.

        The returned ``plaintextKey`` is shown once — store it securely.
        """
        body: dict[str, Any] = {
            "tenantId": tenant_id,
            "clientId": client_id,
            "name": name,
        }
        if scopes is not None:
            body["scopes"] = scopes
        if rpm is not None:
            body["rpm"] = rpm
        if burst is not None:
            body["burst"] = burst
        if expires_at is not None:
            body["expiresAt"] = expires_at
        if app_id is not None:
            body["appId"] = app_id
        return self._c.request("POST", "/v1/iam/api-keys", body=body)

    def rotate_api_key(self, tenant_id: str, api_key_id: str) -> dict[str, Any]:
        """Rotate an API key. The returned ``plaintextKey`` is shown once."""
        return self._c.request(
            "POST",
            f"/v1/iam/api-keys/{api_key_id}/rotate",
            body={"tenantId": tenant_id},
        )

    def revoke_api_key(self, tenant_id: str, api_key_id: str) -> dict[str, Any]:
        return self._c.request(
            "DELETE",
            f"/v1/iam/api-keys/{api_key_id}",
            query={"tenantId": tenant_id},
        )

    # ── Access evaluation ─────────────────────────────────────────────

    def evaluate_access(
        self,
        *,
        tenant_id: str,
        principal_id: str,
        permission: str,
        resource_context: dict[str, str] | None = None,
    ) -> dict[str, Any]:
        """Full ABAC evaluation. Returns ``{allow, reason, matchedPolicies}``."""
        body: dict[str, Any] = {
            "tenantId": tenant_id,
            "principalId": principal_id,
            "permission": permission,
        }
        if resource_context is not None:
            body["resourceContext"] = resource_context
        return self._c.request("POST", "/v1/iam/evaluate", body=body)

    def check_permission(self, permission: str) -> dict[str, Any]:
        """Probe whether the authenticated principal has ``permission``.

        Mirrors the Go SDK's ``CheckPermission``; hits
        ``GET /v1/iam/check?permission=...`` and returns ``{"allowed": bool}``.
        Useful as lightweight middleware: ``if not res["allowed"]: 403``.
        """
        return self._c.request(
            "GET", "/v1/iam/check", query={"permission": permission}
        )

    # ── Audit ─────────────────────────────────────────────────────────

    def list_audit_events(
        self,
        *,
        tenant_id: str,
        page: int | None = None,
        page_size: int | None = None,
    ) -> dict[str, Any]:
        return self._c.request(
            "GET",
            "/v1/iam/audit-events",
            query={"tenantId": tenant_id, "page": page, "pageSize": page_size},
        )

    # ── App scopes ────────────────────────────────────────────────────

    def register_app_scope(
        self, *, client_id: str, scope_code: str, description: str | None = None
    ) -> dict[str, Any]:
        body: dict[str, Any] = {"clientId": client_id, "scopeCode": scope_code}
        if description is not None:
            body["description"] = description
        return self._c.request("POST", "/v1/iam/app-scopes", body=body)

    def list_app_scopes(self, client_id: str) -> dict[str, Any]:
        return self._c.request(
            "GET", "/v1/iam/app-scopes", query={"clientId": client_id}
        )

    def grant_scope_to_role(
        self, *, tenant_id: str, role_id: str, scope_code: str, client_id: str
    ) -> dict[str, Any]:
        return self._c.request(
            "POST",
            "/v1/iam/app-scopes/grant",
            body={
                "tenantId": tenant_id,
                "roleId": role_id,
                "scopeCode": scope_code,
                "clientId": client_id,
            },
        )

    def revoke_scope_from_role(
        self, tenant_id: str, role_id: str, scope_code: str
    ) -> dict[str, Any]:
        return self._c.request(
            "POST",
            "/v1/iam/app-scopes/revoke",
            body={
                "tenantId": tenant_id,
                "roleId": role_id,
                "scopeCode": scope_code,
            },
        )

    def get_effective_scopes(
        self, *, tenant_id: str, principal_id: str, client_id: str
    ) -> dict[str, Any]:
        return self._c.request(
            "GET",
            "/v1/iam/effective-scopes",
            query={
                "tenantId": tenant_id,
                "principalId": principal_id,
                "clientId": client_id,
            },
        )

    # ── Tenant lifecycle ──────────────────────────────────────────────

    def create_tenant(
        self,
        *,
        code: str,
        name: str,
        mode: str | None = None,
        metadata: dict[str, str] | None = None,
    ) -> dict[str, Any]:
        """Create a Mashgate tenant. Provisioning is async — poll
        :meth:`get_tenant_provisioning_status` for completion.

        :param code: Unique slug, e.g. ``"acme-corp"``. Stored as ``code``.
        :param mode: ``"sandbox"`` | ``"live"``. Defaults to ``"sandbox"``.
        :param metadata: Arbitrary key-value pairs to link back to your
            system, e.g. ``{"grid_workspace_id": "uuid"}``.
        """
        body: dict[str, Any] = {"code": code, "name": name}
        if mode is not None:
            body["mode"] = mode
        if metadata is not None:
            body["metadata"] = metadata
        return self._c.request("POST", "/v1/iam/tenants", body=body)

    def get_tenant_provisioning_status(self, tenant_id: str) -> dict[str, Any]:
        return self._c.request(
            "GET", f"/v1/iam/tenants/{tenant_id}/provisioning"
        )

    def delete_tenant(self, tenant_id: str) -> dict[str, Any]:
        return self._c.request("DELETE", f"/v1/iam/tenants/{tenant_id}")

    # ── Admin tenant management ───────────────────────────────────────

    def list_tenants(
        self,
        *,
        page: int | None = None,
        page_size: int | None = None,
        search: str | None = None,
        status: str | None = None,
        sort_by: str | None = None,
        sort_order: str | None = None,
    ) -> dict[str, Any]:
        """List tenants visible to the authenticated principal.

        Used for cold-start backfill in downstream verticals before
        subscribing to the Kafka ``tenant-events`` topic (ADR-0020 Phase B).
        """
        return self._c.request(
            "GET",
            "/v1/iam/tenants",
            query={
                "page": page,
                "pageSize": page_size,
                "search": search,
                "status": status,
                "sortBy": sort_by,
                "sortOrder": sort_order,
            },
        )

    def get_tenant(self, tenant_id: str) -> dict[str, Any]:
        return self._c.request("GET", f"/v1/iam/tenants/{tenant_id}")

    def update_tenant(
        self,
        tenant_id: str,
        *,
        name: str | None = None,
        mode: str | None = None,
        metadata: dict[str, str] | None = None,
    ) -> dict[str, Any]:
        body: dict[str, Any] = {}
        if name is not None:
            body["name"] = name
        if mode is not None:
            body["mode"] = mode
        if metadata is not None:
            body["metadata"] = metadata
        return self._c.request("PUT", f"/v1/iam/tenants/{tenant_id}", body=body)

    def suspend_tenant(
        self, tenant_id: str, *, reason: str | None = None
    ) -> dict[str, Any]:
        body: dict[str, Any] = {}
        if reason is not None:
            body["reason"] = reason
        return self._c.request(
            "POST", f"/v1/iam/tenants/{tenant_id}/suspend", body=body or None
        )

    def list_tenant_quotas(self, tenant_id: str) -> dict[str, Any]:
        return self._c.request("GET", f"/v1/iam/tenants/{tenant_id}/quotas")

    def update_tenant_quota(
        self, tenant_id: str, resource: str, limit: int
    ) -> dict[str, Any]:
        return self._c.request(
            "PUT",
            f"/v1/iam/tenants/{tenant_id}/quotas/{resource}",
            body={"limit": limit},
        )

    def bulk_tenant_action(
        self, *, tenant_ids: list[str], action: str
    ) -> dict[str, Any]:
        """Apply ``action`` to many tenants at once.

        :param action: ``"suspend"`` | ``"activate"`` | ``"delete"``.
        :returns: ``{"affectedCount": int, "failedIds": [...]}``.
        """
        return self._c.request(
            "POST",
            "/v1/iam/tenants/bulk",
            body={"tenantIds": tenant_ids, "action": action},
        )
