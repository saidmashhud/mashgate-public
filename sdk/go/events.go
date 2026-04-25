package mashgate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

// ────────────────────────────────────────────────────────────────────────────
// EventsConfig
// ────────────────────────────────────────────────────────────────────────────

// EventsConfig controls how EventsClient routes calls.
//
//   - Default (mashgate mode): delegates to the Mashgate API gateway which
//     internally routes to the mg-events gRPC service. Uses the parent
//     Client's base URL and API key; optionally overridden by MashgateEventsURL.
//   - Internal hookline mode: calls the self-hosted HookLine API directly.
//     Created via InternalHooklineConfig — for Mashgate infrastructure only.
type EventsConfig struct {
	// MashgateEventsURL overrides the parent Client's baseURL for events calls
	// only (e.g. an internal mg-events REST gateway). Leave empty to use the
	// parent baseURL.
	MashgateEventsURL string

	// unexported: internal hookline mode fields
	internalMode     string // "mashgate" | "hookline"
	internalHLURL    string
	internalHLAPIKey string
}

// InternalHooklineConfig creates an EventsConfig that talks directly to
// HookLine. For Mashgate infrastructure and integration tests ONLY.
// SDK consumers should use the default mashgate mode.
func InternalHooklineConfig(hooklineURL, apiKey string) EventsConfig {
	return EventsConfig{
		internalMode:     "hookline",
		internalHLURL:    hooklineURL,
		internalHLAPIKey: apiKey,
	}
}

// ────────────────────────────────────────────────────────────────────────────
// EventsClient
// ────────────────────────────────────────────────────────────────────────────

// EventsClient manages webhook endpoints and delivery history.
// Obtain one via client.WithEvents(cfg).
type EventsClient struct {
	mode   string
	mgBase string // base URL for mashgate mode (may differ from parent)
	mgKey  string // api key for mashgate mode
	hlBase string
	hlKey  string
	http   *http.Client
}

// WithEvents attaches an EventsClient to the Mashgate client and returns a
// copy with the Events field populated.
func (c *Client) WithEvents(cfg EventsConfig) *Client {
	mode := cfg.internalMode
	if mode == "" {
		mode = "mashgate"
	}
	mgBase := cfg.MashgateEventsURL
	if mgBase == "" {
		mgBase = c.baseURL
	}
	c.Events = &EventsClient{
		mode:   mode,
		mgBase: mgBase,
		mgKey:  c.apiKey,
		hlBase: cfg.internalHLURL,
		hlKey:  cfg.internalHLAPIKey,
		http:   &http.Client{Timeout: 30 * time.Second},
	}
	return c
}

// ────────────────────────────────────────────────────────────────────────────
// Endpoint management
// ────────────────────────────────────────────────────────────────────────────

func (e *EventsClient) CreateEndpoint(ctx context.Context, req CreateEndpointRequest) (*Endpoint, error) {
	if e.mode == "hookline" {
		return e.hlCreateEndpoint(ctx, req)
	}
	body := mgCreateEndpointRequestDTO{
		URL:         req.URL,
		Description: req.Description,
		EventTypes:  req.EventTypes,
	}
	var out mgEndpointEnvelope
	if err := e.mgDo(ctx, "POST", "/v1/events/endpoints", body, &out); err != nil {
		return nil, err
	}
	ep := out.Endpoint.toPublic()
	return &ep, nil
}

func (e *EventsClient) ListEndpoints(ctx context.Context) ([]*Endpoint, error) {
	if e.mode == "hookline" {
		return e.hlListEndpoints(ctx)
	}
	var out struct {
		Endpoints []mgEndpointDTO `json:"endpoints"`
	}
	if err := e.mgDo(ctx, "GET", "/v1/events/endpoints", nil, &out); err != nil {
		return nil, err
	}
	endpoints := make([]*Endpoint, 0, len(out.Endpoints))
	for i := range out.Endpoints {
		ep := out.Endpoints[i].toPublic()
		endpoints = append(endpoints, &ep)
	}
	return endpoints, nil
}

