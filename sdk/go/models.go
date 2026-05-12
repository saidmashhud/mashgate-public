package mashgate

import (
	"encoding/json"
	"time"
)

// ────────────────────────────────────────────────────────────────────────────
// Money
// ────────────────────────────────────────────────────────────────────────────

// Money represents a monetary amount as a decimal string to avoid
// floating-point precision issues.
//
//	mashgate.Money{Amount: "150000.00", Currency: "UZS"}
type Money struct {
	Amount   string `json:"amount"`   // decimal string, e.g. "150000.00"
	Currency string `json:"currency"` // ISO 4217: "UZS", "KZT", "USD", etc.
}

// ────────────────────────────────────────────────────────────────────────────
// Payment
// ────────────────────────────────────────────────────────────────────────────

// Payment represents a Mashgate payment intent.
type Payment struct {
	PaymentID   string `json:"paymentId"`
	TenantID    string `json:"tenantId"`
	Status      string `json:"status"` // pending, authorized, captured, refunded, voided, failed
	Amount      Money  `json:"amount"`
	OrderID     string `json:"orderId"`
	CaptureMode string `json:"captureMode"` // "AUTO" | "MANUAL"
	CreatedAt   int64  `json:"createdAt"`
	UpdatedAt   int64  `json:"updatedAt"`
}

// ────────────────────────────────────────────────────────────────────────────
// Checkout
// ────────────────────────────────────────────────────────────────────────────

// CheckoutSession represents a hosted checkout session.
type CheckoutSession struct {
	SessionID     string `json:"sessionId"`
	Status        string `json:"status"` // pending, completed, expired, cancelled
	TotalAmount   Money  `json:"totalAmount"`
	CheckoutURL   string `json:"checkoutUrl"`
	SuccessURL    string `json:"successUrl"`
	CancelURL     string `json:"cancelUrl"`
	CustomerEmail string `json:"customerEmail"`
	ExpiresAt     int64  `json:"expiresAt"`
	CreatedAt     int64  `json:"createdAt"`
}

// LineItem is a single product or service in a checkout session.
type LineItem struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Quantity    int    `json:"quantity"`
	UnitPrice   Money  `json:"unitPrice"`
}

// ────────────────────────────────────────────────────────────────────────────
// Webhooks
// ────────────────────────────────────────────────────────────────────────────

// WebhookEndpoint is a registered destination for Mashgate event notifications.
type WebhookEndpoint struct {
	EndpointID    string   `json:"endpointId"`
	URL           string   `json:"url"`
	Description   string   `json:"description"`
	EventTypes    []string `json:"events"`
	Status        string   `json:"status"`                  // "active" | "disabled"
	SigningSecret string   `json:"signingSecret,omitempty"` // only returned on Create
	CreatedAt     int64    `json:"createdAt"`
}

// WebhookEvent is the envelope that Mashgate POSTs to your endpoint.
//
// Envelope v1 (ADR-0013 §4, contracts/events/_envelope.v1.json) introduces
// these canonical fields: ID, Topic, OccurredAt, CreatedAt, Payload, Source,
// plus optional Trace. Legacy emitters (pre-envelope) populate EventID,
// EventType, EventVersion, CorrelationID, AggregateID, and Data. This struct
// keeps both sets for compatibility — new consumers should prefer ID/Topic/
// Payload; PayloadBytes() picks the right one transparently.
type WebhookEvent struct {
	// Envelope v1 canonical fields.
	ID        string          `json:"id,omitempty"`
	Topic     string          `json:"topic,omitempty"`
	CreatedAt int64           `json:"created_at,omitempty"`
	Payload   json.RawMessage `json:"payload,omitempty"`
	Source    string          `json:"source,omitempty"`
	Trace     *TraceContext   `json:"_trace,omitempty"`

	// Fields present in both legacy and envelope v1 emissions.
	EventID      string `json:"event_id"`
	TenantID     string `json:"tenant_id"`
	OccurredAt   int64  `json:"occurred_at"`
	EventType    string `json:"event_type,omitempty"`
	EventVersion int    `json:"event_version,omitempty"`

	// Legacy fields (pre-envelope; kept for back-compat).
	CorrelationID string `json:"correlation_id,omitempty"`
	AggregateID   string `json:"aggregate_id,omitempty"`
	// Data is the pre-envelope flat payload. Deprecated: use Payload.
	Data json.RawMessage `json:"data,omitempty"`
}

