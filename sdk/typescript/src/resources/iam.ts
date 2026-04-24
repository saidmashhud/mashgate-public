import type { MashgateClient } from "../client.js";
import type {
  Permission,
  Role,
  Group,
  Policy,
  ApiKey,
  AuditEvent,
  Tenant,
  TenantQuota,
  AppScope,
} from "../types.js";

// ── Roles ──────────────────────────────────────────────────────────────────

export interface UpsertRoleRequest {
  tenantId: string;
  roleId?: string;
  code: string;
  name: string;
  permissions: string[];
  metadata?: Record<string, string>;
}

// ── Groups ─────────────────────────────────────────────────────────────────

export interface UpsertGroupRequest {
  tenantId: string;
  groupId?: string;
  code: string;
  name: string;
  metadata?: Record<string, string>;
}

// ── Policies ───────────────────────────────────────────────────────────────

export interface UpsertPolicyRequest {
  tenantId: string;
  policyId?: string;
  code: string;
  description?: string;
  conditionsJson: string;
  metadata?: Record<string, string>;
}

export interface BindPolicyRequest {
  tenantId: string;
  policyId: string;
  bindingType: string;
  bindingId: string;
}

// ── API Keys ───────────────────────────────────────────────────────────────

export interface CreateApiKeyRequest {
  tenantId: string;
  clientId: string;
  name: string;
  scopes?: string[];
  rpm?: number;
  burst?: number;
  expiresAt?: number;
  appId?: string;
}

export interface CreateApiKeyResponse {
  apiKey: ApiKey;
  plaintextKey: string;
}

export interface RotateApiKeyResponse {
  apiKey: ApiKey;
  plaintextKey: string;
}

// ── Access evaluation ──────────────────────────────────────────────────────

export interface EvaluateAccessRequest {
  tenantId: string;
  principalId: string;
  permission: string;
  resourceContext?: Record<string, string>;
}

export interface EvaluateAccessResponse {
  allow: boolean;
  reason: string;
  matchedPolicies: string[];
}

// ── Audit ──────────────────────────────────────────────────────────────────

export interface ListAuditEventsOptions {
  tenantId: string;
  page?: number;
  pageSize?: number;
}

export interface ListAuditEventsResponse {
  events: AuditEvent[];
  totalCount: number;
}

// ── App scopes ─────────────────────────────────────────────────────────────

export interface RegisterAppScopeRequest {
  clientId: string;
  scopeCode: string;
  description?: string;
}

export interface GrantScopeToRoleRequest {
  tenantId: string;
  roleId: string;
  scopeCode: string;
  clientId: string;
}

export interface GetEffectiveScopesRequest {
  tenantId: string;
  principalId: string;
  clientId: string;
}

// ── Tenants ────────────────────────────────────────────────────────────────

export interface CreateTenantRequest {
  /** Unique slug, e.g. "acme-corp". Stored as `code` in Mashgate. */
  code: string;
  name: string;
  /** "sandbox" | "live". Defaults to "sandbox". */
  mode?: string;
  /** Arbitrary key-value pairs. Use to link back to your system, e.g. { grid_workspace_id: "uuid" } */
  metadata?: Record<string, string>;
}

export interface CreateTenantResponse {
  tenantId: string;
  code: string;
  name: string;
  mode: string;
  /** "pending" | "active" | "failed" — provisioning is async */
  provStatus: string;
  createdAt: number;
}

export interface TenantProvisioningStatus {
  tenantId: string;
  /** "pending" | "active" | "failed" | "deleting" | "deleted" */
  provStatus: string;
  errorMsg?: string;
  attempt: number;
  updatedAt: number;
}

export interface ListTenantsOptions {
  page?: number;
  pageSize?: number;
  search?: string;
  status?: string;
  sortBy?: string;
  sortOrder?: string;
}

export interface ListTenantsResponse {
  tenants: Tenant[];
  totalCount: number;
}

export interface UpdateTenantRequest {
  name?: string;
  mode?: string;
  metadata?: Record<string, string>;
}

export interface SuspendTenantResponse {
  success: boolean;
  tenant: Tenant;
}

export interface BulkTenantActionRequest {
  tenantIds: string[];
  /** "suspend" | "activate" | "delete" */
  action: string;
}

export interface BulkTenantActionResponse {
  affectedCount: number;
  failedIds: string[];
}

// ── Resource class ─────────────────────────────────────────────────────────

export class IamResource {
  constructor(private readonly client: MashgateClient) {}

