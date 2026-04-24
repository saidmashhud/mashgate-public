package mashgate

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ────────────────────────────────────────────────────────────────────────────
// Signature verification errors
// ────────────────────────────────────────────────────────────────────────────

var (
	// ErrInvalidSignature is returned when the HMAC does not match.
	ErrInvalidSignature = errors.New("invalid webhook signature")
	// ErrTimestampTooOld is returned when the timestamp is older than 5 minutes,
	// indicating a possible replay attack.
	ErrTimestampTooOld = errors.New("webhook timestamp too old — possible replay attack")
	// ErrMalformedHeader is returned when the signature header cannot be parsed.
	ErrMalformedHeader = errors.New("malformed webhook header")
)

const signatureMaxAge = 5 * time.Minute

// VerifySignature verifies an incoming webhook from Mashgate.
//
// Parameters come directly from the HTTP request:
//
//	secret    — SigningSecret from WebhookEndpoint (set as env var, never log it)
//	timestamp — x-hl-timestamp header value (Unix ms or Unix seconds as string)
//	body      — raw request body, read before any JSON parsing
//	signature — x-hl-signature header value, e.g. "v1=abc123..."
//
// Algorithm: HMAC-SHA256(secret, "{timestamp}.{body}"), hex-encoded, prefixed "v1=".
//
// Rejects if the timestamp is more than 5 minutes old to prevent replay attacks.
//
// Usage:
//
//	body, _ := io.ReadAll(r.Body)
//	err := mashgate.VerifySignature(
//	    os.Getenv("MASHGATE_WEBHOOK_SECRET"),
//	    r.Header.Get("x-hl-timestamp"),
//	    string(body),
//	    r.Header.Get("x-hl-signature"),
//	)
//	if err != nil {
//	    http.Error(w, "unauthorized", http.StatusUnauthorized)
//	    return
//	}
func VerifySignature(secret, timestamp, body, signature string) error {
	if secret == "" {
		return ErrInvalidSignature
	}
	if timestamp == "" || signature == "" {
		return ErrMalformedHeader
	}

	// Validate and parse timestamp
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return fmt.Errorf("%w: timestamp must be a Unix integer", ErrMalformedHeader)
	}
	// HookLine sends milliseconds; accept both seconds and milliseconds.
	tsSec := ts
	if ts >= 1_000_000_000_000 {
		tsSec = ts / 1000
	}
	age := time.Since(time.Unix(tsSec, 0))
	if age > signatureMaxAge || age < -signatureMaxAge {
		return ErrTimestampTooOld
	}

	// Parse signature — must be "v1=<hex>"
	parts := strings.SplitN(signature, "=", 2)
	if len(parts) != 2 || parts[0] != "v1" {
		return fmt.Errorf("%w: expected v1=<hex> format", ErrMalformedHeader)
	}
	expectedHex := parts[1]

	// Compute expected HMAC
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp + "." + body))
	computed := hex.EncodeToString(mac.Sum(nil))

	// Constant-time comparison
	if !hmac.Equal([]byte(computed), []byte(expectedHex)) {
		return ErrInvalidSignature
	}
	return nil
}

// ParseEvent parses a verified webhook body into a WebhookEvent.
// Always call VerifySignature before ParseEvent.
//
//	event, err := mashgate.ParseEvent(body)
//	switch event.EventType {
//	case mashgate.EventPaymentCaptured:
//	    // handle capture
//	}
func ParseEvent(body []byte) (*WebhookEvent, error) {
	var event WebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return nil, fmt.Errorf("mashgate: parse event: %w", err)
	}
	return &event, nil
}

// ────────────────────────────────────────────────────────────────────────────
// Request types
// ────────────────────────────────────────────────────────────────────────────

// CreateWebhookEndpointRequest registers a new webhook endpoint.
// Leave EventTypes empty to receive all event types.
type CreateWebhookEndpointRequest struct {
	URL         string   `json:"url"`
	Description string   `json:"description,omitempty"`
	EventTypes  []string `json:"events,omitempty"` // empty = all events
}

// UpdateWebhookEndpointRequest updates an existing endpoint.
// Only set the fields you want to change.
type UpdateWebhookEndpointRequest struct {
	URL         string   `json:"url,omitempty"`
	Description string   `json:"description,omitempty"`
	EventTypes  []string `json:"events,omitempty"`
	Status      string   `json:"status,omitempty"` // "active" | "disabled"
}

type listEndpointsResponse struct {
	Endpoints []*WebhookEndpoint `json:"endpoints"`
}

