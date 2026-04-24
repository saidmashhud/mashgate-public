import type { MashgateClient } from "../client.js";
import type { AuditLogEntry, LogsPage } from "../types.js";

export interface LogsQueryOptions {
  tenantId: string;
  from?: string;
  to?: string;
  cursor?: string;
  limit?: number;
}

export interface AuditLogsOptions extends LogsQueryOptions {
  actor?: string;
  action?: string;
}

export interface PaymentLogsOptions extends LogsQueryOptions {
  status?: string;
}

export interface WebhookLogsOptions extends LogsQueryOptions {
  endpointId?: string;
}

export interface ActivityLogsOptions extends LogsQueryOptions {
  type?: string;
}

export class LogsResource {
  constructor(private readonly client: MashgateClient) {}

  async audit(options: AuditLogsOptions): Promise<LogsPage<AuditLogEntry>> {
    return this.client.request<LogsPage<AuditLogEntry>>("GET", "/v1/logs/audit", {
      query: {
        tenantId: options.tenantId,
        from: options.from,
        to: options.to,
        actor: options.actor,
        action: options.action,
        cursor: options.cursor,
        limit: options.limit,
      },
    });
  }

  async activity(options: ActivityLogsOptions): Promise<LogsPage<AuditLogEntry>> {
    return this.client.request<LogsPage<AuditLogEntry>>("GET", "/v1/logs/activity", {
      query: {
        tenantId: options.tenantId,
        from: options.from,
        to: options.to,
        type: options.type,
        cursor: options.cursor,
        limit: options.limit,
      },
    });
  }

  async payments(options: PaymentLogsOptions): Promise<LogsPage<AuditLogEntry>> {
    return this.client.request<LogsPage<AuditLogEntry>>("GET", "/v1/logs/payments", {
      query: {
        tenantId: options.tenantId,
        from: options.from,
        to: options.to,
        status: options.status,
        cursor: options.cursor,
        limit: options.limit,
      },
    });
  }

  async webhooks(options: WebhookLogsOptions): Promise<LogsPage<AuditLogEntry>> {
    return this.client.request<LogsPage<AuditLogEntry>>("GET", "/v1/logs/webhooks", {
      query: {
        tenantId: options.tenantId,
        from: options.from,
        to: options.to,
        endpoint_id: options.endpointId,
        cursor: options.cursor,
        limit: options.limit,
      },
    });
  }
}
