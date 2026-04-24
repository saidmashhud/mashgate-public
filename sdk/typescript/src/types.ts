// ── Money ──────────────────────────────────────────────────────────────
export interface Money {
  amount: string;
  currency: string;
}

// ── Auth ───────────────────────────────────────────────────────────────
export interface RegisterRequest {
  email: string;
  password: string;
  fullName?: string;
  tenantId: string;
  role?: string;
}

export interface LoginRequest {
  email: string;
  password: string;
  /** Optional: scope login to a specific tenant when a user belongs to multiple tenants. */
  tenantId?: string;
}

/** Returned by register — no tokens, just the created user record. */
export interface RegisterResponse {
  userId: string;
  email: string;
  tenantId: string;
  createdAt: number;
}

/** Returned by login — short-lived access token + long-lived refresh token. */
export interface LoginResponse {
  accessToken: string;
  refreshToken: string;
  /** Unix timestamp (seconds) when accessToken expires. */
  expiresAt: number;
  user: UserProfile;
}

/** @deprecated Use LoginResponse. Kept for backward compatibility. */
export interface AuthTokens extends LoginResponse {}

export interface RefreshResponse {
  accessToken: string;
  expiresAt: number;
}

export interface UserProfile {
  userId: string;
  email: string;
  fullName?: string;
  role: string;
  tenantId: string;
  permissions?: string[];
  metadata?: Record<string, string>;
  createdAt: number;
  lastLoginAt?: number;
}

export interface UserCapabilities {
  userId: string;
  tenantId: string;
  roles: string[];
  permissions: string[];
  metadata?: Record<string, string>;
}

export interface ValidateTokenResponse {
  valid: boolean;
  /** Mashgate userId (maps to users.mashgate_principal_id in Grid). */
  userId: string;
  tenantId: string;
  role: string;
  permissions: string[];
  /** Unix timestamp (seconds). */
  expiresAt: number;
}

export interface UpdateProfileRequest {
  fullName?: string;
  email?: string;
  metadata?: Record<string, string>;
}

// ── Payments ───────────────────────────────────────────────────────────
export interface CreatePaymentRequest {
  amount: string;
  currency: string;
  orderId?: string;
  autoCapture?: boolean;
  idempotencyKey?: string;
  paymentMethodToken?: string;
  paymentMethodType?: string;
  paymentMethodBrand?: string;
  paymentMethodLast4?: string;
  paymentMethodBin?: string;
  paymentMethodProvider?: string;
  metadata?: Record<string, string>;
}

export interface PaymentIntent {
  paymentId: string;
  status: string;
  amount: Money;
}

export interface PaymentDetails {
  paymentId: string;
  status: string;
  amount: Money;
  authorizedAmount?: Money;
  capturedAmount?: Money;
  refundedAmount?: Money;
  provider: string;
  providerPaymentRef: string;
  orderId: string;
  customerId: string;
  createdAt: number;
  updatedAt: number;
  refunds: RefundSummary[];
}

export interface PaymentListResponse {
  data: PaymentDetails[];
  total: number;
  page: number;
  pageSize: number;
}

export interface CaptureRequest {
  amount?: string;
  currency?: string;
  idempotencyKey?: string;
}

export interface RefundRequest {
  amount: string;
  currency?: string;
  reason?: string;
  note?: string;
  idempotencyKey?: string;
}

export interface RefundSummary {
  refundId: string;
  status: string;
  amount: Money;
  reason: string;
  createdAt: number;
}

// ── Checkout ───────────────────────────────────────────────────────────
export interface CreateCheckoutSessionRequest {
  successUrl: string;
  cancelUrl: string;
  lineItems: CheckoutLineItem[];
  currency: string;
  expiresInMinutes?: number;
  metadata?: Record<string, string>;
}

export interface CheckoutLineItem {
  name: string;
  description?: string;
  quantity: number;
  unitPrice: Money;
}

export interface CheckoutSession {
  sessionId: string;
  status: string;
  url: string;
  totalAmount: Money;
  lineItems: CheckoutLineItem[];
  successUrl: string;
  cancelUrl: string;
  expiresAt: number;
  createdAt: number;
}

