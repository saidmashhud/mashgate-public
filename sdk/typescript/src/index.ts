export { MashgateClient } from "./client.js";
export { MashgateError } from "./errors.js";
export { verifyWebhookSignature } from "./webhooks.js";
export { WebhookTopic } from "./types.js";
export { eventPayload, eventKey } from "./resources/webhooks.js";

// Resource classes
export { AuthResource } from "./resources/auth.js";
export { PaymentsResource } from "./resources/payments.js";
export { CheckoutResource } from "./resources/checkout.js";
export { WalletResource } from "./resources/wallet.js";
export { RiskResource } from "./resources/risk.js";
export { WebhooksResource } from "./resources/webhooks.js";
export { DeveloperResource } from "./resources/developer.js";
export { SettingsResource } from "./resources/settings.js";
export { ChatResource } from "./resources/chat.js";
export { NotifyResource } from "./resources/notify.js";
export { StorageResource } from "./resources/storage.js";
export { FlagsResource } from "./resources/flags.js";
export { LogsResource } from "./resources/logs.js";
export { SubscriptionsResource } from "./resources/subscriptions.js";
export { InvoicesResource } from "./resources/invoices.js";
export { PaymentLinksResource } from "./resources/paymentLinks.js";
export { GuardResource } from "./resources/guard.js";
export { ChainResource } from "./resources/chain.js";
export { LocalPaymentsResource } from "./resources/localPayments.js";
export { IamResource } from "./resources/iam.js";
export { MeteringResource } from "./resources/metering.js";
export { BillingResource } from "./resources/billing.js";
export { AnalyticsResource } from "./resources/analytics.js";
export {
  WalletAdminResource,
  Currency,
  Network,
  Mint,
  WalletType,
  WalletStatus,
  TransactionType,
  TransactionStatus,
  TransactionReason,
} from "./resources/walletAdmin.js";
export type {
  Wallet as AdminWallet,
  WalletTransaction as AdminWalletTransaction,
  DepositAddress,
  CreateWalletRequest as AdminCreateWalletRequest,
  CreateChainWalletRequest,
  CreateChainWalletResponse,
  CreditDebitRequest,
  InitiateWithdrawalRequest,
  ListWalletsQuery,
  ListWalletsResponse as AdminListWalletsResponse,
  ListTransactionsQuery,
  ListTransactionsResponse as AdminListTransactionsResponse,
} from "./resources/walletAdmin.js";

// Types
export type {
  // Common
  Money,
  MashgateClientOptions,

  // Auth
  RegisterRequest,
  RegisterResponse,
  LoginRequest,
  LoginResponse,
  AuthTokens,
  RefreshResponse,
  UserProfile,
  UserCapabilities,
  ValidateTokenResponse,
  UpdateProfileRequest,

  // Payments
  CreatePaymentRequest,
  PaymentIntent,
  PaymentDetails,
  PaymentListResponse,
  CaptureRequest,
  RefundRequest,
  RefundSummary,

  // Checkout
  CreateCheckoutSessionRequest,
  CheckoutLineItem,
  CheckoutSession,
  CompleteCheckoutRequest,

  // Wallet
  AddPaymentMethodRequest,
  PaymentMethod,
  WalletBalance,
  WalletMovement,
  WalletMovementsResponse,

  // Risk
  RiskAssessmentRequest,
  RiskScore,
  RiskFactor,
  RiskAssessmentResult,
  RiskRule,
  CreateRuleRequest,
  UpdateRuleRequest,
  BlocklistEntry,
  AddBlocklistEntryRequest,

  // Webhooks
  CreateWebhookEndpointRequest,
  WebhookEndpoint,
  WebhookDelivery,
  WebhookEvent,
  TraceContext,

  // Developer
  CreateApplicationRequest,
  Application,

  // Settings
  MerchantSettings,
  UpdateSettingsRequest,

  // OIDC
  OAuthClient,
  AuthorizeRequest,
  TokenRequest,
  TokenResponse,

  // Notify
  Template,
  NotificationLog,

  // Chat
  Channel,
  Message,

  // Storage
  UploadUrlResponse,
  FileEntry,

  // Flags
  Flag,
  EvaluateResponse,

  // Logs
  AuditLogEntry,
  LogsPage,

  // Subscriptions
  Plan,
  Subscription,

  // Invoices
  LineItem,
  Invoice,

  // Payment Links
  PaymentLink,

  // Guard
  RateLimitConfig,
  IpBlocklistEntry,

  // IAM
  Permission,
  Role,
  Group,
  Policy,
  ApiKey,
  AuditEvent,
  Tenant,
  TenantQuota,
  AppScope,

  // Billing
  BillingPlan,
  BillingSubscription,
  BillingPaymentMethod,
  BillingInvoice,
  CreditBalance,

  // Analytics
  PaymentMetrics,
  TimeSeriesPoint,
  VolumeTimeSeries,
  TransactionCountSeries,
  PaymentMethodShare,
  PaymentMethodBreakdown,
  GeoEntry,
  GeoDistribution,
  FailureReason,
  FailureAnalysis,
  CustomerMetrics,
  CohortRow,
  CohortAnalysis,
  CustomerSegment,
  TopCustomer,
} from "./types.js";

// Chain types (re-exported from resource file)
export type {
  CryptoWallet,
  WalletAddress,
  AssetBalance,
  CryptoPayment,
  SwapResult,
  Escrow,
  OnRampResult,
  OffRampResult,
  ScreenResult,
  GasEstimate,
  ExchangeRate,
  BatchPayout,
} from "./resources/chain.js";

// Local Payments types
export type {
  LocalPayment,
  OtpConfirmResult,
  RedirectPaymentResult,
  LocalRefundResult,
  ProviderConfig,
} from "./resources/localPayments.js";

// IAM types (request/response shapes defined in resource file)
export type {
  UpsertRoleRequest,
  UpsertGroupRequest,
  UpsertPolicyRequest,
  BindPolicyRequest,
  CreateApiKeyRequest,
  CreateApiKeyResponse,
  RotateApiKeyResponse,
  EvaluateAccessRequest,
  EvaluateAccessResponse,
  ListAuditEventsOptions,
  ListAuditEventsResponse,
  RegisterAppScopeRequest,
  GrantScopeToRoleRequest,
  GetEffectiveScopesRequest,
  CreateTenantRequest,
  CreateTenantResponse,
  TenantProvisioningStatus,
  ListTenantsOptions,
  ListTenantsResponse,
  UpdateTenantRequest,
  SuspendTenantResponse,
  BulkTenantActionRequest,
  BulkTenantActionResponse,
} from "./resources/iam.js";

// Metering types
export type {
  UsageResource,
  QuotaEnforcement,
  ResourceUsage,
  UsageSummary,
  UsageTimeSeriesPoint,
  UsageTimeSeriesResponse,
  ResourceQuota,
  QuotaStatus,
} from "./resources/metering.js";

// Billing types (request/response shapes defined in resource file)
export type {
  ChangePlanRequest,
  CancelPlanRequest,
  PreviewPlanChangeRequest,
  PreviewPlanChangeResponse,
  AddBillingPaymentMethodRequest,
  RedeemPromoCodeRequest,
  RedeemPromoCodeResponse,
} from "./resources/billing.js";

// Analytics types (query shapes defined in resource file)
export type {
  AnalyticsPeriod,
  Granularity,
  AnalyticsQuery,
  TimeSeriesQuery,
  TopCustomersQuery,
} from "./resources/analytics.js";
