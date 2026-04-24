import type { MashgateClient } from "../client.js";
import type { Flag, EvaluateResponse } from "../types.js";

export interface CreateFlagRequest {
  tenantId: string;
  flagKey: string;
  enabled: boolean;
  rolloutPct?: number;
  targetUsers?: string[];
  targetGroups?: string[];
  description?: string;
}

export interface UpdateFlagRequest {
  enabled?: boolean;
  rolloutPct?: number;
  targetUsers?: string[];
  targetGroups?: string[];
  description?: string;
}

export interface EvaluateFlagRequest {
  tenantId: string;
  flagKey: string;
  userId?: string;
  groups?: string[];
}

export class FlagsResource {
  constructor(private readonly client: MashgateClient) {}

  async create(data: CreateFlagRequest): Promise<Flag> {
    return this.client.request<Flag>("POST", "/v1/flags", { body: data });
  }

  async list(tenantId: string): Promise<Flag[]> {
    return this.client.request<Flag[]>("GET", "/v1/flags", { query: { tenantId } });
  }

  async get(flagKey: string, tenantId: string): Promise<Flag> {
    return this.client.request<Flag>("GET", `/v1/flags/${flagKey}`, { query: { tenantId } });
  }

  async update(flagKey: string, tenantId: string, data: UpdateFlagRequest): Promise<Flag> {
    return this.client.request<Flag>("PUT", `/v1/flags/${flagKey}`, {
      body: data,
      query: { tenantId },
    });
  }

  async evaluate(data: EvaluateFlagRequest): Promise<EvaluateResponse> {
    return this.client.request<EvaluateResponse>("POST", "/v1/flags/evaluate", { body: data });
  }
}
