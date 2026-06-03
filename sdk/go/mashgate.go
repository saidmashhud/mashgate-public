// Package mashgate provides a Go client for the Mashgate payment API.
//
// Mashgate is a payment gateway for Central Asia supporting Uzcard, Humo,
// Click, Payme, Oson, Visa, and Mastercard.
//
// Usage:
//
//	client := mashgate.New("https://api.mashgate.uz", os.Getenv("MASHGATE_API_KEY"))
//	// or with functional options:
//	client := mashgate.NewClient(os.Getenv("MASHGATE_API_KEY"),
//	    mashgate.WithTimeout(10*time.Second),
//	    mashgate.WithMaxRetries(3),
//	)
//	session, err := client.CreateCheckout(ctx, mashgate.CreateCheckoutRequest{...})
package mashgate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

// ────────────────────────────────────────────────────────────────────────────
// Error type
// ────────────────────────────────────────────────────────────────────────────

// Error is returned by all Mashgate API calls.
//
//	var mgErr *mashgate.Error
//	if errors.As(err, &mgErr) {
//	    log.Printf("failed: code=%s request_id=%s doc=%s", mgErr.Code, mgErr.RequestID, mgErr.DocURL())
//	}
type Error struct {
	// Code is a machine-readable error identifier (e.g. "card_declined").
	Code string
	// Message is a human-readable description.
	Message string
	// StatusCode is the HTTP status code.
	StatusCode int
	// RequestID is the X-Request-Id header value — include this in support requests.
	RequestID string
	// Param is the request parameter that caused a validation error, if applicable.
	Param string
}

func (e *Error) Error() string {
	if e.RequestID != "" {
		return fmt.Sprintf("mashgate: %s (code=%s, request_id=%s)", e.Message, e.Code, e.RequestID)
	}
	return fmt.Sprintf("mashgate: %s (code=%s)", e.Message, e.Code)
}

// DocURL returns a link to the documentation for this error code.
func (e *Error) DocURL() string {
	return "https://docs.mashgate.io/errors#" + e.Code
}

// ────────────────────────────────────────────────────────────────────────────
// Client options
// ────────────────────────────────────────────────────────────────────────────

// Option configures the Mashgate client.
type Option func(*Client)

// WithBaseURL overrides the default API base URL ("https://api.mashgate.uz").
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.baseURL = strings.TrimRight(url, "/")
	}
}

// WithTimeout sets the HTTP client timeout (default: 30s).
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = d
	}
}

// WithMaxRetries sets the maximum number of retries on 5xx / network errors (default: 0).
// Retries use exponential backoff with jitter: ~1s, ~2s, ~4s … up to ~32s.
func WithMaxRetries(n int) Option {
	return func(c *Client) {
		c.maxRetries = n
	}
}

// WithHTTPClient replaces the default HTTP client (useful for testing or custom transports).
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		c.httpClient = hc
	}
}

// ────────────────────────────────────────────────────────────────────────────
// Client
// ────────────────────────────────────────────────────────────────────────────

// Client is the Mashgate API client. Create one with New() or NewClient() and reuse it.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	maxRetries int

	// Events is populated by calling client.WithEvents(cfg).
	// It provides endpoint and delivery management for webhook infrastructure.
	Events *EventsClient

	// BaaS service clients — available after calling New().
	Chat    *ChatClient
	Notify  *NotifyClient
	Storage *StorageClient
	Flags   *FlagsClient
	Logs    *LogsClient
	// Mail capability — Mashgate core primitive (ADR-0019). Mailbox / message
	// CRUD plus admin tenant operations (mailboxes, domains, DKIM rotation).
	// Subscribe to mail.received / mail.sent / mail.delivered / mail.bounced
	// events via webhooks.
	Mail *MailClient

	// Part 4 service clients — available after calling New().
	Billing       *BillingClient
	Subscriptions *SubscriptionsClient
	Invoices      *InvoicesClient
	PaymentLinks  *PaymentLinksClient
	Guard         *GuardClient

	// v1.7.0 — eight resources added to close TS-SDK gap.
	Analytics     *AnalyticsClient
	Chain         *ChainClient
	Developer     *DeveloperClient
	LocalPayments *LocalPaymentsClient
	Metering      *MeteringClient
	Risk          *RiskClient
	Settings      *SettingsClient
	WalletAdmin   *WalletAdminClient
}

func initClients(c *Client) {
	c.Chat = &ChatClient{c: c}
	c.Notify = &NotifyClient{c: c}
	c.Storage = &StorageClient{c: c}
	c.Flags = &FlagsClient{c: c}
	c.Logs = &LogsClient{c: c}
	c.Mail = &MailClient{c: c}
	c.Billing = &BillingClient{c: c}
	c.Subscriptions = &SubscriptionsClient{c: c}
	c.Invoices = &InvoicesClient{c: c}
	c.PaymentLinks = &PaymentLinksClient{c: c}
	c.Guard = &GuardClient{c: c}

	// v1.7.0 resources
	c.Analytics = &AnalyticsClient{c: c}
	c.Chain = &ChainClient{c: c}
	c.Developer = &DeveloperClient{c: c}
	c.LocalPayments = &LocalPaymentsClient{c: c}
	c.Metering = &MeteringClient{c: c}
	c.Risk = &RiskClient{c: c}
	c.Settings = &SettingsClient{c: c}
	c.WalletAdmin = &WalletAdminClient{c: c}
}

// New creates a Mashgate API client.
//
//	baseURL — e.g. "https://api.mashgate.uz" or "http://localhost:9661" for local dev
//	apiKey  — mg_test_... or mg_live_... key from the Developer Portal
func New(baseURL, apiKey string) *Client {
	c := &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	initClients(c)
	return c
}