export interface CompleteCheckoutRequest {
  sessionId: string;
  paymentMethodToken: string;
  paymentMethodType: string;
  paymentMethodBrand?: string;
  paymentMethodLast4?: string;
  walletProvider?: WalletProvider;
  walletPhone?: string;
}

export type WalletProvider = "click" | "payme" | "oson" | "apple_pay" | "google_pay";

// ── Wallet ─────────────────────────────────────────────────────────────
export interface AddPaymentMethodRequest {
  token: string;
  provider?: string;
  brand?: string;
  last4?: string;
  expMonth?: number;
  expYear?: number;
  setDefault?: boolean;
}

export interface PaymentMethod {
  paymentMethodId: string;
  token: string;
  provider: string;
  brand: string;
  last4: string;
  expMonth: number;
  expYear: number;
  isDefault: boolean;
  createdAt: number;
}

export interface WalletBalance {
  availableBalance: string;
  holdsBalance: string;
  currency: string;
}

export interface WalletMovement {
  movementId: string;
  movementType: string;
  amount: string;
  currency: string;
  paymentId?: string;
  refundId?: string;
  createdAt: number;
}

export interface WalletMovementsResponse {
  movements: WalletMovement[];
  total: number;
}

// ── Risk ───────────────────────────────────────────────────────────────
export interface RiskAssessmentRequest {
  amount: string;
  currency: string;
  customerId?: string;
  metadata?: Record<string, string>;
}

export interface RiskScore {
  score: number;
  level: string;
}

export interface RiskFactor {
  factor: string;
  description: string;
  weight: number;
}

export interface RiskAssessmentResult {
  recommendation: string;
  riskScore: RiskScore;
  factors: RiskFactor[];
}

export interface RiskRule {
  ruleId: string;
  tenantId: string;
  name: string;
  description: string;
  ruleType: string;
  condition: string;
  weight: number;
  action: string;
  enabled: boolean;
  createdAt: number;
  updatedAt: number;
}

export interface CreateRuleRequest {
  name: string;
  description?: string;
  ruleType: string;
  condition: string;
  weight: number;
  action: string;
}

export interface UpdateRuleRequest {
  name?: string;
  description?: string;
  condition?: string;
  weight?: number;
  action?: string;
  enabled?: boolean;
}

export interface BlocklistEntry {
  entryId: string;
  entryType: string;
  value: string;
  reason: string;
  expiresAt: number;
  createdAt: number;
}

export interface AddBlocklistEntryRequest {
  entryType: string;
  value: string;
  reason?: string;
  expiresInSeconds?: number;
}

// ── Webhooks ───────────────────────────────────────────────────────────
export interface CreateWebhookEndpointRequest {
  url: string;
  description?: string;
  eventTypes?: string[];
}

export interface WebhookEndpoint {
  endpointId: string;
  url: string;
  description: string;
  eventTypes: string[];
  status: string;
  signingSecret?: string;
  createdAt: number;
  updatedAt: number;
}

export interface WebhookDelivery {
  deliveryId: string;
  endpointId: string;
  eventType: string;
  eventId: string;
  status: string;
  attemptCount: number;
  responseStatus: number;
  errorMessage: string;
  createdAt: number;
  deliveredAt: number;
}

/**
 * W3C Trace Context object carried under the envelope's `_trace` field.
 * See https://www.w3.org/TR/trace-context/
 */
export interface TraceContext {
  traceparent: string;
  tracestate?: string;
}

/**
 * Shape of the JSON body Mashgate (and HookLine on its behalf) POSTs to a
 * webhook endpoint.
 *
 * Envelope v1 (ADR-0013 §4, contracts/events/_envelope.v1.json) is the
 * canonical form: `id`, `topic`, `tenantId`, `occurredAt`, `createdAt`,
 * `payload`, optional `source` and `_trace`. Legacy emitters populate
 * `eventId`, `eventType`, `eventVersion`, and a flat `data` field instead of
 * `payload`. Consumers should prefer the envelope fields and fall back to
 * the legacy ones; use `eventPayload()` from `./resources/webhooks` for that.
 */
export interface WebhookEvent {
  // Envelope v1
  id?: string;
  topic?: string;
  createdAt?: number;
  payload?: unknown;
  source?: string;
  _trace?: TraceContext;

  // Present in both legacy and envelope v1 emissions
  eventId: string;
  tenantId: string;
  occurredAt: number;
  eventType?: string;
  eventVersion?: number;

