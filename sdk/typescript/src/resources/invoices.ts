import type { MashgateClient } from "../client.js";
import type { Invoice, LineItem } from "../types.js";

export interface CreateInvoiceRequest {
  tenantId: string;
  customerId?: string;
  paymentId?: string;
  subscriptionId?: string;
  amount: number;
  currency: string;
  lineItems?: LineItem[];
  dueDate?: string;
}

export class InvoicesResource {
  constructor(private readonly client: MashgateClient) {}

  async create(data: CreateInvoiceRequest): Promise<Invoice> {
    return this.client.request<Invoice>("POST", "/v1/invoices", { body: data });
  }

  async list(tenantId: string, status?: string): Promise<Invoice[]> {
    return this.client.request<Invoice[]>("GET", "/v1/invoices", {
      query: { tenantId, ...(status ? { status } : {}) },
    });
  }

  async get(id: string): Promise<Invoice> {
    return this.client.request<Invoice>("GET", `/v1/invoices/${id}`);
  }

  async getPdfUrl(id: string): Promise<string> {
    return `/v1/invoices/${id}/pdf`;
  }

  async send(id: string): Promise<{ sent: boolean }> {
    return this.client.request<{ sent: boolean }>("POST", `/v1/invoices/${id}/send`);
  }

  async void(id: string): Promise<Invoice> {
    return this.client.request<Invoice>("POST", `/v1/invoices/${id}/void`);
  }
}