// TraceContext is the envelope v1 `_trace` object (W3C Trace Context).
type TraceContext struct {
	Traceparent string `json:"traceparent"`
	Tracestate  string `json:"tracestate,omitempty"`
}

// PayloadBytes returns the event payload regardless of whether the producer
// emitted envelope v1 (Payload) or the legacy flat format (Data).
func (e *WebhookEvent) PayloadBytes() json.RawMessage {
	if len(e.Payload) > 0 {
		return e.Payload
	}
	return e.Data
}

// CanonicalID returns the event id preferring envelope v1 `id` and falling
// back to the legacy `event_id`.
func (e *WebhookEvent) CanonicalID() string {
	if e.ID != "" {
		return e.ID
	}
	return e.EventID
}

// Event type constants — use these instead of raw strings.
//
// Legacy constants keep the pre-envelope dotted form (e.g. "payment.created")
// that the pre-envelope emitter produces. Envelope v1 topics use the longer
// "<product>.<resource>.<verb>" form (e.g. "payments.payment.created") — see
// TopicPaymentCreated below. Both forms are emitted depending on producer
// configuration during migration; consumers should tolerate either.
const (
	EventPaymentCreated             = "payment.created"
	EventPaymentAuthorized          = "payment.authorized"
	EventPaymentAuthorizationFailed = "payment.authorization_failed"
	EventPaymentCaptured            = "payment.captured"
	EventPaymentCaptureFailed       = "payment.capture_failed"
	EventPaymentVoided              = "payment.voided"
	EventPaymentFailed              = "payment.failed"
	EventRefundRequested            = "refund.requested"
	EventRefundSettled              = "refund.settled"
	EventRefundFailed               = "refund.failed"
	EventCheckoutCompleted          = "checkout.completed"
	EventCheckoutExpired            = "checkout.expired"
)

// Envelope v1 topic constants (ADR-0013 §4 `<product>.<resource>.<verb>`).
// Prefer these over the legacy Event* constants when emitting new integrations.
const (
	TopicPaymentCreated           = "payments.payment.created"
	TopicPaymentCompleted         = "payments.payment.completed"
	TopicPaymentFailed            = "payments.payment.failed"
	TopicPaymentAuthorized        = "payments.payment.authorized"
	TopicPaymentVoided            = "payments.payment.voided"
	TopicRefundCreated            = "payments.refund.created"
	TopicRefundCompleted          = "payments.refund.completed"
	TopicRefundFailed             = "payments.refund.failed"
	TopicCheckoutSessionCreated   = "payments.checkout_session.created"
	TopicCheckoutSessionCompleted = "payments.checkout_session.completed"
	TopicUserRegistered           = "iam.user.registered"
	TopicNotificationSent         = "notifications.notification.sent"
)

// ────────────────────────────────────────────────────────────────────────────
// Auth
// ────────────────────────────────────────────────────────────────────────────

// LoginRequest holds credentials for the POST /v1/auth/login endpoint.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// TokenPair is returned by Login and RefreshToken.
//
// User and ExpiresAt are populated when the auth-service includes them
// (Login + chained Register flows). RefreshToken/Logout responses may
// not carry User; check for nil.
//
// ExpiresAt is unix-seconds. Mashgate currently emits it as a string —
// json.Number transparently handles both number и string forms.
type TokenPair struct {
	AccessToken  string        `json:"accessToken"`
	RefreshToken string        `json:"refreshToken"`
	ExpiresAt    json.Number   `json:"expiresAt,omitempty"`
	User         *AuthUserInfo `json:"user,omitempty"`
}

// ExpiresAtUnix returns ExpiresAt parsed as int64 unix seconds, or 0 если empty.
func (t *TokenPair) ExpiresAtUnix() int64 {
	if t == nil || t.ExpiresAt == "" {
		return 0
	}
	v, err := t.ExpiresAt.Int64()
	if err != nil {
		return 0
	}
	return v
}

// AuthUserInfo carries identity claims surfaced by Mashgate auth-service
// alongside a TokenPair (when the upstream response includes a user object).
type AuthUserInfo struct {
	UserID   string   `json:"userId"`
	Email    string   `json:"email"`
	FullName string   `json:"fullName,omitempty"`
	TenantID string   `json:"tenantId,omitempty"`
	Role     string   `json:"role,omitempty"`
	Roles    []string `json:"roles,omitempty"`
}