// NewClient creates a Mashgate API client using functional options.
// Default base URL is "https://api.mashgate.uz".
//
//	client := mashgate.NewClient(os.Getenv("MASHGATE_API_KEY"),
//	    mashgate.WithTimeout(10*time.Second),
//	    mashgate.WithMaxRetries(3),
//	)
func NewClient(apiKey string, opts ...Option) *Client {
	c := &Client{
		baseURL: "https://api.mashgate.uz",
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	initClients(c)
	return c
}

// WithHTTPClient replaces the default HTTP client. Useful for testing or custom transports.
// Deprecated: use the WithHTTPClient Option with NewClient instead.
func (c *Client) WithHTTPClient(hc *http.Client) *Client {
	c.httpClient = hc
	return c
}

// ────────────────────────────────────────────────────────────────────────────
// Internal HTTP helpers
// ────────────────────────────────────────────────────────────────────────────

// do executes an HTTP request against the Mashgate API with automatic retry.
func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	return c.doRequest(ctx, method, path, nil, body, out)
}

// doWithHeader is like do but also sets extra headers (e.g. Idempotency-Key).
func (c *Client) doWithHeader(ctx context.Context, method, path string, headers map[string]string, body, out any) error {
	return c.doRequest(ctx, method, path, headers, body, out)
}

func (c *Client) doRequest(ctx context.Context, method, path string, extraHeaders map[string]string, body, out any) error {
	// Pre-marshal the body once so it can be replayed on retries.
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("mashgate: marshal request: %w", err)
		}
	}

	maxAttempts := 1 + c.maxRetries
	var lastErr error

	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			// Exponential backoff with jitter: ~1s, ~2s, ~4s … up to ~32s
			base := math.Min(1000*math.Pow(2, float64(attempt-1)), 32_000)
			jitter := float64(rand.Intn(1000))
			delay := time.Duration(base+jitter) * time.Millisecond
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		var bodyReader io.Reader
		if bodyBytes != nil {
			bodyReader = bytes.NewReader(bodyBytes)
		}

		req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
		if err != nil {
			return fmt.Errorf("mashgate: build request: %w", err)
		}

		if c.apiKey != "" {
			req.Header.Set("X-API-Key", c.apiKey)
		}
		req.Header.Set("Accept", "application/json")
		if bodyBytes != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		for k, v := range extraHeaders {
			req.Header.Set(k, v)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			apiErr := &Error{Code: "network_error", Message: err.Error()}
			if attempt < maxAttempts-1 {
				lastErr = apiErr
				continue
			}
			return apiErr
		}

		raw, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return fmt.Errorf("mashgate: read response: %w", readErr)
		}

		requestID := resp.Header.Get("X-Request-Id")

		if resp.StatusCode >= 400 {
			apiErr := parseAPIError(resp.StatusCode, raw, requestID)
			if resp.StatusCode >= 500 && attempt < maxAttempts-1 {
				lastErr = apiErr
				continue
			}
			return apiErr
		}

		if out != nil && len(raw) > 0 {
			if err := json.Unmarshal(raw, out); err != nil {
				return fmt.Errorf("mashgate: decode response: %w", err)
			}
		}
		return nil
	}

	return lastErr
}

// ────────────────────────────────────────────────────────────────────────────
// Error parsing
// ────────────────────────────────────────────────────────────────────────────

func parseAPIError(status int, body []byte, requestID string) error {
	var envelope struct {
		Message string `json:"message"`
		Error   string `json:"error"`
		Code    string `json:"code"`
		Param   string `json:"param"`
		Field   string `json:"field"`
	}
	_ = json.Unmarshal(body, &envelope)

	msg := envelope.Message
	if msg == "" {
		msg = envelope.Error
	}
	if msg == "" {
		msg = string(body)
	}

	code := envelope.Code
	if code == "" {
		switch {
		case status == 401 || status == 403:
			code = "unauthorized"
		case status == 404:
			code = "not_found"
		case status == 422:
			code = "validation_error"
		case status == 429:
			code = "rate_limit_exceeded"
		default:
			code = "api_error"
		}
	}

	return &Error{
		Code:       code,
		Message:    msg,
		StatusCode: status,
		RequestID:  requestID,
		Param:      firstNonEmpty(envelope.Param, envelope.Field),
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

// ────────────────────────────────────────────────────────────────────────────
// Legacy error types (kept for backward compatibility)
// ────────────────────────────────────────────────────────────────────────────

// NotFoundError is returned when the requested resource does not exist (HTTP 404).
// Deprecated: use *Error with Code == "not_found" instead.
type NotFoundError struct{ ID string }

func (e *NotFoundError) Error() string {
	if e.ID != "" {
		return fmt.Sprintf("mashgate: not found: %s", e.ID)
	}
	return "mashgate: not found"
}

// UnauthorizedError is returned on HTTP 401 or 403.
// Deprecated: use *Error with Code == "unauthorized" instead.
type UnauthorizedError struct{ Message string }

func (e *UnauthorizedError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("mashgate: unauthorized: %s", e.Message)
	}
	return "mashgate: unauthorized"
}

// ValidationError is returned on HTTP 422.
// Deprecated: use *Error with Code == "validation_error" instead.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("mashgate: validation error on %q: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("mashgate: validation error: %s", e.Message)
}

// APIError is the fallback for any other 4xx/5xx response.
// Deprecated: use *Error instead.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("mashgate: api error %d: %s", e.StatusCode, e.Body)
}
