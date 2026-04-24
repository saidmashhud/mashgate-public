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
}
