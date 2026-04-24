package mashgate

import "encoding/json"

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
	Status      string `json:"status"`      // pending, authorized, captured, refunded, voided, failed
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
	Status        string `json:"status"`     // pending, completed, expired, cancelled
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
	ID         string          `json:"id,omitempty"`
	Topic      string          `json:"topic,omitempty"`
	CreatedAt  int64           `json:"created_at,omitempty"`
	Payload    json.RawMessage `json:"payload,omitempty"`
	Source     string          `json:"source,omitempty"`
	Trace      *TraceContext   `json:"_trace,omitempty"`

	// Fields present in both legacy and envelope v1 emissions.
	EventID      string `json:"event_id"`
	TenantID     string `json:"tenant_id"`
	OccurredAt   int64  `json:"occurred_at"`
	EventType    string `json:"event_type,omitempty"`
	EventVersion int    `json:"event_version,omitempty"`

	// Legacy fields (pre-envelope; kept for back-compat).
	CorrelationID string          `json:"correlation_id,omitempty"`
	AggregateID   string          `json:"aggregate_id,omitempty"`
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
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// ────────────────────────────────────────────────────────────────────────────
// IAM — API Keys
// ────────────────────────────────────────────────────────────────────────────

// APIKey represents a programmatic access key (for machine-to-machine auth).
type APIKey struct {
	KeyID       string `json:"keyId"`
	Name        string `json:"name"`
	TenantID    string `json:"tenantId"`
	Prefix      string `json:"prefix"` // first 8 chars of the key, for display
	Permissions []string `json:"permissions"`
	CreatedAt   int64  `json:"createdAt"`
	ExpiresAt   int64  `json:"expiresAt,omitempty"` // 0 = never
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
	Status         string `json:"status"`         // "pending" | "succeeded" | "failed"
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
	ID          string   `json:"id"`
	EndpointID  string   `json:"endpointId"`
	TenantID    string   `json:"tenantId"`
	EventTypes  []string `json:"eventTypes"`
	Status      string   `json:"status"` // "active" | "paused"
	CreatedAt   int64    `json:"createdAt"`
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
	Brand     string `json:"brand"`     // "uzcard" | "humo" | "visa" | "mastercard" etc.
	Last4     string `json:"last4"`
	BIN       string `json:"bin"`
	LuhnValid bool   `json:"luhnValid"`
}

// WalletPaymentMethod describes an e-wallet for use in checkout (redirect flow).
type WalletPaymentMethod struct {
	Provider string `json:"provider"` // "click" | "payme" | "oson"
	Phone    string `json:"phone"`    // Uzbekistan phone: +998XXXXXXXXX
}
