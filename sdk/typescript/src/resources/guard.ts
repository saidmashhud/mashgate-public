import type { MashgateClient } from "../client.js";
import type { RateLimitConfig, IpBlocklistEntry } from "../types.js";

export interface CheckRequest {
  tenantId: string;
  path: string;
  method: string;
  ip: string;
}

export interface CheckResult {
  allowed: boolean;
  remaining: number;
  resetAt: number;
  reason?: string;
}

export interface UpsertRateLimitRequest {
  tenantId: string;
  path: string;
  method: string;
  rpm: number;
}

export interface BlockIpRequest {
  tenantId: string;
  ip: string;
  reason?: string;
  expiresAt?: number;
}

export class GuardResource {
  constructor(private readonly client: MashgateClient) {}

  async check(data: CheckRequest): Promise<CheckResult> {
    return this.client.request<CheckResult>("POST", "/v1/guard/check", { body: data });
  }

  async upsertRateLimit(data: UpsertRateLimitRequest): Promise<RateLimitConfig> {
    return this.client.request<RateLimitConfig>("POST", "/v1/guard/rate-limits", { body: data });
  }

  async listRateLimits(tenantId: string): Promise<RateLimitConfig[]> {
    return this.client.request<RateLimitConfig[]>("GET", "/v1/guard/rate-limits", {
      query: { tenantId },
    });
  }

  async blockIp(data: BlockIpRequest): Promise<IpBlocklistEntry> {
    return this.client.request<IpBlocklistEntry>("POST", "/v1/guard/blocklist/ips", { body: data });
  }

  async unblockIp(tenantId: string, ip: string): Promise<{ success: boolean }> {
    return this.client.request<{ success: boolean }>(
      "DELETE",
      `/v1/guard/blocklist/ips/${encodeURIComponent(ip)}`,
      { query: { tenantId } },
    );
  }

  async listBlockedIps(tenantId: string): Promise<IpBlocklistEntry[]> {
    return this.client.request<IpBlocklistEntry[]>("GET", "/v1/guard/blocklist/ips", {
      query: { tenantId },
    });
  }
}
