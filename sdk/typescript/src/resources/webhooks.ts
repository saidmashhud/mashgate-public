import type { MashgateClient } from "../client.js";
import type {
  CreateWebhookEndpointRequest,
  WebhookEndpoint,
  WebhookDelivery,
  WebhookEvent,
} from "../types.js";
import { verifyWebhookSignature } from "../webhooks.js";

type WebhookHandler = (event: WebhookEvent) => void | Promise<void>;

/**
 * Extract the business payload from a webhook event regardless of whether
 * it was emitted in envelope v1 form (`event.payload`) or the legacy flat
 * form (`event.data`). Returns an empty object when neither is present.
 */
export function eventPayload(event: WebhookEvent): unknown {
  if (event.payload !== undefined) return event.payload;
  if (event.data !== undefined) return event.data;
  return {};
}

/**
 * Returns the canonical routing key for a webhook event: the envelope v1
 * `topic` when present, otherwise the legacy `eventType`.
 */
export function eventKey(event: WebhookEvent): string {
  return event.topic ?? event.eventType ?? "";
}

export class WebhooksResource {
  private readonly handlers = new Map<string, WebhookHandler[]>();

  constructor(private readonly client: MashgateClient) {}

  // ── Endpoint management ─────────────────────────────────────────────

  async createEndpoint(data: CreateWebhookEndpointRequest): Promise<{ endpoint: WebhookEndpoint }> {
    return this.client.request("POST", "/v1/events/endpoints", { body: data });
  }

  async listEndpoints(): Promise<{ endpoints: WebhookEndpoint[] }> {
    return this.client.request("GET", "/v1/events/endpoints");
  }

  async getEndpoint(endpointId: string): Promise<WebhookEndpoint> {
    return this.client.request<WebhookEndpoint>("GET", `/v1/events/endpoints/${endpointId}`);
  }

  async deleteEndpoint(endpointId: string): Promise<void> {
    return this.client.request<void>("DELETE", `/v1/events/endpoints/${endpointId}`);
  }

  async listDeliveries(options: {
    endpointId: string;
    page?: number;
    pageSize?: number;
  }): Promise<{ deliveries: WebhookDelivery[] }> {
    return this.client.request("GET", "/v1/events/deliveries", {
      query: {
        endpoint_id: options.endpointId,
        page: options.page,
        page_size: options.pageSize,
      },
    });
  }

  async retryDelivery(deliveryId: string): Promise<void> {
    return this.client.request<void>("POST", `/v1/events/deliveries/${deliveryId}/retry`);
  }

  async testEndpoint(endpointId: string): Promise<void> {
    return this.client.request<void>("POST", `/v1/events/endpoints/${endpointId}/test`);
  }

  // ── Signature verification ──────────────────────────────────────────

  /**
   * Verify an incoming webhook signature using HMAC-SHA256.
   *
   * @param body      Raw request body string (before any JSON parsing)
   * @param signature `x-hl-signature` header value (e.g. `"v1=abc123..."`)
   * @param secret    Signing secret from your webhook endpoint settings
   * @param timestamp `x-hl-timestamp` header value (Unix epoch string)
   * @returns `true` if the signature is valid and timestamp is within 5 minutes
   *
   * @example
   * ```typescript
   * const valid = await mg.webhooks.verify(rawBody, req.headers['x-hl-signature'], secret, req.headers['x-hl-timestamp'])
   * if (!valid) return res.status(401).end()
   * ```
   */
  async verify(
    body: string | Uint8Array,
    signature: string,
    secret: string,
    timestamp: string,
  ): Promise<boolean> {
    return verifyWebhookSignature(body, signature, secret, timestamp);
  }

  // ── Event routing ───────────────────────────────────────────────────

  /**
   * Register a handler for a specific event type (or `"*"` for all events).
   * Used together with `handleRequest()`.
   *
   * @example
   * ```typescript
   * mg.webhooks.on('payment.captured', async (event) => {
   *   await fulfillOrder(event.data.paymentId)
   * })
   * mg.webhooks.on('*', (event) => console.log('received', event.eventType))
   * ```
   */
  on(eventType: string, handler: WebhookHandler): this {
    if (!this.handlers.has(eventType)) this.handlers.set(eventType, []);
    this.handlers.get(eventType)!.push(handler);
    return this;
  }

  /**
   * Verify and dispatch an incoming webhook to registered handlers.
   * Throws if the signature is invalid.
   *
   * @example
   * ```typescript
   * app.post('/webhooks', express.raw({ type: '*\/*' }), async (req, res) => {
   *   await mg.webhooks.handleRequest(req.body.toString(), req.headers['x-hl-signature'], secret, req.headers['x-hl-timestamp'])
   *   res.json({ ok: true })
   * })
   * ```
   */
  async handleRequest(
    body: string,
    signature: string,
    secret: string,
    timestamp: string,
  ): Promise<void> {
    const valid = await this.verify(body, signature, secret, timestamp);
    if (!valid) throw new Error("Invalid webhook signature");

    const event = JSON.parse(body) as WebhookEvent;
    // Dispatch on envelope v1 `topic` when present, falling back to the
    // legacy `eventType` so pre- and post-migration producers both work.
    const eventType = eventKey(event);

    const specificHandlers = this.handlers.get(eventType) ?? [];
    const wildcardHandlers = this.handlers.get("*") ?? [];

    for (const handler of [...specificHandlers, ...wildcardHandlers]) {
      await handler(event);
    }
  }
}
