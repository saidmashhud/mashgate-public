import { MashgateError } from "./errors.js";
import type { MashgateClientOptions } from "./types.js";
import { AuthResource } from "./resources/auth.js";
import { PaymentsResource } from "./resources/payments.js";
import { CheckoutResource } from "./resources/checkout.js";
import { WalletResource } from "./resources/wallet.js";
import { RiskResource } from "./resources/risk.js";
import { WebhooksResource } from "./resources/webhooks.js";
import { DeveloperResource } from "./resources/developer.js";
import { SettingsResource } from "./resources/settings.js";
import { ChatResource } from "./resources/chat.js";
import { NotifyResource } from "./resources/notify.js";
import { StorageResource } from "./resources/storage.js";
import { FlagsResource } from "./resources/flags.js";
import { LogsResource } from "./resources/logs.js";
import { SubscriptionsResource } from "./resources/subscriptions.js";
import { InvoicesResource } from "./resources/invoices.js";
import { PaymentLinksResource } from "./resources/paymentLinks.js";
import { GuardResource } from "./resources/guard.js";
import { ChainResource } from "./resources/chain.js";
import { LocalPaymentsResource } from "./resources/localPayments.js";
import { IamResource } from "./resources/iam.js";
import { MeteringResource } from "./resources/metering.js";
import { BillingResource } from "./resources/billing.js";
import { AnalyticsResource } from "./resources/analytics.js";
import { WalletAdminResource } from "./resources/walletAdmin.js";
import { MailResource } from "./resources/mail.js";

export class MashgateClient {
  private readonly baseUrl: string;
  private readonly fetchFn: typeof globalThis.fetch;
  private readonly defaultHeaders: Record<string, string>;
  private readonly timeout: number;
  private readonly maxRetries: number;
  private readonly idempotencyKeyFn?: () => string;
  private accessToken?: string;

  readonly auth: AuthResource;
  readonly payments: PaymentsResource;
  readonly checkout: CheckoutResource;
  readonly wallet: WalletResource;
  readonly risk: RiskResource;
  readonly webhooks: WebhooksResource;
  readonly developer: DeveloperResource;
  readonly settings: SettingsResource;
  readonly chat: ChatResource;
  readonly notify: NotifyResource;
  readonly storage: StorageResource;
  readonly flags: FlagsResource;
  readonly logs: LogsResource;
  readonly subscriptions: SubscriptionsResource;
  readonly invoices: InvoicesResource;
  readonly paymentLinks: PaymentLinksResource;
  readonly guard: GuardResource;
  readonly chain: ChainResource;
  readonly localPayments: LocalPaymentsResource;
  readonly iam: IamResource;
  readonly metering: MeteringResource;
  readonly billing: BillingResource;
  readonly analytics: AnalyticsResource;
  /**
   * Admin/merchant-side WalletService — full `wallet.v1.WalletService` for
   * tenant-scoped operations (create/freeze/credit/debit/withdraw, on-chain
   * wallet creation, deposit-address resolution with SPL ATA support).
   * For end-user wallet operations (saved cards, balance), see `wallet`.
   */
  readonly walletAdmin: WalletAdminResource;
  /**
   * Mail capability — `mail.v1.MailService` (ADR-0019). Self-service mailbox
   * operations (read/send/update/delete) and admin-scoped tenant operations
   * (mailboxes, domains, DKIM rotation). Backed by Mashgate
   * `mail-service` in `services/orchestration/`. Subscribe to `mail.received`
   * / `mail.sent` / `mail.delivered` / `mail.bounced` events via webhooks.
   */
  readonly mail: MailResource;

