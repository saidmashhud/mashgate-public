package mashgate

import "context"

// ────────────────────────────────────────────────────────────────────────────
// Types
// ────────────────────────────────────────────────────────────────────────────

// WalletBalance is the ledger-derived balance view for a user or merchant.
// Available balance = captured payments credited to the account.
// Holds balance = authorized but not yet captured.
type WalletBalance struct {
	TenantID         string `json:"tenantId"`
	UserID           string `json:"userId"`
	Currency         string `json:"currency"`
	AvailableBalance string `json:"availableBalance"` // decimal string
	HoldsBalance     string `json:"holdsBalance"`     // decimal string
	UpdatedAt        int64  `json:"updatedAt"`
}

// SavedPaymentMethod is a tokenized card stored in a user's wallet for
// repeat use at checkout.
type SavedPaymentMethod struct {
	PaymentMethodID string `json:"paymentMethodId"`
	Brand           string `json:"brand"`    // "uzcard" | "humo" | "visa" | "mastercard" etc.
	Last4           string `json:"last4"`
	ExpMonth        int    `json:"expMonth"`
	ExpYear         int    `json:"expYear"`
	IsDefault       bool   `json:"isDefault"`
	CreatedAt       int64  `json:"createdAt"`
}

type listSavedPaymentMethodsResponse struct {
	PaymentMethods []*SavedPaymentMethod `json:"paymentMethods"`
}

// ────────────────────────────────────────────────────────────────────────────
// Wallet methods
// ────────────────────────────────────────────────────────────────────────────

// GetWalletBalance returns the current balance for the authenticated user.
// currency is an ISO 4217 code, e.g. "UZS", "KZT", "USD".
func (c *Client) GetWalletBalance(ctx context.Context, currency string) (*WalletBalance, error) {
	path := "/v1/wallet/balance"
	if currency != "" {
		path += "?currency=" + currency
	}
	var out WalletBalance
	if err := c.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListSavedPaymentMethods returns all saved cards for the authenticated user.
func (c *Client) ListSavedPaymentMethods(ctx context.Context) ([]*SavedPaymentMethod, error) {
	var out listSavedPaymentMethodsResponse
	if err := c.do(ctx, "GET", "/v1/wallet/payment-methods", nil, &out); err != nil {
		return nil, err
	}
	return out.PaymentMethods, nil
}

// RemoveSavedPaymentMethod deletes a saved card by ID.
func (c *Client) RemoveSavedPaymentMethod(ctx context.Context, paymentMethodID string) error {
	return c.do(ctx, "DELETE", "/v1/wallet/payment-methods/"+paymentMethodID, nil, nil)
}

// SetDefaultPaymentMethod designates a saved card as the default for checkout.
func (c *Client) SetDefaultPaymentMethod(ctx context.Context, paymentMethodID string) error {
	return c.do(ctx, "POST", "/v1/wallet/payment-methods/"+paymentMethodID+"/default", nil, nil)
}