  // Legacy (pre-envelope)
  correlationId?: string;
  aggregateId?: string;
  /** @deprecated Use `payload`. */
  data?: unknown;
}

/**
 * Envelope v1 topic constants (ADR-0013 §4 `<product>.<resource>.<verb>`).
 * Prefer these over the legacy dotted `event.type` strings.
 */
export const WebhookTopic = {
  PaymentCreated:           "payments.payment.created",
  PaymentCompleted:         "payments.payment.completed",
  PaymentFailed:            "payments.payment.failed",
  PaymentAuthorized:        "payments.payment.authorized",
  PaymentVoided:            "payments.payment.voided",
  RefundCreated:            "payments.refund.created",
  RefundCompleted:          "payments.refund.completed",
  RefundFailed:             "payments.refund.failed",
  CheckoutSessionCreated:   "payments.checkout_session.created",
  CheckoutSessionCompleted: "payments.checkout_session.completed",
  UserRegistered:           "iam.user.registered",
  NotificationSent:         "notifications.notification.sent",
} as const;
export type WebhookTopic = (typeof WebhookTopic)[keyof typeof WebhookTopic];

// ── Developer ──────────────────────────────────────────────────────────
export interface CreateApplicationRequest {
  name: string;
  appType: string;
  description?: string;
}

export interface Application {
  appId: string;
  tenantId: string;
  name: string;
  appType: string;
  status: string;
  description: string;
  metadata: Record<string, string>;
  createdAt: number;
  updatedAt: number;
}

// ── Settings ───────────────────────────────────────────────────────────
export interface MerchantSettings {
  tenantId: string;
  refundEnabled: boolean;
  maxRefundAmount: string;
  maxRefundsPerDay: number;
  autoRefundEnabled: boolean;
  policies: Record<string, unknown>;
}

export interface UpdateSettingsRequest {
  refundEnabled?: boolean;
  maxRefundAmount?: string;
  maxRefundsPerDay?: number;
  autoRefundEnabled?: boolean;
  policiesJson?: string;
}

// ── OIDC ───────────────────────────────────────────────────────────────
export interface OAuthClient {
  clientId: string;
  clientSecret?: string;
  name: string;
  redirectUris: string[];
  scopes: string[];
  grantTypes: string[];
}

export interface AuthorizeRequest {
  responseType: string;
  clientId: string;
  redirectUri: string;
  scope: string;
  state?: string;
  nonce?: string;
  codeChallenge?: string;
  codeChallengeMethod?: string;
}

export interface TokenRequest {
  grantType: string;
  code?: string;
  redirectUri?: string;
  clientId: string;
  clientSecret?: string;
  codeVerifier?: string;
  refreshToken?: string;
}

export interface TokenResponse {
  accessToken: string;
  tokenType: string;
  expiresIn: number;
  refreshToken?: string;
  idToken?: string;
  scope: string;
}

// ── Notify ─────────────────────────────────────────────────────────────
export interface Template {
  id: string;
  tenantId: string;
  templateKey: string;
  channels: string[];
  emailSubject?: string;
  emailBodyHtml?: string;
  smsText?: string;
  vars: string[];
  createdAt: string;
}

export interface NotificationLog {
  id: string;
  tenantId: string;
  channel: string;
  recipient: string;
  templateKey?: string;
  status: string;
  provider?: string;
  providerMsgId?: string;
  error?: string;
  sentAt: string;
}

// ── Chat ───────────────────────────────────────────────────────────────
export interface Channel {
  id: string;
  tenantId: string;
  channelId: string;
  name: string;
  channelType: string;
  memberIds: string[];
  createdAt: string;
}

export interface Message {
  id: string;
  tenantId: string;
  channelId: string;
  senderId: string;
  content?: string;
  contentType: string;
  payload?: string;
  deletedAt?: string;
  createdAt: string;
}

// ── Storage ────────────────────────────────────────────────────────────
export interface UploadUrlResponse {
  fileId: string;
  uploadUrl: string;
  key: string;
  expiresAt: string;
}

export interface FileEntry {
  fileId: string;
  tenantId: string;
  key: string;
  size: number;
  lastModified: string;
}