// RegisterRequest creates a new user via POST /v1/auth/register.
//
// SECURITY NOTE: Mashgate auth-service /v1/auth/register currently accepts
// `Role` без sanitization — кто угодно может self-register с role=platform_admin.
// Это известная проблема upstream (issue открыт), будет закрыто валидацией
// whitelist'а ролей. Используйте Role="merchant" по умолчанию для merchant
// flows; admin-grade роли пусть выдаёт админ через AssignRole.
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name,omitempty"`
	TenantID string `json:"tenant_id,omitempty"`
	// Role is one of: "merchant" | "admin" | "viewer" | "platform_admin".
	// Empty string defaults to "admin" upstream.
	Role string `json:"role,omitempty"`
}

// RegisterResponse carries the new user record (no tokens — caller must
// chain Login to obtain a session).
//
// CreatedAt is unix-seconds, encoded as string upstream — exposed as
// json.Number for transparent parsing.
type RegisterResponse struct {
	UserID    string      `json:"userId"`
	Email     string      `json:"email"`
	TenantID  string      `json:"tenantId"`
	CreatedAt json.Number `json:"createdAt"`
}

// CreatedAtUnix returns CreatedAt parsed as int64 unix seconds.
func (r *RegisterResponse) CreatedAtUnix() int64 {
	if r == nil || r.CreatedAt == "" {
		return 0
	}
	v, err := r.CreatedAt.Int64()
	if err != nil {
		return 0
	}
	return v
}

// SendOtpRequest triggers an OTP delivery via Mashgate notify-service.
//
// Exactly one of UserID or Phone must be set. Purpose ∈
// {"login" | "password_reset" | "phone_verify"}.
type SendOtpRequest struct {
	UserID  string `json:"user_id,omitempty"`
	Phone   string `json:"phone,omitempty"`
	Purpose string `json:"purpose"`
}

// VerifyOtpRequest checks an OTP previously sent. Exactly one of UserID
// or Phone must be set — match the way SendOtpRequest was called.
type VerifyOtpRequest struct {
	UserID  string `json:"user_id,omitempty"`
	Phone   string `json:"phone,omitempty"`
	Code    string `json:"code"`
	Purpose string `json:"purpose"`
}

// ────────────────────────────────────────────────────────────────────────────
// IAM — API Keys
// ────────────────────────────────────────────────────────────────────────────

// APIKey represents a programmatic access key (for machine-to-machine auth).
type APIKey struct {
	KeyID       string   `json:"keyId"`
	Name        string   `json:"name"`
	TenantID    string   `json:"tenantId"`
	Prefix      string   `json:"prefix"` // first 8 chars of the key, for display
	Permissions []string `json:"permissions"`
	CreatedAt   int64    `json:"createdAt"`
	ExpiresAt   int64    `json:"expiresAt,omitempty"` // 0 = never
}

// APIKeyCreated is returned on CreateAPIKey — includes the full secret (shown once).
type APIKeyCreated struct {
	APIKey
	Secret string `json:"secret"` // full key value, store securely
}

// CreateAPIKeyRequest describes a new API key to create.
type CreateAPIKeyRequest struct {
	Name        string   `json:"name"`
	Permissions []string `json:"permissions,omitempty"`
	ExpiresAt   int64    `json:"expiresAt,omitempty"` // Unix timestamp; 0 = no expiry
}

// ────────────────────────────────────────────────────────────────────────────
// Events — Endpoints & Deliveries
// ────────────────────────────────────────────────────────────────────────────

// Endpoint is a registered webhook destination managed via the Events SDK.
type Endpoint struct {
	ID            string   `json:"id"`
	URL           string   `json:"url"`
	Description   string   `json:"description,omitempty"`
	EventTypes    []string `json:"eventTypes"`
	Status        string   `json:"status"`                  // "active" | "disabled"
	SigningSecret string   `json:"signingSecret,omitempty"` // only on Create / RotateSecret
	CreatedAt     int64    `json:"createdAt"`
	UpdatedAt     int64    `json:"updatedAt"`
}

// Delivery represents a single webhook delivery attempt.
type Delivery struct {
	ID             string `json:"id"`
	EndpointID     string `json:"endpointId"`
	EventID        string `json:"eventId"`
	Status         string `json:"status"` // "pending" | "succeeded" | "failed"
	AttemptCount   int    `json:"attemptCount"`
	ResponseStatus int    `json:"responseStatus"` // HTTP status returned by the endpoint
	NextRetryAt    int64  `json:"nextRetryAt,omitempty"`
	CreatedAt      int64  `json:"createdAt"`
	UpdatedAt      int64  `json:"updatedAt"`
}

