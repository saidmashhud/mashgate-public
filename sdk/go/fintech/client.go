// Package mashgate is Kiro's client for Mashgate Fintech Pack APIs.
//
// Wire contracts mirror mashgate/contracts/proto/v1/{kyc,compliance,merchant,wallet}.proto
// and their TypeScript counterparts in kiro/packages/mashgate-types.
//
// All mutating calls accept an idempotency key — callers MUST provide one
// derived from Kiro domain IDs. See docs/mashgate-integration-matrix.md §8.
package fintech

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL  string
	tenantID string
	apiKey   string
	http     *http.Client

	KYC        *KYCService
	Compliance *ComplianceService
	Merchant   *MerchantService
	Wallet     *WalletService
}

func New(baseURL, tenantID, apiKey string) *Client {
	c := &Client{
		baseURL:  strings.TrimRight(baseURL, "/"),
		tenantID: tenantID,
		apiKey:   apiKey,
		http:     &http.Client{Timeout: 30 * time.Second},
	}
	c.KYC = &KYCService{c: c}
	c.Compliance = &ComplianceService{c: c}
	c.Merchant = &MerchantService{c: c}
	c.Wallet = &WalletService{c: c}
	return c
}

// APIError wraps a non-2xx response from Mashgate.
type APIError struct {
	Status int
	Body   string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("mashgate api %d: %s", e.Status, e.Body)
}

// do executes a JSON request against Mashgate.
// idempotencyKey is forwarded as `Idempotency-Key` header when non-empty.
func (c *Client) do(ctx context.Context, method, path string, body any, idempotencyKey string, out any) error {
	var reader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal: %w", err)
		}
		reader = bytes.NewReader(buf)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("X-Tenant-ID", c.tenantID)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}

	// Propagate trace context if present in ctx (populated by HTTP middleware).
	if tp, ok := ctx.Value(traceparentKey{}).(string); ok && tp != "" {
		req.Header.Set("traceparent", tp)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &APIError{Status: resp.StatusCode, Body: string(respBytes)}
	}
	if out != nil && len(respBytes) > 0 {
		return json.Unmarshal(respBytes, out)
	}
	return nil
}

type traceparentKey struct{}

// WithTraceparent attaches a W3C traceparent header value to the context so
// subsequent Mashgate calls forward it. Use for end-to-end distributed tracing.
func WithTraceparent(ctx context.Context, tp string) context.Context {
	return context.WithValue(ctx, traceparentKey{}, tp)
}