func mapToWebhookEndpoint(dto mgEndpointDTO) *WebhookEndpoint {
	return &WebhookEndpoint{
		EndpointID:    dto.EndpointID,
		URL:           dto.URL,
		Description:   dto.Description,
		EventTypes:    dto.EventTypes,
		Status:        dto.Status,
		SigningSecret: dto.SigningSecret,
		CreatedAt:     decodeProtoTimestampMillis(dto.CreatedAt),
	}
}

// ────────────────────────────────────────────────────────────────────────────
// Webhook endpoint methods
// ────────────────────────────────────────────────────────────────────────────

// CreateWebhookEndpoint registers a new webhook endpoint.
// The response includes SigningSecret — store it securely, it is only returned once.
func (c *Client) CreateWebhookEndpoint(ctx context.Context, req CreateWebhookEndpointRequest) (*WebhookEndpoint, error) {
	body := map[string]any{
		"url":         req.URL,
		"description": req.Description,
		"event_types": req.EventTypes,
	}
	var out mgEndpointEnvelope
	if err := c.do(ctx, "POST", "/v1/events/endpoints", body, &out); err != nil {
		return nil, err
	}
	return mapToWebhookEndpoint(out.Endpoint), nil
}

// ListWebhookEndpoints returns all registered webhook endpoints for the tenant.
func (c *Client) ListWebhookEndpoints(ctx context.Context) ([]*WebhookEndpoint, error) {
	var out struct {
		Endpoints []mgEndpointDTO `json:"endpoints"`
	}
	if err := c.do(ctx, "GET", "/v1/events/endpoints", nil, &out); err != nil {
		return nil, err
	}
	endpoints := make([]*WebhookEndpoint, 0, len(out.Endpoints))
	for i := range out.Endpoints {
		endpoints = append(endpoints, mapToWebhookEndpoint(out.Endpoints[i]))
	}
	return endpoints, nil
}

// GetWebhookEndpoint retrieves a single endpoint by its ID.
func (c *Client) GetWebhookEndpoint(ctx context.Context, endpointID string) (*WebhookEndpoint, error) {
	var out mgEndpointEnvelope
	if err := c.do(ctx, "GET", "/v1/events/endpoints/"+endpointID, nil, &out); err != nil {
		return nil, err
	}
	return mapToWebhookEndpoint(out.Endpoint), nil
}

// UpdateWebhookEndpoint updates an existing endpoint.
// Only the fields set in req are changed.
func (c *Client) UpdateWebhookEndpoint(ctx context.Context, endpointID string, req UpdateWebhookEndpointRequest) (*WebhookEndpoint, error) {
	body := map[string]any{
		"endpoint_id": endpointID,
		"url":         req.URL,
		"description": req.Description,
		"event_types": req.EventTypes,
		"status":      req.Status,
	}
	var out mgEndpointEnvelope
	if err := c.do(ctx, "PUT", "/v1/events/endpoints/"+endpointID, body, &out); err != nil {
		return nil, err
	}
	return mapToWebhookEndpoint(out.Endpoint), nil
}

// DeleteWebhookEndpoint permanently removes a webhook endpoint.
func (c *Client) DeleteWebhookEndpoint(ctx context.Context, endpointID string) error {
	return c.do(ctx, "DELETE", "/v1/events/endpoints/"+endpointID, nil, nil)
}

// RetryWebhookDelivery retries a failed webhook delivery attempt.
func (c *Client) RetryWebhookDelivery(ctx context.Context, endpointID, deliveryID string) error {
	_ = endpointID // kept for backward-compatible method signature.
	path := fmt.Sprintf("/v1/events/deliveries/%s/retry", deliveryID)
	return c.do(ctx, "POST", path, map[string]any{}, nil)
}

// TestWebhookEndpoint sends a synthetic test event to the endpoint.
func (c *Client) TestWebhookEndpoint(ctx context.Context, endpointID string) error {
	return c.do(ctx, "POST", "/v1/events/endpoints/"+endpointID+"/test", map[string]any{}, nil)
}

// ConstructEvent verifies a webhook and returns the parsed event in one call.
// It combines VerifySignature + ParseEvent.
//
// Pass the raw header values from your HTTP request:
//
//	event, err := client.ConstructEvent(
//	    body,
//	    r.Header.Get("x-hl-signature"),
//	    r.Header.Get("x-hl-timestamp"),
//	    os.Getenv("MASHGATE_WEBHOOK_SECRET"),
//	)
//	if err != nil {
//	    http.Error(w, "unauthorized", http.StatusUnauthorized)
//	    return
//	}
//	switch event.EventType {
//	case mashgate.EventPaymentCaptured:
//	    // handle
//	}
func (c *Client) ConstructEvent(body []byte, signature, timestamp, secret string) (*WebhookEvent, error) {
	if err := VerifySignature(secret, timestamp, string(body), signature); err != nil {
		return nil, err
	}
	return ParseEvent(body)
}
