package mashgate

import (
	"context"
	"fmt"
	"net/url"

	"github.com/google/uuid"
)

// ────────────────────────────────────────────────────────────────────────────
// Request types
// ────────────────────────────────────────────────────────────────────────────

// CreatePaymentRequest creates a payment intent.
// Set CaptureMode to "MANUAL" to authorize first and capture later.
type CreatePaymentRequest struct {
	Amount         Money              `json:"amount"`
	OrderID        string             `json:"orderId"`
	CaptureMode    string             `json:"captureMode,omitempty"` // "AUTO" | "MANUAL", default "AUTO"
	Card           *CardPaymentMethod `json:"card,omitempty"`
	Metadata       map[string]string  `json:"metadata,omitempty"`
	// IdempotencyKey is sent as a header. Auto-generated with uuid if empty.
	IdempotencyKey string `json:"-"`
}

// RefundRequest requests a (partial) refund on a captured payment.
type RefundRequest struct {
	Amount         Money  `json:"amount"`
	Reason         string `json:"reason,omitempty"`
	Note           string `json:"note,omitempty"`
	// IdempotencyKey is sent as a header. Auto-generated with uuid if empty.
	IdempotencyKey string `json:"-"`
}

// ListPaymentsParams filters the ListPayments response.
type ListPaymentsParams struct {
	Status   string // optional: "pending", "authorized", "captured", etc.
	Page     int    // 1-based, default 1
	PageSize int    // default 20, max 100
}

// ────────────────────────────────────────────────────────────────────────────
// Response wrappers
// ────────────────────────────────────────────────────────────────────────────

type listPaymentsResponse struct {
	Payments []*Payment `json:"payments"`
}

// ────────────────────────────────────────────────────────────────────────────
// Payment methods
// ────────────────────────────────────────────────────────────────────────────

// CreatePayment creates a new payment intent.
// An idempotency key is auto-generated if not provided in req.IdempotencyKey.
func (c *Client) CreatePayment(ctx context.Context, req CreatePaymentRequest) (*Payment, error) {
	key := req.IdempotencyKey
	if key == "" {
		key = uuid.NewString()
	}

	var out Payment
	err := c.doWithHeader(ctx, "POST", "/v1/payments",
		map[string]string{"Idempotency-Key": key},
		req, &out,
	)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// GetPayment retrieves a payment by its ID.
func (c *Client) GetPayment(ctx context.Context, paymentID string) (*Payment, error) {
	var out Payment
	if err := c.do(ctx, "GET", "/v1/payments/"+paymentID, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AuthorizePayment submits an authorization request for a created payment.
// The payment must be in "pending" status.
func (c *Client) AuthorizePayment(ctx context.Context, paymentID string) (*Payment, error) {
	var out Payment
	if err := c.do(ctx, "POST", "/v1/payments/"+paymentID+"/authorize", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CapturePayment captures an authorized payment.
// Only valid when CaptureMode is "MANUAL" and status is "authorized".
func (c *Client) CapturePayment(ctx context.Context, paymentID string) (*Payment, error) {
	var out Payment
	if err := c.do(ctx, "POST", "/v1/payments/"+paymentID+"/capture", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// VoidPayment voids an authorized payment, releasing the hold.
// Only valid when status is "authorized".
func (c *Client) VoidPayment(ctx context.Context, paymentID string) (*Payment, error) {
	var out Payment
	if err := c.do(ctx, "POST", "/v1/payments/"+paymentID+"/void", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// RefundPayment creates a refund against a captured payment.
// Partial refunds are supported — req.Amount may be less than the captured amount.
func (c *Client) RefundPayment(ctx context.Context, paymentID string, req RefundRequest) (*Payment, error) {
	key := req.IdempotencyKey
	if key == "" {
		key = uuid.NewString()
	}

	var out Payment
	err := c.doWithHeader(ctx, "POST", "/v1/payments/"+paymentID+"/refund",
		map[string]string{"Idempotency-Key": key},
		req, &out,
	)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ListPayments returns a paginated list of payments for the current tenant.
func (c *Client) ListPayments(ctx context.Context, params ListPaymentsParams) ([]*Payment, error) {
	q := url.Values{}
	if params.Status != "" {
		q.Set("status", params.Status)
	}
	if params.Page > 0 {
		q.Set("page", fmt.Sprintf("%d", params.Page))
	}
	if params.PageSize > 0 {
		q.Set("pageSize", fmt.Sprintf("%d", params.PageSize))
	}

	path := "/v1/payments"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	var out listPaymentsResponse
	if err := c.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return out.Payments, nil
}