// ── Flags ──────────────────────────────────────────────────────────────
export interface Flag {
  id: string;
  tenantId: string;
  flagKey: string;
  enabled: boolean;
  rolloutPct: number;
  targetUsers: string[];
  targetGroups: string[];
  description?: string;
  createdAt: string;
  updatedAt: string;
}

export interface EvaluateResponse {
  flagKey: string;
  enabled: boolean;
  reason: string;
}

// ── Logs ───────────────────────────────────────────────────────────────
export interface AuditLogEntry {
  tenantId: string;
  appId?: string;
  actorId?: string;
  actorType?: string;
  action: string;
  resourceType: string;
  resourceId: string;
  changes?: string;
  ipAddress?: string;
  userAgent?: string;
  ts: string;
}

export interface LogsPage<T> {
  data: T[];
  nextCursor?: string;
}

// ── Subscriptions ──────────────────────────────────────────────────────
export interface Plan {
  id: string;
  tenantId: string;
  name: string;
  amount: number;
  currency: string;
  interval: "monthly" | "yearly" | "weekly";
  trialDays: number;
  active: boolean;
  createdAt: string;
}

export interface Subscription {
  id: string;
  tenantId: string;
  customerId: string;
  planId: string;
  status: "active" | "past_due" | "cancelled" | "trialing" | "paused";
  paymentMethodToken: string;
  currentPeriodStart?: string;
  currentPeriodEnd?: string;
  trialEndsAt?: string;
  cancelledAt?: string;
  retryCount: number;
  nextRetryAt?: string;
  createdAt: string;
}

// ── Invoices ────────────────────────────────────────────────────────────
export interface LineItem {
  description: string;
  quantity: number;
  unitAmount: number;
}

export interface Invoice {
  id: string;
  tenantId: string;
  invoiceNumber: string;
  customerId?: string;
  paymentId?: string;
  subscriptionId?: string;
  amount: number;
  currency: string;
  status: "draft" | "open" | "paid" | "void";
  lineItems: LineItem[];
  dueDate?: string;
  paidAt?: string;
  voidedAt?: string;
  createdAt: string;
  updatedAt: string;
}

// ── Payment Links ───────────────────────────────────────────────────────
export interface PaymentLink {
  id: string;
  tenantId: string;
  linkId: string;
  url: string;
  amount: number;
  currency: string;
  description?: string;
  expiresAt?: string;
  status: "active" | "paid" | "expired";
  paymentId?: string;
  createdAt: string;
}

// ── Guard ───────────────────────────────────────────────────────────────
export interface RateLimitConfig {
  id: string;
  tenantId: string;
  path: string;
  method: string;
  rpm: number;
  createdAt: string;
  updatedAt: string;
}

export interface IpBlocklistEntry {
  id: string;
  tenantId: string;
  ipAddress: string;
  reason?: string;
  expiresAt?: string;
  createdAt: string;
}

// ── IAM ────────────────────────────────────────────────────────────────────
export interface Permission {
  code: string;
  description: string;
  ownerService: string;
  resourceType: string;
}

export interface Role {
  roleId: string;
  tenantId: string;
  code: string;
  name: string;
  permissions: string[];
  systemRole: boolean;
  metadata: Record<string, string>;
}

export interface Group {
  groupId: string;
  tenantId: string;
  code: string;
  name: string;
  metadata: Record<string, string>;
}

export interface Policy {
  policyId: string;
  tenantId: string;
  code: string;
  description: string;
  conditionsJson: string;
  metadata: Record<string, string>;
}

export interface ApiKey {
  apiKeyId: string;
  tenantId: string;
  clientId: string;
  name: string;
  prefix: string;
  scopes: string[];
  rpm: number;
  burst: number;
  status: string;
  createdAt: number;
  rotatedAt?: number;
  expiresAt?: number;
}

export interface AuditEvent {
  eventId: string;
  tenantId: string;
  actorPrincipalId: string;
  action: string;
  targetType: string;
  targetId: string;
  detailsJson?: string;
  occurredAt: number;
}

export interface Tenant {
  tenantId: string;
  code: string;
  name: string;
  /** "sandbox" | "live" */
  mode: string;
  /** "active" | "suspended" | "pending" | "deleted" */
  status: string;
  planId?: string;
  planName?: string;
  userCount: number;
  createdAt: number;
  updatedAt: number;
  /** Arbitrary key-value pairs, e.g. { grid_workspace_id: "uuid" } */
  metadata: Record<string, string>;
}