// CreateEndpointRequest is the payload for creating a new webhook endpoint.
type CreateEndpointRequest struct {
	URL         string   `json:"url"`
	Description string   `json:"description,omitempty"`
	EventTypes  []string `json:"eventTypes"`
}

// UpdateEndpointRequest is the payload for updating an existing endpoint.
type UpdateEndpointRequest struct {
	URL         string   `json:"url,omitempty"`
	Description string   `json:"description,omitempty"`
	EventTypes  []string `json:"eventTypes,omitempty"`
	Status      string   `json:"status,omitempty"` // "active" | "disabled"
}

// WebhookSubscription binds an endpoint to a set of event types.
type WebhookSubscription struct {
	ID         string   `json:"id"`
	EndpointID string   `json:"endpointId"`
	TenantID   string   `json:"tenantId"`
	EventTypes []string `json:"eventTypes"`
	Status     string   `json:"status"` // "active" | "paused"
	CreatedAt  int64    `json:"createdAt"`
}

// DLQEntry is a failed delivery in the dead-letter queue.
type DLQEntry struct {
	DeliveryID   string `json:"deliveryId"`
	EndpointID   string `json:"endpointId"`
	EventID      string `json:"eventId"`
	EventType    string `json:"eventType"`
	Payload      string `json:"payload"`
	AttemptCount int    `json:"attemptCount"`
	LastError    string `json:"lastError"`
	FailedAt     int64  `json:"failedAt"`
}

// ────────────────────────────────────────────────────────────────────────────
// Payment methods
// ────────────────────────────────────────────────────────────────────────────

// CardPaymentMethod describes a tokenized card for use in a payment.
type CardPaymentMethod struct {
	Token     string `json:"token"`
	Brand     string `json:"brand"` // "uzcard" | "humo" | "visa" | "mastercard" etc.
	Last4     string `json:"last4"`
	BIN       string `json:"bin"`
	LuhnValid bool   `json:"luhnValid"`
}

// WalletPaymentMethod describes an e-wallet for use in checkout (redirect flow).
type WalletPaymentMethod struct {
	Provider string `json:"provider"` // "click" | "payme" | "oson"
	Phone    string `json:"phone"`    // Uzbekistan phone: +998XXXXXXXXX
}

// ────────────────────────────────────────────────────────────────────────────
// Billing types — mirror BillingService schema from contracts/proto/v1/billing.proto.
// Append to models.go on next regen.
// ────────────────────────────────────────────────────────────────────────────