  // Permissions
  async listPermissions(tenantId: string): Promise<Permission[]> {
    const res = await this.client.request<{ permissions: Permission[] }>("GET", "/v1/iam/permissions", {
      query: { tenantId },
    });
    return res.permissions;
  }

  // Roles
  async listRoles(tenantId: string): Promise<Role[]> {
    const res = await this.client.request<{ roles: Role[] }>("GET", "/v1/iam/roles", {
      query: { tenantId },
    });
    return res.roles;
  }

  async upsertRole(data: UpsertRoleRequest): Promise<Role> {
    const res = await this.client.request<{ role: Role }>("POST", "/v1/iam/roles", { body: data });
    return res.role;
  }

  async deleteRole(tenantId: string, roleId: string): Promise<boolean> {
    const res = await this.client.request<{ success: boolean }>("DELETE", `/v1/iam/roles/${roleId}`, {
      query: { tenantId },
    });
    return res.success;
  }

  async assignRole(tenantId: string, principalId: string, roleId: string): Promise<boolean> {
    const res = await this.client.request<{ success: boolean }>("POST", "/v1/iam/roles/assign", {
      body: { tenantId, principalId, roleId },
    });
    return res.success;
  }

  async revokeRole(tenantId: string, principalId: string, roleId: string): Promise<boolean> {
    const res = await this.client.request<{ success: boolean }>("POST", "/v1/iam/roles/revoke", {
      body: { tenantId, principalId, roleId },
    });
    return res.success;
  }

  // Groups
  async listGroups(tenantId: string): Promise<Group[]> {
    const res = await this.client.request<{ groups: Group[] }>("GET", "/v1/iam/groups", {
      query: { tenantId },
    });
    return res.groups;
  }

  async upsertGroup(data: UpsertGroupRequest): Promise<Group> {
    const res = await this.client.request<{ group: Group }>("POST", "/v1/iam/groups", { body: data });
    return res.group;
  }

  async deleteGroup(tenantId: string, groupId: string): Promise<boolean> {
    const res = await this.client.request<{ success: boolean }>("DELETE", `/v1/iam/groups/${groupId}`, {
      query: { tenantId },
    });
    return res.success;
  }

  async addGroupMember(tenantId: string, groupId: string, principalId: string): Promise<boolean> {
    const res = await this.client.request<{ success: boolean }>("POST", `/v1/iam/groups/${groupId}/members`, {
      body: { tenantId, principalId },
    });
    return res.success;
  }

  async removeGroupMember(tenantId: string, groupId: string, principalId: string): Promise<boolean> {
    const res = await this.client.request<{ success: boolean }>(
      "DELETE",
      `/v1/iam/groups/${groupId}/members/${principalId}`,
      { query: { tenantId } },
    );
    return res.success;
  }

  // Policies
  async upsertPolicy(data: UpsertPolicyRequest): Promise<Policy> {
    const res = await this.client.request<{ policy: Policy }>("POST", "/v1/iam/policies", { body: data });
    return res.policy;
  }

  async deletePolicy(tenantId: string, policyId: string): Promise<boolean> {
    const res = await this.client.request<{ success: boolean }>("DELETE", `/v1/iam/policies/${policyId}`, {
      query: { tenantId },
    });
    return res.success;
  }

  async bindPolicy(data: BindPolicyRequest): Promise<boolean> {
    const res = await this.client.request<{ success: boolean }>("POST", "/v1/iam/policies/bind", { body: data });
    return res.success;
  }

  async unbindPolicy(data: BindPolicyRequest): Promise<boolean> {
    const res = await this.client.request<{ success: boolean }>("POST", "/v1/iam/policies/unbind", { body: data });
    return res.success;
  }

  // API Keys
  async listApiKeys(tenantId: string): Promise<ApiKey[]> {
    const res = await this.client.request<{ apiKeys: ApiKey[] }>("GET", "/v1/iam/api-keys", {
      query: { tenantId },
    });
    return res.apiKeys;
  }

  async createApiKey(data: CreateApiKeyRequest): Promise<CreateApiKeyResponse> {
    return this.client.request<CreateApiKeyResponse>("POST", "/v1/iam/api-keys", { body: data });
  }

  async rotateApiKey(tenantId: string, apiKeyId: string): Promise<RotateApiKeyResponse> {
    return this.client.request<RotateApiKeyResponse>("POST", `/v1/iam/api-keys/${apiKeyId}/rotate`, {
      body: { tenantId },
    });
  }

