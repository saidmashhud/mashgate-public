import type { MashgateClient } from "../client.js";
import type {
  RiskAssessmentRequest,
  RiskAssessmentResult,
  RiskRule,
  CreateRuleRequest,
  UpdateRuleRequest,
  BlocklistEntry,
  AddBlocklistEntryRequest,
} from "../types.js";

export class RiskResource {
  constructor(private readonly client: MashgateClient) {}

  async assessPayment(data: RiskAssessmentRequest): Promise<RiskAssessmentResult> {
    return this.client.request<RiskAssessmentResult>("POST", "/v1/risk/assess/payment", {
      body: data,
    });
  }

  async assessRefund(
    paymentId: string,
    amount: string,
    currency: string,
  ): Promise<RiskAssessmentResult> {
    return this.client.request<RiskAssessmentResult>("POST", "/v1/risk/assess/refund", {
      body: { paymentId, amount, currency },
    });
  }

  async investigatePayment(
    paymentId: string,
    depth?: number,
  ): Promise<{
    graphData: { nodes: unknown[]; edges: unknown[] };
    findings: Array<{ findingType: string; description: string }>;
  }> {
    return this.client.request("POST", "/v1/risk/investigate", {
      body: { paymentId, depth },
    });
  }

  async listRules(): Promise<{ rules: RiskRule[] }> {
    return this.client.request("GET", "/v1/risk/rules");
  }

  async createRule(data: CreateRuleRequest): Promise<RiskRule> {
    return this.client.request<RiskRule>("POST", "/v1/risk/rules", { body: data });
  }

  async updateRule(ruleId: string, data: UpdateRuleRequest): Promise<RiskRule> {
    return this.client.request<RiskRule>("PUT", `/v1/risk/rules/${ruleId}`, { body: data });
  }

  async deleteRule(ruleId: string): Promise<void> {
    return this.client.request<void>("DELETE", `/v1/risk/rules/${ruleId}`);
  }

  async listBlocklist(entryType?: string): Promise<{ entries: BlocklistEntry[] }> {
    return this.client.request("GET", "/v1/risk/blocklist", {
      query: { entry_type: entryType },
    });
  }

  async addBlocklistEntry(data: AddBlocklistEntryRequest): Promise<BlocklistEntry> {
    return this.client.request<BlocklistEntry>("POST", "/v1/risk/blocklist", { body: data });
  }

  async removeBlocklistEntry(entryId: string): Promise<void> {
    return this.client.request<void>("DELETE", `/v1/risk/blocklist/${entryId}`);
  }
}
