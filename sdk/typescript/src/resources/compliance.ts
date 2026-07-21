import type { MashgateClient } from "../client.js";

export const AlertStatus = {
  Open: "ALERT_STATUS_OPEN",
  UnderReview: "ALERT_STATUS_UNDER_REVIEW",
  Resolved: "ALERT_STATUS_RESOLVED",
  Escalated: "ALERT_STATUS_ESCALATED",
  SARFiled: "ALERT_STATUS_SAR_FILED",
  Closed: "ALERT_STATUS_CLOSED",
} as const;
export type AlertStatus = (typeof AlertStatus)[keyof typeof AlertStatus];

export const AlertSeverity = {
  Low: "ALERT_SEVERITY_LOW",
  Medium: "ALERT_SEVERITY_MEDIUM",
  High: "ALERT_SEVERITY_HIGH",
  Critical: "ALERT_SEVERITY_CRITICAL",
} as const;
export type AlertSeverity = (typeof AlertSeverity)[keyof typeof AlertSeverity];

export const AlertCategory = {
  AML: "ALERT_CATEGORY_AML",
  Sanctions: "ALERT_CATEGORY_SANCTIONS",
  PEP: "ALERT_CATEGORY_PEP",
  Fraud: "ALERT_CATEGORY_FRAUD",
  Transaction: "ALERT_CATEGORY_TRANSACTION",
  KYC: "ALERT_CATEGORY_KYC",
  Watchlist: "ALERT_CATEGORY_WATCHLIST",
} as const;
export type AlertCategory = (typeof AlertCategory)[keyof typeof AlertCategory];

export interface AlertEvidence {
  evidenceType: string;
  referenceId: string;
  description: string;
  url: string;
}

export interface ComplianceAlert {
  alertId: string;
  tenantId: string;
  subjectId: string;
  subjectType: string;
  category: AlertCategory;
  severity: AlertSeverity;
  status: AlertStatus;
  source: string;
  sourceRef: string;
  description: string;
  evidence: AlertEvidence[];
  assignedTo: string;
  resolvedBy: string;
  resolveNote: string;
  escalatedTo: string;
  escalateNote: string;
  sarId: string;
  createdAt: string;
  updatedAt: string;
  dueAt?: string;
}

export interface RaiseComplianceAlert {
  subjectId: string;
  subjectType: string;
  category: AlertCategory;
  severity: AlertSeverity;
  source: string;
  sourceRef: string;
  description: string;
  evidence?: AlertEvidence[];
}

export interface ListComplianceAlertsQuery {
  subjectId?: string;
  status?: AlertStatus;
  severity?: AlertSeverity;
  limit?: number;
  cursor?: string;
}

export interface ListComplianceAlertsResponse {
  alerts: ComplianceAlert[];
  nextCursor?: string;
}

export class ComplianceResource {
  constructor(private readonly client: MashgateClient) {}

  raise(data: RaiseComplianceAlert, idempotencyKey: string): Promise<ComplianceAlert> {
    return this.client.request("POST", "/v1/compliance/alerts", {
      body: { ...data, tenantId: this.client.requireTenantId() },
      headers: { "Idempotency-Key": idempotencyKey },
    });
  }

  get(alertId: string): Promise<ComplianceAlert> {
    return this.client.request("GET", `/v1/compliance/alerts/${alertId}`, {
      query: { tenant_id: this.client.requireTenantId() },
    });
  }

  list(query: ListComplianceAlertsQuery = {}): Promise<ListComplianceAlertsResponse> {
    return this.client.request("GET", "/v1/compliance/alerts", {
      query: {
        tenant_id: this.client.requireTenantId(),
        subject_id: query.subjectId,
        status: query.status,
        severity: query.severity,
        limit: query.limit,
        cursor: query.cursor,
      },
    });
  }

  resolve(alertId: string, resolveNote: string): Promise<ComplianceAlert> {
    return this.client.request("POST", `/v1/compliance/alerts/${alertId}/resolve`, {
      body: { tenant_id: this.client.requireTenantId(), alert_id: alertId, resolve_note: resolveNote },
    });
  }

  escalate(alertId: string, escalatedTo: string, escalateNote: string): Promise<ComplianceAlert> {
    return this.client.request("POST", `/v1/compliance/alerts/${alertId}/escalate`, {
      body: {
        tenant_id: this.client.requireTenantId(),
        alert_id: alertId,
        escalated_to: escalatedTo,
        escalate_note: escalateNote,
      },
    });
  }

  async hasOpenAlerts(subjectId: string): Promise<boolean> {
    const response = await this.list({ subjectId, status: AlertStatus.Open, limit: 1 });
    return response.alerts.length > 0;
  }
}