  constructor(options: MashgateClientOptions) {
    this.baseUrl = options.baseUrl.replace(/\/+$/, "");
    this.fetchFn = options.fetch ?? globalThis.fetch.bind(globalThis);
    this.timeout = options.timeout ?? 30_000;
    this.maxRetries = options.maxRetries ?? 0;
    this.idempotencyKeyFn = options.idempotencyKey;
    this.accessToken = options.accessToken;

    this.defaultHeaders = {
      "Content-Type": "application/json",
      ...(options.apiKey ? { "X-API-Key": options.apiKey } : {}),
      ...options.headers,
    };

    this.auth = new AuthResource(this);
    this.payments = new PaymentsResource(this);
    this.checkout = new CheckoutResource(this);
    this.wallet = new WalletResource(this);
    this.risk = new RiskResource(this);
    this.webhooks = new WebhooksResource(this);
    this.developer = new DeveloperResource(this);
    this.settings = new SettingsResource(this);
    this.chat = new ChatResource(this);
    this.notify = new NotifyResource(this);
    this.storage = new StorageResource(this);
    this.flags = new FlagsResource(this);
    this.logs = new LogsResource(this);
    this.subscriptions = new SubscriptionsResource(this);
    this.invoices = new InvoicesResource(this);
    this.paymentLinks = new PaymentLinksResource(this);
    this.guard = new GuardResource(this);
    this.chain = new ChainResource(this);
    this.localPayments = new LocalPaymentsResource(this);
    this.iam = new IamResource(this);
    this.metering = new MeteringResource(this);
    this.billing = new BillingResource(this);
    this.analytics = new AnalyticsResource(this);
    this.walletAdmin = new WalletAdminResource(this);
    this.mail = new MailResource(this);
  }

  setAccessToken(token: string | undefined): void {
    this.accessToken = token;
  }

  async request<T>(
    method: string,
    path: string,
    options?: {
      body?: unknown;
      query?: Record<string, string | number | boolean | undefined>;
      headers?: Record<string, string>;
    },
  ): Promise<T> {
    const url = new URL(`${this.baseUrl}${path}`);

    if (options?.query) {
      for (const [key, value] of Object.entries(options.query)) {
        if (value !== undefined) url.searchParams.set(key, String(value));
      }
    }

    const reqHeaders: Record<string, string> = { ...this.defaultHeaders };
    if (this.accessToken) reqHeaders["Authorization"] = `Bearer ${this.accessToken}`;
    if (options?.headers) Object.assign(reqHeaders, options.headers);

    // Auto-attach idempotency key for mutating requests (generated once, reused on retries)
    if (
      this.idempotencyKeyFn &&
      ["POST", "PUT", "PATCH"].includes(method.toUpperCase()) &&
      !reqHeaders["Idempotency-Key"]
    ) {
      reqHeaders["Idempotency-Key"] = this.idempotencyKeyFn();
    }

    const maxAttempts = 1 + this.maxRetries;
    let lastError: MashgateError | undefined;

    for (let attempt = 0; attempt < maxAttempts; attempt++) {
      if (attempt > 0) {
        // Exponential backoff with jitter: 1s, 2s, 4s … up to 32s, plus up to 1s of jitter
        const delay =
          Math.min(1_000 * Math.pow(2, attempt - 1), 32_000) + Math.random() * 1_000;
        await new Promise<void>((resolve) => setTimeout(resolve, delay));
      }

      const controller = new AbortController();
      const timer = setTimeout(() => controller.abort(), this.timeout);

      let response: Response;
      try {
        response = await this.fetchFn(url.toString(), {
          method,
          headers: reqHeaders,
          body: options?.body !== undefined ? JSON.stringify(options.body) : undefined,
          signal: controller.signal,
        });
      } catch (err) {
        clearTimeout(timer);
        const isAbort = err instanceof Error && err.name === "AbortError";
        const error = new MashgateError(
          isAbort
            ? {
                message: `Request timed out after ${this.timeout}ms`,
                status: 408,
                code: "request_timeout",
                retryable: true,
              }
            : {
                message: err instanceof Error ? err.message : "Network request failed",
                status: 0,
                code: "network_error",
                retryable: true,
              },
        );
        if (attempt < maxAttempts - 1) {
          lastError = error;
          continue;
        }
        throw error;
      }

      clearTimeout(timer);
      const requestId = response.headers.get("x-request-id") ?? undefined;

      if (!response.ok) {
        let errorBody: Record<string, unknown> = {};
        try {
          errorBody = (await response.json()) as Record<string, unknown>;
        } catch {
          // ignore parse errors
        }

        const error = new MashgateError({
          message:
            (errorBody.message as string) ||
            (errorBody.error as string) ||
            `HTTP ${response.status} ${response.statusText}`,
          status: response.status,
          code: (errorBody.code as string) || undefined,
          param: (errorBody.param as string) || undefined,
          requestId,
          details: errorBody,
        });

        if (error.retryable && attempt < maxAttempts - 1) {
          lastError = error;
          continue;
        }
        throw error;
      }

      if (response.status === 204) return undefined as T;

      return (await response.json()) as T;
    }

    throw lastError!;
  }
}