func (e *EventsClient) UpdateEndpoint(ctx context.Context, id string, req UpdateEndpointRequest) (*Endpoint, error) {
	if e.mode == "hookline" {
		return e.hlUpdateEndpoint(ctx, id, req)
	}
	body := mgUpdateEndpointRequestDTO{
		EndpointID:  id,
		URL:         req.URL,
		Description: req.Description,
		Status:      req.Status,
		EventTypes:  req.EventTypes,
	}
	var out mgEndpointEnvelope
	if err := e.mgDo(ctx, "PUT", "/v1/events/endpoints/"+id, body, &out); err != nil {
		return nil, err
	}
	ep := out.Endpoint.toPublic()
	return &ep, nil
}

func (e *EventsClient) DeleteEndpoint(ctx context.Context, id string) error {
	if e.mode == "hookline" {
		return e.hlDeleteEndpoint(ctx, id)
	}
	return e.mgDo(ctx, "DELETE", "/v1/events/endpoints/"+id, nil, nil)
}

func (e *EventsClient) RotateSecret(ctx context.Context, id string) (*Endpoint, error) {
	if e.mode == "hookline" {
		return e.hlRotateSecret(ctx, id)
	}
	var out mgEndpointEnvelope
	if err := e.mgDo(ctx, "POST", "/v1/events/endpoints/"+id+"/rotate-secret", map[string]any{}, &out); err != nil {
		return nil, err
	}
	ep := out.Endpoint.toPublic()
	return &ep, nil
}

// ────────────────────────────────────────────────────────────────────────────
// Subscriptions
// ────────────────────────────────────────────────────────────────────────────

func (e *EventsClient) CreateSubscription(ctx context.Context, endpointID string, eventTypes []string) (*WebhookSubscription, error) {
	body := map[string]any{"endpoint_id": endpointID, "event_types": eventTypes}
	if e.mode == "hookline" {
		var out WebhookSubscription
		if err := e.hlDo(ctx, "POST", "/v1/subscriptions", body, &out); err != nil {
			return nil, err
		}
		return &out, nil
	}
	var out mgSubscriptionEnvelope
	if err := e.mgDo(ctx, "POST", "/v1/events/subscriptions", body, &out); err != nil {
		return nil, err
	}
	sub := out.Subscription.toPublic()
	return &sub, nil
}

