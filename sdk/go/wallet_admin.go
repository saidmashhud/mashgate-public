package mashgate

import (
	"context"
	"fmt"
	"net/url"
)

// ────────────────────────────────────────────────────────────────────────────
// WalletAdminClient — privileged wallet operations. Mirrors admin/* RPCs of
// WalletService (wallet.proto) requiring `wallet:admin:*` permissions.
//
// REGULAR users use top-level Client.GetWalletBalance / Client.ListSavedPaymentMethods.
// Admin is for platform_admin / customer-support role with explicit RBAC grant.
//
// AUDIT: all admin operations write to platform_audit_log with operator id,
// reason, before/after state. See ADR-0010 (Permissions model).
// ────────────────────────────────────────────────────────────────────────────

type WalletAdminClient struct {
	c *Client
}

// AdminListWallets returns all wallets in a tenant (with optional status filter).
// Reqs: `wallet:admin:read`.
func (w *WalletAdminClient) AdminListWallets(ctx context.Context, tenantID, status string, page, pageSize int) ([]*AdminWallet, error) {
	q := fmt.Sprintf("?tenantId=%s", url.QueryEscape(tenantID))
	if status != "" {
		q += "&status=" + url.QueryEscape(status)
	}
	if page > 0 {
		q += fmt.Sprintf("&page=%d", page)
	}
	if pageSize > 0 {
		q += fmt.Sprintf("&pageSize=%d", pageSize)
	}
	var out struct {
		Wallets []*AdminWallet `json:"wallets"`
	}
	if err := w.c.do(ctx, "GET", "/v1/admin/wallets"+q, nil, &out); err != nil {
		return nil, err
	}
	return out.Wallets, nil
}

// AdminGetWallet retrieves full wallet incl. balance, kyc state, recent activity.
// Reqs: `wallet:admin:read`.
func (w *WalletAdminClient) AdminGetWallet(ctx context.Context, walletID string) (*AdminWallet, error) {
	var out AdminWallet
	if err := w.c.do(ctx, "GET", "/v1/admin/wallets/"+walletID, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AdminFreezeWallet temporarily blocks transactions on a wallet.
// Authentic reasons: "kyc_review", "fraud_suspicion", "compliance_hold", "user_request".
// Reqs: `wallet:admin:freeze`.
func (w *WalletAdminClient) AdminFreezeWallet(ctx context.Context, walletID string, req FreezeWalletRequest) (*AdminWallet, error) {
	var out AdminWallet
	if err := w.c.do(ctx, "POST", "/v1/admin/wallets/"+walletID+"/freeze", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AdminUnfreezeWallet restores transactions. Requires resolved_reason note.
// Reqs: `wallet:admin:freeze`.
func (w *WalletAdminClient) AdminUnfreezeWallet(ctx context.Context, walletID string, req UnfreezeWalletRequest) (*AdminWallet, error) {
	var out AdminWallet
	if err := w.c.do(ctx, "POST", "/v1/admin/wallets/"+walletID+"/unfreeze", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AdminAdjustBalance — manual balance adjustment (correction, refund, promo credit).
// Writes audit + ledger entry. Reqs: `wallet:admin:adjust-balance` (privileged).
//
// AmountCents can be negative (debit). Reason is required and surfaces in audit.
func (w *WalletAdminClient) AdminAdjustBalance(ctx context.Context, walletID string, req AdjustBalanceRequest) (*WalletAdjustment, error) {
	var out WalletAdjustment
	if err := w.c.do(ctx, "POST", "/v1/admin/wallets/"+walletID+"/adjust-balance", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AdminCloseWallet permanently deactivates a wallet. Cannot be undone.
// Reqs: `wallet:admin:close`. Balance must be 0 OR explicit force=true.
func (w *WalletAdminClient) AdminCloseWallet(ctx context.Context, walletID, reason string, force bool) error {
	req := struct {
		Reason string `json:"reason"`
		Force  bool   `json:"force,omitempty"`
	}{Reason: reason, Force: force}
	return w.c.do(ctx, "POST", "/v1/admin/wallets/"+walletID+"/close", req, nil)
}

// AdminAuditLog returns audit history for a wallet (all admin actions taken).
// Reqs: `wallet:admin:read`.
func (w *WalletAdminClient) AdminAuditLog(ctx context.Context, walletID string, page, pageSize int) ([]*WalletAuditEntry, error) {
	q := ""
	if page > 0 {
		q += fmt.Sprintf("?page=%d", page)
	}
	if pageSize > 0 {
		sep := "?"
		if q != "" {
			sep = "&"
		}
		q += fmt.Sprintf("%spageSize=%d", sep, pageSize)
	}
	var out struct {
		Entries []*WalletAuditEntry `json:"entries"`
	}
	if err := w.c.do(ctx, "GET", "/v1/admin/wallets/"+walletID+"/audit"+q, nil, &out); err != nil {
		return nil, err
	}
	return out.Entries, nil
}