// BillingPlan describes a platform subscription plan available to tenants.
type BillingPlan struct {
	ID            string            `json:"id"`
	Code          string            `json:"code"`
	Name          string            `json:"name"`
	Description   string            `json:"description,omitempty"`
	PriceCents    int64             `json:"priceCents"`
	Currency      string            `json:"currency"`
	BillingPeriod string            `json:"billingPeriod"` // "monthly" | "yearly"
	Tier          string            `json:"tier,omitempty"`
	Features      []string          `json:"features,omitempty"`
	Limits        map[string]int64  `json:"limits,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	IsActive      bool              `json:"isActive"`
}

// BillingSubscription is the tenant's active subscription state.
type BillingSubscription struct {
	ID                 string            `json:"id"`
	TenantID           string            `json:"tenantId"`
	PlanID             string            `json:"planId"`
	Status             string            `json:"status"` // active | past_due | canceled | trialing
	CurrentPeriodStart time.Time         `json:"currentPeriodStart"`
	CurrentPeriodEnd   time.Time         `json:"currentPeriodEnd"`
	CancelAtPeriodEnd  bool              `json:"cancelAtPeriodEnd"`
	CanceledAt         *time.Time        `json:"canceledAt,omitempty"`
	TrialEnd           *time.Time        `json:"trialEnd,omitempty"`
	Metadata           map[string]string `json:"metadata,omitempty"`
}

// ChangePlanRequest switches the active subscription to a new plan.
// Prorate=true (default) → prorated charge for remaining period.
// Effective="immediate" | "period_end".
type ChangePlanRequest struct {
	NewPlanID string `json:"newPlanId"`
	Prorate   bool   `json:"prorate"`
	Effective string `json:"effective,omitempty"`
}

// CancelPlanRequest cancels the active subscription.
type CancelPlanRequest struct {
	Reason            string `json:"reason,omitempty"`
	CancelImmediately bool   `json:"cancelImmediately,omitempty"`
}

// PreviewPlanChangeRequest computes hypothetical proration for a plan switch.
type PreviewPlanChangeRequest struct {
	NewPlanID string `json:"newPlanId"`
	Effective string `json:"effective,omitempty"`
}

// PreviewPlanChangeResponse holds proration details.
type PreviewPlanChangeResponse struct {
	NewPlan          *BillingPlan `json:"newPlan"`
	OldPlan          *BillingPlan `json:"oldPlan,omitempty"`
	ProrationCents   int64        `json:"prorationCents"`
	CreditCents      int64        `json:"creditCents"`
	NextChargeCents  int64        `json:"nextChargeCents"`
	NextChargeAt     time.Time    `json:"nextChargeAt"`
	EffectiveAt      time.Time    `json:"effectiveAt"`
}

// BillingPaymentMethod represents a payment method on file for billing.
type BillingPaymentMethod struct {
	ID         string            `json:"id"`
	Type       string            `json:"type"` // "card" | "wallet" | "bank"
	IsDefault  bool              `json:"isDefault"`
	CreatedAt  time.Time         `json:"createdAt"`
	Card       *CardPaymentMethod   `json:"card,omitempty"`
	Wallet     *WalletPaymentMethod `json:"wallet,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// AddBillingPaymentMethodRequest registers a new billing payment method.
// Set Card OR Wallet (not both).
type AddBillingPaymentMethodRequest struct {
	Type      string               `json:"type"`
	Card      *CardPaymentMethod   `json:"card,omitempty"`
	Wallet    *WalletPaymentMethod `json:"wallet,omitempty"`
	IsDefault bool                 `json:"isDefault,omitempty"`
}

// BillingInvoice is an issued invoice for a tenant.
type BillingInvoice struct {
	ID            string            `json:"id"`
	TenantID      string            `json:"tenantId"`
	Number        string            `json:"number"`
	Status        string            `json:"status"` // draft | open | paid | void | uncollectible
	AmountCents   int64             `json:"amountCents"`
	Currency      string            `json:"currency"`
	IssuedAt      time.Time         `json:"issuedAt"`
	DueAt         time.Time         `json:"dueAt"`
	PaidAt        *time.Time        `json:"paidAt,omitempty"`
	HostedPageURL string            `json:"hostedPageUrl,omitempty"`
	PDFURL        string            `json:"pdfUrl,omitempty"`
	Lines         []BillingInvoiceLine `json:"lines,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

// BillingInvoiceLine is one line item on an invoice.
type BillingInvoiceLine struct {
	Description string `json:"description"`
	Quantity    int64  `json:"quantity"`
	UnitCents   int64  `json:"unitCents"`
	AmountCents int64  `json:"amountCents"`
	PeriodStart *time.Time `json:"periodStart,omitempty"`
	PeriodEnd   *time.Time `json:"periodEnd,omitempty"`
}

// CreditBalance is the tenant's available credit (applied to future invoices).
type CreditBalance struct {
	AmountCents int64  `json:"amountCents"`
	Currency    string `json:"currency"`
}

// RedeemPromoCodeResponse is returned after applying a promo code.
type RedeemPromoCodeResponse struct {
	Applied      bool          `json:"applied"`
	CreditCents  int64         `json:"creditCents,omitempty"`
	Currency     string        `json:"currency,omitempty"`
	ValidUntil   *time.Time    `json:"validUntil,omitempty"`
	Balance      *CreditBalance `json:"balance,omitempty"`
	ErrorMessage string        `json:"errorMessage,omitempty"`
}

// ────────────────────────────────────────────────────────────────────────────
// Analytics (analytics.proto)
// ────────────────────────────────────────────────────────────────────────────

type PaymentMetrics struct {
	TotalVolumeCents int64   `json:"totalVolumeCents"`
	Currency         string  `json:"currency"`
	TransactionCount int64   `json:"transactionCount"`
	AvgTicketCents   int64   `json:"avgTicketCents"`
	SuccessRate      float64 `json:"successRate"`
	RefundRate       float64 `json:"refundRate"`
	ChargebackRate   float64 `json:"chargebackRate,omitempty"`
}

type TimeSeriesPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

type VolumeTimeSeries struct {
	Points   []TimeSeriesPoint `json:"points"`
	Currency string            `json:"currency"`
}

type TransactionCountSeries struct {
	Points []TimeSeriesPoint `json:"points"`
}

type PaymentMethodBreakdown struct {
	Methods []struct {
		Method      string  `json:"method"` // card | wallet | bank | local
		ShareVolume float64 `json:"shareVolume"`
		ShareCount  float64 `json:"shareCount"`
	} `json:"methods"`
}

type GeoDistribution struct {
	Countries []struct {
		CountryCode      string  `json:"countryCode"`
		CountryName      string  `json:"countryName"`
		VolumeCents      int64   `json:"volumeCents"`
		Currency         string  `json:"currency"`
		TransactionCount int64   `json:"transactionCount"`
		Share            float64 `json:"share"`
	} `json:"countries"`
}

type FailureAnalysis struct {
	TotalFailed int64 `json:"totalFailed"`
	Reasons     []struct {
		Code        string `json:"code"`
		Description string `json:"description"`
		Count       int64  `json:"count"`
	} `json:"reasons"`
}

type CustomerMetrics struct {
	NewCustomers     int64   `json:"newCustomers"`
	ReturningCustomers int64 `json:"returningCustomers"`
	ChurnedCustomers int64   `json:"churnedCustomers"`
	AvgLifetimeValue int64   `json:"avgLifetimeValueCents"`
}

type CohortAnalysis struct {
	Cohorts []struct {
		CohortLabel  string    `json:"cohortLabel"`
		CreatedAt    time.Time `json:"createdAt"`
		Size         int64     `json:"size"`
		RetentionPct []float64 `json:"retentionPct"`
	} `json:"cohorts"`
}

type CustomerSegment struct {
	Segment string `json:"segment"` // VIP | casual | new | churned
	Count   int64  `json:"count"`
	Share   float64 `json:"share"`
}

type TopCustomer struct {
	CustomerID       string `json:"customerId"`
	Email            string `json:"email,omitempty"`
	LifetimeValue    int64  `json:"lifetimeValueCents"`
	Currency         string `json:"currency"`
	TransactionCount int64  `json:"transactionCount"`
}

// ────────────────────────────────────────────────────────────────────────────
// Chain (chain.proto)
// ────────────────────────────────────────────────────────────────────────────

type CreateAddressRequest struct {
	TenantID string `json:"tenantId"`
	Network  string `json:"network"` // e.g. "solana_mainnet", "ethereum_sepolia"
	Label    string `json:"label,omitempty"`
}

type ChainAddress struct {
	AddressID    string    `json:"addressId"`
	TenantID     string    `json:"tenantId"`
	Network      string    `json:"network"`
	Address      string    `json:"address"`
	Label        string    `json:"label,omitempty"`
	DerivationPath string  `json:"derivationPath,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
}

type ChainBalance struct {
	AddressID string `json:"addressId"`
	Network   string `json:"network"`
	Balance   string `json:"balance"` // decimal as string (avoids float precision)
	Asset     string `json:"asset"`   // e.g. "SOL", "USDC"
}

type ChainTransaction struct {
	TxID        string    `json:"txId"`
	Hash        string    `json:"hash"`
	Network     string    `json:"network"`
	From        string    `json:"from"`
	To          string    `json:"to"`
	Amount      string    `json:"amount"`
	Asset       string    `json:"asset"`
	Status      string    `json:"status"` // pending | confirmed | failed
	BlockHeight int64     `json:"blockHeight,omitempty"`
	Confirmations int     `json:"confirmations,omitempty"`
	FeeAmount   string    `json:"feeAmount,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

type EstimateFeeRequest struct {
	TenantID string `json:"tenantId"`
	Network  string `json:"network"`
	From     string `json:"from"`
	To       string `json:"to"`
	Amount   string `json:"amount"`
	Asset    string `json:"asset"`
}

type FeeEstimate struct {
	Amount   string `json:"amount"`
	Asset    string `json:"asset"`
	Priority string `json:"priority"` // low | medium | high
}

type SendTransactionRequest struct {
	TenantID       string `json:"tenantId"`
	Network        string `json:"network"`
	FromAddressID  string `json:"fromAddressId"`
	To             string `json:"to"`
	Amount         string `json:"amount"`
	Asset          string `json:"asset"`
	IdempotencyKey string `json:"-"`
}

type ChainBlock struct {
	Network   string    `json:"network"`
	Hash      string    `json:"hash"`
	Height    int64     `json:"height"`
	Timestamp time.Time `json:"timestamp"`
	TxCount   int       `json:"txCount"`
}

type ChainNetwork struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Testnet  bool   `json:"testnet"`
	Asset    string `json:"asset"`
	Decimals int    `json:"decimals"`
}

// ────────────────────────────────────────────────────────────────────────────
// Developer (developer.proto)
// ────────────────────────────────────────────────────────────────────────────

type DeveloperActivity struct {
	APICalls24h     int64 `json:"apiCalls24h"`
	WebhookDeliveries24h int64 `json:"webhookDeliveries24h"`
	WebhookFailures24h   int64 `json:"webhookFailures24h"`
	LastAPIError    *struct {
		Code        string    `json:"code"`
		Message     string    `json:"message"`
		Timestamp   time.Time `json:"timestamp"`
	} `json:"lastApiError,omitempty"`
}

type IntegrationHealth struct {
	APIKeysActive       int     `json:"apiKeysActive"`
	WebhookEndpoints    int     `json:"webhookEndpoints"`
	WebhookSuccessRate  float64 `json:"webhookSuccessRate"`
	LastWebhookError    string  `json:"lastWebhookError,omitempty"`
	OverallStatus       string  `json:"overallStatus"` // healthy | degraded | down
}

// ────────────────────────────────────────────────────────────────────────────
// Local Payments (local_payments.proto)
// ────────────────────────────────────────────────────────────────────────────

type LocalPaymentMethod struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Country     string `json:"country"`
	Currency    string `json:"currency"`
	LogoURL     string `json:"logoUrl,omitempty"`
	Description string `json:"description,omitempty"`
	Active      bool   `json:"active"`
}

type InitiateLocalPaymentRequest struct {
	TenantID       string            `json:"tenantId"`
	MethodID       string            `json:"methodId"`
	Amount         Money             `json:"amount"`
	CustomerPhone  string            `json:"customerPhone,omitempty"`
	CustomerEmail  string            `json:"customerEmail,omitempty"`
	OrderID        string            `json:"orderId"`
	ReturnURL      string            `json:"returnUrl,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
	IdempotencyKey string            `json:"-"`
}

type LocalPaymentInitiated struct {
	PaymentID  string `json:"paymentId"`
	Status     string `json:"status"` // pending_confirmation | pending_redirect | pending_ussd
	NextStep   struct {
		Type      string `json:"type"` // redirect | ussd | qr | otp
		Value     string `json:"value"`
		ExpiresAt time.Time `json:"expiresAt,omitempty"`
	} `json:"nextStep"`
}

type LocalPayment struct {
	PaymentID    string    `json:"paymentId"`
	TenantID     string    `json:"tenantId"`
	MethodID     string    `json:"methodId"`
	Amount       Money     `json:"amount"`
	Status       string    `json:"status"` // pending | succeeded | failed | cancelled
	OrderID      string    `json:"orderId"`
	ProviderRef  string    `json:"providerRef,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
	CompletedAt  *time.Time `json:"completedAt,omitempty"`
	FailureReason string   `json:"failureReason,omitempty"`
}

type ConfirmLocalPaymentRequest struct {
	OTP        string `json:"otp,omitempty"`
	UssdCode   string `json:"ussdCode,omitempty"`
	ProviderRef string `json:"providerRef,omitempty"`
}

// ────────────────────────────────────────────────────────────────────────────
// Metering (metering.proto)
// ────────────────────────────────────────────────────────────────────────────

type RecordUsageRequest struct {
	TenantID       string            `json:"tenantId"`
	MeterCode      string            `json:"meterCode"`
	Quantity       float64           `json:"quantity"`
	OccurredAt     time.Time         `json:"occurredAt,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
	IdempotencyKey string            `json:"-"`
}

type UsageRecord struct {
	RecordID   string            `json:"recordId"`
	TenantID   string            `json:"tenantId"`
	MeterCode  string            `json:"meterCode"`
	Quantity   float64           `json:"quantity"`
	OccurredAt time.Time         `json:"occurredAt"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

type ListUsageParams struct {
	TenantID  string
	MeterCode string
	From      time.Time
	To        time.Time
	PageSize  int
}

type UsageSummary struct {
	TenantID string `json:"tenantId"`
	From     time.Time `json:"from"`
	To       time.Time `json:"to"`
	Meters   []struct {
		MeterCode    string  `json:"meterCode"`
		TotalQty     float64 `json:"totalQuantity"`
		Unit         string  `json:"unit,omitempty"`
		CostCents    int64   `json:"costCents,omitempty"`
		Currency     string  `json:"currency,omitempty"`
	} `json:"meters"`
}

// ────────────────────────────────────────────────────────────────────────────
// Risk (risk.proto)
// ────────────────────────────────────────────────────────────────────────────

type AssessTransactionRequest struct {
	TenantID      string            `json:"tenantId"`
	OrderID       string            `json:"orderId,omitempty"`
	Amount        Money             `json:"amount"`
	CustomerEmail string            `json:"customerEmail,omitempty"`
	CustomerPhone string            `json:"customerPhone,omitempty"`
	CustomerIP    string            `json:"customerIp,omitempty"`
	BIN           string            `json:"bin,omitempty"`
	Country       string            `json:"country,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

type RiskAssessment struct {
	AssessmentID       string            `json:"assessmentId"`
	TenantID           string            `json:"tenantId"`
	Score              int               `json:"score"` // 0-100
	RiskLevel          string            `json:"riskLevel"` // low | medium | high
	RecommendedAction  string            `json:"recommendedAction"` // approve | review | decline
	TriggeredRules     []string          `json:"triggeredRules,omitempty"`
	Reason             string            `json:"reason,omitempty"`
	Metadata           map[string]string `json:"metadata,omitempty"`
	CreatedAt          time.Time         `json:"createdAt"`
}

type AddBlocklistEntryRequest struct {
	TenantID  string    `json:"tenantId"`
	EntryType string    `json:"entryType"` // email | phone | card_bin | ip | country
	Value     string    `json:"value"`
	Reason    string    `json:"reason,omitempty"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
}

type BlocklistEntry struct {
	EntryID   string     `json:"entryId"`
	TenantID  string     `json:"tenantId"`
	EntryType string     `json:"entryType"`
	Value     string     `json:"value"`
	Reason    string     `json:"reason,omitempty"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
}

type RiskRule struct {
	RuleID      string `json:"ruleId"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Active      bool   `json:"active"`
	Severity    string `json:"severity"` // low | medium | high
}

type RiskProfile struct {
	IdentifierType  string  `json:"identifierType"`
	Identifier      string  `json:"identifier"`
	TotalAttempts   int64   `json:"totalAttempts"`
	SuccessfulPayments int64 `json:"successfulPayments"`
	FailedPayments  int64   `json:"failedPayments"`
	Chargebacks     int64   `json:"chargebacks"`
	Score           int     `json:"score"`
	RiskLevel       string  `json:"riskLevel"`
}

// ────────────────────────────────────────────────────────────────────────────
// Wallet Admin (wallet.proto admin/*)
// ────────────────────────────────────────────────────────────────────────────

type AdminWallet struct {
	WalletID         string    `json:"walletId"`
	TenantID         string    `json:"tenantId"`
	UserID           string    `json:"userId"`
	Status           string    `json:"status"` // active | frozen | closed
	BalanceCents     int64     `json:"balanceCents"`
	Currency         string    `json:"currency"`
	FrozenReason     string    `json:"frozenReason,omitempty"`
	FrozenAt         *time.Time `json:"frozenAt,omitempty"`
	KYCStatus        string    `json:"kycStatus,omitempty"`
	CreatedAt        time.Time `json:"createdAt"`
	LastActivityAt   *time.Time `json:"lastActivityAt,omitempty"`
}

type FreezeWalletRequest struct {
	Reason string `json:"reason"`
	Note   string `json:"note,omitempty"`
}

type UnfreezeWalletRequest struct {
	ResolvedReason string `json:"resolvedReason"`
	Note           string `json:"note,omitempty"`
}

type AdjustBalanceRequest struct {
	AmountCents    int64  `json:"amountCents"` // negative = debit
	Reason         string `json:"reason"`
	ReferenceID    string `json:"referenceId,omitempty"`
	IdempotencyKey string `json:"idempotencyKey,omitempty"`
}

type WalletAdjustment struct {
	AdjustmentID string    `json:"adjustmentId"`
	WalletID     string    `json:"walletId"`
	AmountCents  int64     `json:"amountCents"`
	Reason       string    `json:"reason"`
	OperatorID   string    `json:"operatorId"`
	BalanceAfter int64     `json:"balanceAfterCents"`
	CreatedAt    time.Time `json:"createdAt"`
}

type WalletAuditEntry struct {
	EntryID    string                 `json:"entryId"`
	WalletID   string                 `json:"walletId"`
	Action     string                 `json:"action"` // freeze | unfreeze | adjust | close
	OperatorID string                 `json:"operatorId"`
	Reason     string                 `json:"reason,omitempty"`
	Before     map[string]any         `json:"before,omitempty"`
	After      map[string]any         `json:"after,omitempty"`
	CreatedAt  time.Time              `json:"createdAt"`
}
