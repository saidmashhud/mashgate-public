import type { MashgateClient } from "../client.js";
import type { Template, NotificationLog } from "../types.js";

export interface SendSmsRequest {
  tenantId: string;
  to: string;
  templateKey?: string;
  vars?: Record<string, string>;
  text?: string;
}

export interface SendEmailRequest {
  tenantId: string;
  to: string;
  templateKey: string;
  vars?: Record<string, string>;
}

export interface CreateTemplateRequest {
  tenantId: string;
  templateKey: string;
  channels: string[];
  emailSubject?: string;
  emailBodyHtml?: string;
  smsText?: string;
  vars?: string[];
}

export interface SetSmsProviderRequest {
  tenantId: string;
  /** "osonsms" */
  provider: string;
  login: string;
  /** Empty on update keeps the stored secret. */
  secret?: string;
  sender: string;
  baseUrl?: string;
  isConfidential?: boolean;
}

/** Masked view — the secret itself is never returned. */
export interface SmsProviderInfo {
  tenantId: string;
  provider: string;
  login: string;
  sender: string;
  baseUrl?: string;
  isConfidential: boolean;
  hasSecret: boolean;
  secretPreview?: string;
  createdAt: string;
  updatedAt: string;
  lastCheckAt?: string;
  lastCheckOk: boolean;
  lastCheckError?: string;
  lastBalance?: string;
}

export interface TestSmsProviderResponse {
  ok: boolean;
  balance?: string;
  error?: string;
}

export interface ListLogsOptions {
  tenantId: string;
  from?: string;
  to?: string;
  page?: number;
}

export class NotifyResource {
  constructor(private readonly client: MashgateClient) {}

  async sendSms(data: SendSmsRequest): Promise<NotificationLog> {
    return this.client.request<NotificationLog>("POST", "/v1/notify/sms", { body: data });
  }

  async sendEmail(data: SendEmailRequest): Promise<NotificationLog> {
    return this.client.request<NotificationLog>("POST", "/v1/notify/email", { body: data });
  }

  async createTemplate(data: CreateTemplateRequest): Promise<Template> {
    return this.client.request<Template>("POST", "/v1/notify/templates", { body: data });
  }

  async listTemplates(tenantId: string): Promise<Template[]> {
    return this.client.request<Template[]>("GET", "/v1/notify/templates", {
      query: { tenantId },
    });
  }

  async listLogs(options: ListLogsOptions): Promise<NotificationLog[]> {
    return this.client.request<NotificationLog[]>("GET", "/v1/notify/logs", {
      query: {
        tenantId: options.tenantId,
        from: options.from,
        to: options.to,
        page: options.page,
      },
    });
  }

  // ── Tenant SMS provider (v1.10.0) — requires notify:providers:manage ──────

  async getSmsProvider(tenantId: string): Promise<SmsProviderInfo> {
    return this.client.request<SmsProviderInfo>("GET", "/v1/notify/sms-provider", {
      query: { tenantId },
    });
  }

  async setSmsProvider(data: SetSmsProviderRequest): Promise<SmsProviderInfo> {
    return this.client.request<SmsProviderInfo>("PUT", "/v1/notify/sms-provider", { body: data });
  }

  async deleteSmsProvider(tenantId: string): Promise<void> {
    await this.client.request<unknown>("DELETE", "/v1/notify/sms-provider", {
      query: { tenantId },
    });
  }

  async testSmsProvider(tenantId: string): Promise<TestSmsProviderResponse> {
    return this.client.request<TestSmsProviderResponse>("POST", "/v1/notify/sms-provider/test", {
      body: { tenantId },
    });
  }
}
