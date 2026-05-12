package mashgate

// ────────────────────────────────────────────────────────────────────────────
// TODO: Resource wrappers — present in TypeScript SDK, MISSING in Go SDK.
//
// Generated types for all of these already exist in _generated/types.gen.go
// (compiled from contracts/proto/v1/*.proto via oapi-codegen). What's missing
// is the idiomatic Go method surface — same pattern as billing.go.
//
// Add resources here in priority order. Bump sdk_versions.go on PR.
// ────────────────────────────────────────────────────────────────────────────

// --- analytics (contracts/proto/v1/analytics.proto · AnalyticsService · 12 RPCs)
// TODO(P2): ListMetrics, GetMetric, QueryTimeSeries, ListReports, GetReport,
// CreateReport, RunReport, ExportReport, ListDashboards, …
// Consumers: ops/observability dashboards (currently bypass to ClickHouse direct).

// --- chain (contracts/proto/v1/chain.proto · ChainService · 21 RPCs)
// TODO(P2): GetBalance, ListTransactions, EstimateFee, BroadcastTx, GetBlock,
// GetTransaction, ListAddresses, CreateAddress, …
// Consumers: any vertical needing crypto rails (qrapp /pay-with-crypto, future).

// --- developer (contracts/proto/v1/developer.proto · DeveloperService · 8 RPCs)
// TODO(P3): ListAPIKeys, CreateAPIKey, RevokeAPIKey, ListWebhookEndpoints, …
// Consumers: self-service portal (tenant admin UI).

// --- local_payments (contracts/proto/v1/local_payments.proto · LocalPaymentsService · 9 RPCs)
// TODO(P1 once provider wired): InitiatePayment, ConfirmPayment, GetPaymentStatus,
// ListSupportedMethods (TJ-specific: Tcell, Korti Milli, Alif, …), …
// ADR-0009 governs.

// --- metering (contracts/proto/v1/metering.proto · MeteringService · 3 RPCs)
// TODO(P2): RecordUsage, ListUsage, GetUsageSummary
// Consumers: any service that charges per-usage (storage GB, API calls, etc.).

// --- risk (contracts/proto/v1/risk.proto · RiskService · 9 RPCs)
// TODO(P2): AssessTransaction, ListAssessments, GetAssessment, AddBlocklistEntry,
// ListBlocklistEntries, RemoveBlocklistEntry, …
// Consumers: payments-orchestrator (internal), high-risk vertical fraud screens.

// --- settings (no proto — backed by openapi only?)
// TODO(P3): GetSettings, UpdateSettings — tenant-level config blob.
// Verify proto location before implementing.

// --- wallet_admin (in wallet.proto under admin/* RPCs)
// TODO(P3): AdminListWallets, AdminFreezeWallet, AdminUnfreezeWallet, AdminAdjustBalance, …
// Privileged. RBAC-gated by IamService (auth-service).