func (e *EventsClient) ListSubscriptions(ctx context.Context, endpointID string) ([]*WebhookSubscription, error) {
	path := fmt.Sprintf("/v1/events/subscriptions?endpoint_id=%s", endpointID)
	if e.mode == "hookline" {
		var out struct {
			Subscriptions []*WebhookSubscription `json:"subscriptions"`
		}
		if err := e.hlDo(ctx, "GET", fmt.Sprintf("/v1/subscriptions?endpoint_id=%s", endpointID), nil, &out); err != nil {
			return nil, err
		}
		return out.Subscriptions, nil
	}
	var out struct {
		Subscriptions []mgSubscriptionDTO `json:"subscriptions"`
	}
	if err := e.mgDo(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	subscriptions := make([]*WebhookSubscription, 0, len(out.Subscriptions))
	for i := range out.Subscriptions {
		sub := out.Subscriptions[i].toPublic()
		subscriptions = append(subscriptions, &sub)
	}
	return subscriptions, nil
}

func (e *EventsClient) DeleteSubscription(ctx context.Context, subscriptionID string) error {
	if e.mode == "hookline" {
		return e.hlDo(ctx, "DELETE", "/v1/subscriptions/"+subscriptionID, nil, nil)
	}
	return e.mgDo(ctx, "DELETE", "/v1/events/subscriptions/"+subscriptionID, nil, nil)
}

// ────────────────────────────────────────────────────────────────────────────
// Delivery history
// ────────────────────────────────────────────────────────────────────────────

func (e *EventsClient) ListDeliveries(ctx context.Context, endpointID string) ([]*Delivery, error) {
	if e.mode == "hookline" {
		return e.hlListDeliveries(ctx, endpointID)
	}
	var out struct {
		Deliveries []mgDeliveryDTO `json:"deliveries"`
	}
	path := "/v1/events/deliveries"
	if endpointID != "" {
		path += "?endpoint_id=" + endpointID
	}
	if err := e.mgDo(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	deliveries := make([]*Delivery, 0, len(out.Deliveries))
	for i := range out.Deliveries {
		d := out.Deliveries[i].toPublic()
		deliveries = append(deliveries, &d)
	}
	return deliveries, nil
}

func (e *EventsClient) RetryDelivery(ctx context.Context, endpointID, deliveryID string) error {
	if e.mode == "hookline" {
		return e.hlRetryDelivery(ctx, endpointID, deliveryID)
	}
	_ = endpointID // kept for API compatibility; mg-events retries by delivery_id.
	path := fmt.Sprintf("/v1/events/deliveries/%s/retry", deliveryID)
	return e.mgDo(ctx, "POST", path, map[string]any{}, nil)
}

// ────────────────────────────────────────────────────────────────────────────
// DLQ
// ────────────────────────────────────────────────────────────────────────────

func (e *EventsClient) ListDLQ(ctx context.Context, endpointID string) ([]*DLQEntry, error) {
	path := "/v1/dlq"
	if endpointID != "" {
		path += "?endpoint_id=" + endpointID
	}
	if e.mode == "hookline" {
		var out struct {
			Entries []*DLQEntry `json:"items"`
		}
		if err := e.hlDo(ctx, "GET", path, nil, &out); err != nil {
			return nil, err
		}
		return out.Entries, nil
	}
	var out struct {
		Entries []mgDLQEntryDTO `json:"entries"`
	}
	path = "/v1/events/dlq"
	if endpointID != "" {
		path += "?endpoint_id=" + endpointID
	}
	if err := e.mgDo(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	entries := make([]*DLQEntry, 0, len(out.Entries))
	for i := range out.Entries {
		entry := out.Entries[i].toPublic()
		entries = append(entries, &entry)
	}
	return entries, nil
}

func (e *EventsClient) ReplayDLQ(ctx context.Context, deliveryIDs []string) (int, error) {
	body := map[string]any{"delivery_ids": deliveryIDs}
	var out struct {
		Replayed int `json:"replayed"`
	}
	if e.mode == "hookline" {
		if err := e.hlDo(ctx, "POST", "/v1/dlq/replay", body, &out); err != nil {
			return 0, err
		}
		return out.Replayed, nil
	}
	if err := e.mgDo(ctx, "POST", "/v1/events/dlq/replay", body, &out); err != nil {
		return 0, err
	}
	return out.Replayed, nil
}

// ────────────────────────────────────────────────────────────────────────────
// mg-events DTOs (snake_case contract via grpc_json_transcoder)
// ────────────────────────────────────────────────────────────────────────────

type mgCreateEndpointRequestDTO struct {
	URL         string   `json:"url"`
	Description string   `json:"description,omitempty"`
	EventTypes  []string `json:"event_types,omitempty"`
}

type mgUpdateEndpointRequestDTO struct {
	EndpointID  string   `json:"endpoint_id"`
	URL         string   `json:"url,omitempty"`
	Description string   `json:"description,omitempty"`
	Status      string   `json:"status,omitempty"`
	EventTypes  []string `json:"event_types,omitempty"`
}

type mgEndpointDTO struct {
	EndpointID    string          `json:"endpoint_id"`
	URL           string          `json:"url"`
	Description   string          `json:"description"`
	EventTypes    []string        `json:"event_types"`
	Status        string          `json:"status"`
	SigningSecret string          `json:"signing_secret"`
	CreatedAt     json.RawMessage `json:"created_at"`
	UpdatedAt     json.RawMessage `json:"updated_at"`
}

func (d mgEndpointDTO) toPublic() Endpoint {
	return Endpoint{
		ID:            d.EndpointID,
		URL:           d.URL,
		Description:   d.Description,
		EventTypes:    d.EventTypes,
		Status:        d.Status,
		SigningSecret: d.SigningSecret,
		CreatedAt:     decodeProtoTimestampMillis(d.CreatedAt),
		UpdatedAt:     decodeProtoTimestampMillis(d.UpdatedAt),
	}
}

type mgEndpointEnvelope struct {
	Endpoint mgEndpointDTO
}

func (e *mgEndpointEnvelope) UnmarshalJSON(data []byte) error {
	var wrapped struct {
		Endpoint *mgEndpointDTO `json:"endpoint"`
	}
	if err := json.Unmarshal(data, &wrapped); err == nil && wrapped.Endpoint != nil {
		e.Endpoint = *wrapped.Endpoint
		return nil
	}
	var direct mgEndpointDTO
	if err := json.Unmarshal(data, &direct); err != nil {
		return err
	}
	e.Endpoint = direct
	return nil
}

type mgSubscriptionDTO struct {
	SubscriptionID string          `json:"subscription_id"`
	EndpointID     string          `json:"endpoint_id"`
	TenantID       string          `json:"tenant_id"`
	EventTypes     []string        `json:"event_types"`
	Status         string          `json:"status"`
	CreatedAt      json.RawMessage `json:"created_at"`
}

func (d mgSubscriptionDTO) toPublic() WebhookSubscription {
	return WebhookSubscription{
		ID:         d.SubscriptionID,
		EndpointID: d.EndpointID,
		TenantID:   d.TenantID,
		EventTypes: d.EventTypes,
		Status:     d.Status,
		CreatedAt:  decodeProtoTimestampMillis(d.CreatedAt),
	}
}

type mgSubscriptionEnvelope struct {
	Subscription mgSubscriptionDTO
}

func (e *mgSubscriptionEnvelope) UnmarshalJSON(data []byte) error {
	var wrapped struct {
		Subscription *mgSubscriptionDTO `json:"subscription"`
	}
	if err := json.Unmarshal(data, &wrapped); err == nil && wrapped.Subscription != nil {
		e.Subscription = *wrapped.Subscription
		return nil
	}
	var direct mgSubscriptionDTO
	if err := json.Unmarshal(data, &direct); err != nil {
		return err
	}
	e.Subscription = direct
	return nil
}

type mgDeliveryDTO struct {
	DeliveryID     string          `json:"delivery_id"`
	EndpointID     string          `json:"endpoint_id"`
	EventID        string          `json:"event_id"`
	Status         string          `json:"status"`
	AttemptCount   int             `json:"attempt_count"`
	ResponseStatus int             `json:"response_status"`
	CreatedAt      json.RawMessage `json:"created_at"`
	LastAttemptAt  json.RawMessage `json:"last_attempt_at"`
	NextRetryAt    json.RawMessage `json:"next_retry_at"`
}

func (d mgDeliveryDTO) toPublic() Delivery {
	return Delivery{
		ID:             d.DeliveryID,
		EndpointID:     d.EndpointID,
		EventID:        d.EventID,
		Status:         d.Status,
		AttemptCount:   d.AttemptCount,
		ResponseStatus: d.ResponseStatus,
		NextRetryAt:    decodeProtoTimestampMillis(d.NextRetryAt),
		CreatedAt:      decodeProtoTimestampMillis(d.CreatedAt),
		UpdatedAt:      decodeProtoTimestampMillis(d.LastAttemptAt),
	}
}

type mgDLQEntryDTO struct {
	DeliveryID   string          `json:"delivery_id"`
	EndpointID   string          `json:"endpoint_id"`
	EventID      string          `json:"event_id"`
	EventType    string          `json:"event_type"`
	Payload      string          `json:"payload"`
	AttemptCount int             `json:"attempt_count"`
	LastError    string          `json:"last_error"`
	FailedAt     json.RawMessage `json:"failed_at"`
}

func (d mgDLQEntryDTO) toPublic() DLQEntry {
	return DLQEntry{
		DeliveryID:   d.DeliveryID,
		EndpointID:   d.EndpointID,
		EventID:      d.EventID,
		EventType:    d.EventType,
		Payload:      d.Payload,
		AttemptCount: d.AttemptCount,
		LastError:    d.LastError,
		FailedAt:     decodeProtoTimestampMillis(d.FailedAt),
	}
}

func decodeProtoTimestampMillis(raw json.RawMessage) int64 {
	if len(raw) == 0 || string(raw) == "null" {
		return 0
	}

	var n int64
	if err := json.Unmarshal(raw, &n); err == nil {
		return n
	}

	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		if s == "" {
			return 0
		}
		if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
			return t.UnixMilli()
		}
		if i, err := strconv.ParseInt(s, 10, 64); err == nil {
			return i
		}
	}

	var obj struct {
		Seconds json.RawMessage `json:"seconds"`
		Nanos   int64           `json:"nanos"`
	}
	if err := json.Unmarshal(raw, &obj); err == nil {
		var sec int64
		if len(obj.Seconds) > 0 {
			if err := json.Unmarshal(obj.Seconds, &sec); err != nil {
				var secStr string
				if err := json.Unmarshal(obj.Seconds, &secStr); err == nil {
					if parsed, err := strconv.ParseInt(secStr, 10, 64); err == nil {
						sec = parsed
					}
				}
			}
		}
		if sec != 0 || obj.Nanos != 0 {
			return sec*1000 + obj.Nanos/1_000_000
		}
	}

	return 0
}

// ────────────────────────────────────────────────────────────────────────────
// Mashgate-mode HTTP helper (with retry + trace propagation)
// ────────────────────────────────────────────────────────────────────────────

func (e *EventsClient) mgDo(ctx context.Context, method, path string, body, out any) error {
	return withRetry(ctx, func() error {
		return e.doOnce(ctx, method, e.mgBase+path, "", map[string]string{"X-API-Key": e.mgKey}, body, out)
	})
}

// ────────────────────────────────────────────────────────────────────────────
// HookLine-mode HTTP helper (with retry + trace propagation)
// ────────────────────────────────────────────────────────────────────────────

func (e *EventsClient) hlDo(ctx context.Context, method, path string, body, out any) error {
	return withRetry(ctx, func() error {
		return e.doOnce(ctx, method, e.hlBase+path, e.hlKey, nil, body, out)
	})
}

// doOnce executes a single HTTP attempt with auth + trace headers.
func (e *EventsClient) doOnce(ctx context.Context, method, url, apiKey string, extraHeaders map[string]string, body, out any) error {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("events: marshal: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("events: build request: %w", err)
	}

	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	// Trace propagation: forward traceparent/tracestate if present in context.
	if tp, ok := ctx.Value(traceparentKey{}).(string); ok && tp != "" {
		req.Header.Set("Traceparent", tp)
	}
	if ts, ok := ctx.Value(tracestateKey{}).(string); ok && ts != "" {
		req.Header.Set("Tracestate", ts)
	}
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}

	resp, err := e.http.Do(req)
	if err != nil {
		return fmt.Errorf("events: http: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("events: read body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return parseEventsError(resp.StatusCode, raw)
	}
	if out != nil && len(raw) > 0 {
		if err := json.Unmarshal(raw, out); err != nil {
			return fmt.Errorf("events: decode: %w", err)
		}
	}
	return nil
}

// ────────────────────────────────────────────────────────────────────────────
// HookLine convenience wrappers
// ────────────────────────────────────────────────────────────────────────────

func (e *EventsClient) hlCreateEndpoint(ctx context.Context, req CreateEndpointRequest) (*Endpoint, error) {
	var out Endpoint
	if err := e.hlDo(ctx, "POST", "/v1/endpoints", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (e *EventsClient) hlListEndpoints(ctx context.Context) ([]*Endpoint, error) {
	var out struct {
		Endpoints []*Endpoint `json:"endpoints"`
	}
	if err := e.hlDo(ctx, "GET", "/v1/endpoints", nil, &out); err != nil {
		return nil, err
	}
	return out.Endpoints, nil
}

func (e *EventsClient) hlUpdateEndpoint(ctx context.Context, id string, req UpdateEndpointRequest) (*Endpoint, error) {
	var out Endpoint
	if err := e.hlDo(ctx, "PUT", "/v1/endpoints/"+id, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (e *EventsClient) hlDeleteEndpoint(ctx context.Context, id string) error {
	return e.hlDo(ctx, "DELETE", "/v1/endpoints/"+id, nil, nil)
}

func (e *EventsClient) hlRotateSecret(ctx context.Context, id string) (*Endpoint, error) {
	var out Endpoint
	if err := e.hlDo(ctx, "POST", "/v1/endpoints/"+id+"/rotate-secret", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (e *EventsClient) hlListDeliveries(ctx context.Context, endpointID string) ([]*Delivery, error) {
	var out struct {
		Deliveries []*Delivery `json:"deliveries"`
	}
	if err := e.hlDo(ctx, "GET", fmt.Sprintf("/v1/deliveries?endpoint_id=%s", endpointID), nil, &out); err != nil {
		return nil, err
	}
	return out.Deliveries, nil
}

func (e *EventsClient) hlRetryDelivery(ctx context.Context, endpointID, deliveryID string) error {
	return e.hlDo(ctx, "POST", fmt.Sprintf("/v1/deliveries/%s/retry", deliveryID), nil, nil)
}

// ────────────────────────────────────────────────────────────────────────────
// Retry: exponential backoff with jitter on 429 / 5xx
// ────────────────────────────────────────────────────────────────────────────

const (
	retryMaxAttempts = 3
	retryBaseMs      = 200
)

func withRetry(ctx context.Context, fn func() error) error {
	var err error
	for attempt := 0; attempt < retryMaxAttempts; attempt++ {
		err = fn()
		if err == nil {
			return nil
		}
		if !isRetryable(err) {
			return err
		}
		delay := time.Duration(retryBaseMs*(1<<attempt)+rand.Intn(100)) * time.Millisecond
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}
	return err
}

func isRetryable(err error) bool {
	switch err.(type) {
	case *QuotaExceededError, *EventsServiceUnavailableError:
		return true
	}
	return false
}

// ────────────────────────────────────────────────────────────────────────────
// Events-domain typed errors
// ────────────────────────────────────────────────────────────────────────────

// EndpointNotFoundError is returned when the requested endpoint does not exist (HTTP 404).
type EndpointNotFoundError struct {
	ID string
}

func (e *EndpointNotFoundError) Error() string {
	return fmt.Sprintf("events: endpoint not found: %s", e.ID)
}

// SubscriptionConflictError is returned when a subscription already exists (HTTP 409).
type SubscriptionConflictError struct {
	Message string
}

func (e *SubscriptionConflictError) Error() string {
	return fmt.Sprintf("events: subscription conflict: %s", e.Message)
}

// QuotaExceededError is returned on HTTP 429.
type QuotaExceededError struct {
	Message    string
	RetryAfter int // seconds
}

func (e *QuotaExceededError) Error() string {
	return fmt.Sprintf("events: quota exceeded: %s (retry after %ds)", e.Message, e.RetryAfter)
}

// EventsServiceUnavailableError wraps 503 / transient upstream errors.
type EventsServiceUnavailableError struct {
	Message string
}

func (e *EventsServiceUnavailableError) Error() string {
	return fmt.Sprintf("events: service unavailable: %s", e.Message)
}

// parseEventsError maps HTTP status to a typed error.
func parseEventsError(status int, body []byte) error {
	var envelope struct {
		Message    string `json:"message"`
		Error      string `json:"error"`
		ID         string `json:"id"`
		RetryAfter int    `json:"retry_after"`
	}
	_ = json.Unmarshal(body, &envelope)
	msg := envelope.Message
	if msg == "" {
		msg = envelope.Error
	}
	if msg == "" {
		msg = string(body)
	}
	switch status {
	case http.StatusNotFound:
		return &EndpointNotFoundError{ID: envelope.ID}
	case http.StatusConflict:
		return &SubscriptionConflictError{Message: msg}
	case http.StatusTooManyRequests:
		return &QuotaExceededError{Message: msg, RetryAfter: envelope.RetryAfter}
	case http.StatusServiceUnavailable, http.StatusBadGateway, http.StatusGatewayTimeout:
		return &EventsServiceUnavailableError{Message: msg}
	case http.StatusUnauthorized, http.StatusForbidden:
		return &UnauthorizedError{Message: msg}
	default:
		return &APIError{StatusCode: status, Body: string(body)}
	}
}

// ────────────────────────────────────────────────────────────────────────────
// Trace propagation context keys
// ────────────────────────────────────────────────────────────────────────────

// WithTraceparent stores a W3C traceparent value in the context for
// automatic propagation through all Events API calls.
func WithTraceparent(ctx context.Context, traceparent string) context.Context {
	return context.WithValue(ctx, traceparentKey{}, traceparent)
}

// WithTracestate stores a W3C tracestate value in the context.
func WithTracestate(ctx context.Context, tracestate string) context.Context {
	return context.WithValue(ctx, tracestateKey{}, tracestate)
}

type traceparentKey struct{}
type tracestateKey struct{}