  async revokeApiKey(tenantId: string, apiKeyId: string): Promise<boolean> {
    const res = await this.client.request<{ success: boolean }>("DELETE", `/v1/iam/api-keys/${apiKeyId}`, {
      query: { tenantId },
    });
    return res.success;
  }

  // Access evaluation
  async evaluateAccess(data: EvaluateAccessRequest): Promise<EvaluateAccessResponse> {
    return this.client.request<EvaluateAccessResponse>("POST", "/v1/iam/evaluate", { body: data });
  }

  // Audit
  async listAuditEvents(options: ListAuditEventsOptions): Promise<ListAuditEventsResponse> {
    return this.client.request<ListAuditEventsResponse>("GET", "/v1/iam/audit-events", {
      query: {
        tenantId: options.tenantId,
        page: options.page,
        pageSize: options.pageSize,
      },
    });
  }

  // App scopes
  async registerAppScope(data: RegisterAppScopeRequest): Promise<AppScope> {
    const res = await this.client.request<{ scope: AppScope }>("POST", "/v1/iam/app-scopes", { body: data });
    return res.scope;
  }

  async listAppScopes(clientId: string): Promise<AppScope[]> {
    const res = await this.client.request<{ scopes: AppScope[] }>("GET", "/v1/iam/app-scopes", {
      query: { clientId },
    });
    return res.scopes;
  }

  async grantScopeToRole(data: GrantScopeToRoleRequest): Promise<boolean> {
    const res = await this.client.request<{ success: boolean }>("POST", "/v1/iam/app-scopes/grant", { body: data });
    return res.success;
  }

  async revokeScopeFromRole(tenantId: string, roleId: string, scopeCode: string): Promise<boolean> {
    const res = await this.client.request<{ success: boolean }>("POST", "/v1/iam/app-scopes/revoke", {
      body: { tenantId, roleId, scopeCode },
    });
    return res.success;
  }

  async getEffectiveScopes(data: GetEffectiveScopesRequest): Promise<string[]> {
    const res = await this.client.request<{ scopes: string[] }>("GET", "/v1/iam/effective-scopes", {
      query: { tenantId: data.tenantId, principalId: data.principalId, clientId: data.clientId },
    });
    return res.scopes;
  }

  // Tenant lifecycle
  async createTenant(data: CreateTenantRequest): Promise<CreateTenantResponse> {
    return this.client.request<CreateTenantResponse>("POST", "/v1/iam/tenants", { body: data });
  }

  async getTenantProvisioningStatus(tenantId: string): Promise<TenantProvisioningStatus> {
    return this.client.request<TenantProvisioningStatus>(
      "GET",
      `/v1/iam/tenants/${tenantId}/provisioning`,
    );
  }

  async deleteTenant(tenantId: string): Promise<{ accepted: boolean; provStatus: string }> {
    return this.client.request("DELETE", `/v1/iam/tenants/${tenantId}`);
  }

  // Admin tenant management
  async listTenants(options?: ListTenantsOptions): Promise<ListTenantsResponse> {
    return this.client.request<ListTenantsResponse>("GET", "/v1/iam/tenants", { query: options as any });
  }

  async getTenant(tenantId: string): Promise<Tenant> {
    return this.client.request<Tenant>("GET", `/v1/iam/tenants/${tenantId}`);
  }

  async updateTenant(tenantId: string, data: UpdateTenantRequest): Promise<Tenant> {
    return this.client.request<Tenant>("PUT", `/v1/iam/tenants/${tenantId}`, { body: data });
  }

  async suspendTenant(tenantId: string, reason?: string): Promise<SuspendTenantResponse> {
    return this.client.request<SuspendTenantResponse>("POST", `/v1/iam/tenants/${tenantId}/suspend`, {
      body: { reason },
    });
  }

  async listTenantQuotas(tenantId: string): Promise<TenantQuota[]> {
    const res = await this.client.request<{ quotas: TenantQuota[] }>(
      "GET",
      `/v1/iam/tenants/${tenantId}/quotas`,
    );
    return res.quotas;
  }

  async updateTenantQuota(tenantId: string, resource: string, limit: number): Promise<TenantQuota> {
    return this.client.request<TenantQuota>(
      "PUT",
      `/v1/iam/tenants/${tenantId}/quotas/${resource}`,
      { body: { limit } },
    );
  }

  async bulkTenantAction(data: BulkTenantActionRequest): Promise<BulkTenantActionResponse> {
    return this.client.request<BulkTenantActionResponse>("POST", "/v1/iam/tenants/bulk", { body: data });
  }
}