export interface TenantQuota {
  tenantId: string;
  resource: string;
  limit: number;
  usage: number;
  planTier: string;
}

export interface AppScope {
  clientId: string;
  scopeCode: string;
  description: string;
  createdAt: number;
}

// ── Billing ───────────────────────────────────────────────────────────────

export interface BillingPlan {
  planId: string;
  name: string;
  description?: string;
  amount: string;
  currency: string;
  interval: "monthly" | "yearly";
  features: string[];
  active: boolean;
  createdAt: string;
}

export interface BillingSubscription {
  subscriptionId: string;
  planId: string;
  planName: string;
  status: "active" | "past_due" | "cancelled" | "trialing";
  currentPeriodStart: string;
  currentPeriodEnd: string;
  cancelAtPeriodEnd: boolean;
  trialEndsAt?: string;
  cancelledAt?: string;
  createdAt: string;
}

export interface BillingPaymentMethod {
  methodId: string;
  provider: string;
  brand: string;
  last4: string;
  expMonth: number;
  expYear: number;
  isDefault: boolean;
  createdAt: string;
}

export interface BillingInvoice {
  invoiceId: string;
  invoiceNumber: string;
  amount: string;
  currency: string;
  status: "draft" | "open" | "paid" | "void" | "uncollectible";
  periodStart: string;
  periodEnd: string;
  dueDate?: string;
  paidAt?: string;
  hostedUrl?: string;
  pdfUrl?: string;
  createdAt: string;
}

export interface CreditBalance {
  balance: string;
  currency: string;
  updatedAt: string;
}

// ── Analytics ─────────────────────────────────────────────────────────────

export interface PaymentMetrics {
  totalVolume: string;
  currency: string;
  totalTransactions: number;
  successRate: number;
  avgTransactionValue: string;
  refundRate: number;
  chargebackRate: number;
}

export interface TimeSeriesPoint {
  timestamp: string;
  value: number;
}

export interface VolumeTimeSeries {
  points: TimeSeriesPoint[];
  total: string;
  currency: string;
}

export interface TransactionCountSeries {
  points: TimeSeriesPoint[];
  total: number;
}

export interface PaymentMethodShare {
  method: string;
  count: number;
  volume: string;
  percentage: number;
}

export interface PaymentMethodBreakdown {
  methods: PaymentMethodShare[];
}

export interface GeoEntry {
  country: string;
  countryCode: string;
  count: number;
  volume: string;
  percentage: number;
}

export interface GeoDistribution {
  entries: GeoEntry[];
}

export interface FailureReason {
  code: string;
  description: string;
  count: number;
  percentage: number;
}

export interface FailureAnalysis {
  totalFailed: number;
  failureRate: number;
  reasons: FailureReason[];
}

export interface CustomerMetrics {
  totalCustomers: number;
  newCustomers: number;
  returningCustomers: number;
  churnRate: number;
  avgLifetimeValue: string;
  currency: string;
}

export interface CohortRow {
  cohort: string;
  size: number;
  retentionByPeriod: number[];
}

export interface CohortAnalysis {
  cohorts: CohortRow[];
  periods: string[];
}

export interface CustomerSegment {
  segmentId: string;
  name: string;
  description: string;
  customerCount: number;
  avgTransactionValue: string;
  totalVolume: string;
  currency: string;
}

export interface TopCustomer {
  customerId: string;
  email?: string;
  name?: string;
  totalSpent: string;
  currency: string;
  transactionCount: number;
  lastTransactionAt: string;
}

// ── Client Options ─────────────────────────────────────────────────────
export interface MashgateClientOptions {
  baseUrl: string;
  apiKey?: string;
  accessToken?: string;
  fetch?: typeof globalThis.fetch;
  headers?: Record<string, string>;
  /** Request timeout in milliseconds. Default: 30 000. */
  timeout?: number;
  /** Maximum number of retries on 5xx/network errors. Default: 0. */
  maxRetries?: number;
  /**
   * Factory that produces a unique Idempotency-Key for each mutating request.
   * @example `() => crypto.randomUUID()`
   */
  idempotencyKey?: () => string;
}
