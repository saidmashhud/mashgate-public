package mashgate

import (
	"context"

	"github.com/google/uuid"
)

// ────────────────────────────────────────────────────────────────────────────
// Request / response types
// ────────────────────────────────────────────────────────────────────────────

// CreateCheckoutRequest creates a hosted checkout session.
// Redirect the customer to CheckoutSession.CheckoutURL after creation.
type CreateCheckoutRequest struct {
	TotalAmount      Money             `json:"totalAmount"`
	Items            []LineItem        `json:"items,omitempty"`
	SuccessURL       string            `json:"successUrl"`
	CancelURL        string            `json:"cancelUrl"`
	CustomerEmail    string            `json:"customerEmail,omitempty"`
	CustomerID       string            `json:"customerId,omitempty"`
	ExpiresInMinutes int               `json:"expiresInMinutes,omitempty"` // default 30
	Metadata         map[string]string `json:"metadata,omitempty"`
	// IdempotencyKey is sent as a header. Auto-generated with uuid if empty.
	IdempotencyKey string `json:"-"`
}

// CompleteCheckoutRequest is posted by the hosted checkout page when the
// customer submits their payment details. In most integrations Mashgate's
// hosted checkout handles this automatically.
type CompleteCheckoutRequest struct {
	PaymentMethodToken string               `json:"paymentMethodToken,omitempty"`
	PaymentMethodType  string               `json:"paymentMethodType"`            // "card" | "wallet"
	PaymentMethodBrand string               `json:"paymentMethodBrand,omitempty"`
	PaymentMethodLast4 string               `json:"paymentMethodLast4,omitempty"`
	Wallet             *WalletPaymentMethod `json:"wallet,omitempty"`
}

// CompleteCheckoutResponse is returned by CompleteCheckout.
type CompleteCheckoutResponse struct {
	Success             bool   `json:"success"`
	PaymentID           string `json:"paymentId"`
	RedirectURL         string `json:"redirectUrl,omitempty"`
	WalletRedirectURL   string `json:"walletRedirectUrl,omitempty"`
	WalletTransactionID string `json:"walletTransactionId,omitempty"`
}

type createCheckoutResponse struct {
	Session     *CheckoutSession `json:"session"`
	CheckoutURL string           `json:"checkoutUrl"`
}

// ────────────────────────────────────────────────────────────────────────────
// Checkout methods
// ────────────────────────────────────────────────────────────────────────────

// CreateCheckout creates a hosted checkout session and returns the session
// with its CheckoutURL. Redirect the customer to CheckoutURL to begin payment.
//
//	session, err := client.CreateCheckout(ctx, mashgate.CreateCheckoutRequest{
//	    TotalAmount:   mashgate.Money{Amount: "2500000.00", Currency: "UZS"},
//	    SuccessURL:    "https://zist.uz/booking/success",
//	    CancelURL:     "https://zist.uz/booking/cancel",
//	    CustomerEmail: "guest@example.com",
//	})
//	// redirect customer to session.CheckoutURL
func (c *Client) CreateCheckout(ctx context.Context, req CreateCheckoutRequest) (*CheckoutSession, error) {
	key := req.IdempotencyKey
	if key == "" {
		key = uuid.NewString()
	}

	var raw createCheckoutResponse
	err := c.doWithHeader(ctx, "POST", "/v1/checkout/sessions",
		map[string]string{"Idempotency-Key": key},
		req, &raw,
	)
	if err != nil {
		return nil, err
	}

	// API may return the session directly or nested under "session"
	if raw.Session != nil {
		if raw.Session.CheckoutURL == "" && raw.CheckoutURL != "" {
			raw.Session.CheckoutURL = raw.CheckoutURL
		}
		return raw.Session, nil
	}

	// Fallback: the response itself is a CheckoutSession
	var session CheckoutSession
	err = c.doWithHeader(ctx, "POST", "/v1/checkout/sessions",
		map[string]string{"Idempotency-Key": uuid.NewString()},
		req, &session,
	)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// GetCheckout retrieves a checkout session by its ID.
func (c *Client) GetCheckout(ctx context.Context, sessionID string) (*CheckoutSession, error) {
	var out CheckoutSession
	if err := c.do(ctx, "GET", "/v1/checkout/sessions/"+sessionID, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CompleteCheckout submits payment details for a checkout session.
// In most integrations this is called by Mashgate's hosted checkout page,
// not by the merchant directly.
func (c *Client) CompleteCheckout(ctx context.Context, sessionID string, req CompleteCheckoutRequest) (*CompleteCheckoutResponse, error) {
	var out CompleteCheckoutResponse
	err := c.do(ctx, "POST", "/v1/checkout/sessions/"+sessionID+"/complete", req, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ExpireCheckout marks a checkout session as expired, preventing further use.
func (c *Client) ExpireCheckout(ctx context.Context, sessionID string) error {
	return c.do(ctx, "POST", "/v1/checkout/sessions/"+sessionID+"/expire", nil, nil)
}
