package fintech

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type WalletService struct{ c *Client }

// ── Enums ─────────────────────────────────────────────────────────────

type WalletStatus string

const (
	WalletStatusActive WalletStatus = "WALLET_STATUS_ACTIVE"
	WalletStatusFrozen WalletStatus = "WALLET_STATUS_FROZEN"
	WalletStatusClosed WalletStatus = "WALLET_STATUS_CLOSED"
)

type WalletType string

const (
	WalletTypeFiat    WalletType = "WALLET_TYPE_FIAT"
	WalletTypeCrypto  WalletType = "WALLET_TYPE_CRYPTO"
	WalletTypeOmnibus WalletType = "WALLET_TYPE_OMNIBUS"
)

type TransactionType string

const (
	TransactionCredit TransactionType = "TRANSACTION_TYPE_CREDIT"
	TransactionDebit  TransactionType = "TRANSACTION_TYPE_DEBIT"
)

type TransactionStatus string

const (
	TxStatusPending  TransactionStatus = "TRANSACTION_STATUS_PENDING"
	TxStatusSettled  TransactionStatus = "TRANSACTION_STATUS_SETTLED"
	TxStatusFailed   TransactionStatus = "TRANSACTION_STATUS_FAILED"
	TxStatusReversed TransactionStatus = "TRANSACTION_STATUS_REVERSED"
)

type TransactionReason string

const (
	ReasonDeposit    TransactionReason = "TRANSACTION_REASON_DEPOSIT"
	ReasonWithdrawal TransactionReason = "TRANSACTION_REASON_WITHDRAWAL"
	ReasonPayment    TransactionReason = "TRANSACTION_REASON_PAYMENT"
	ReasonRefund     TransactionReason = "TRANSACTION_REASON_REFUND"
	ReasonSettlement TransactionReason = "TRANSACTION_REASON_SETTLEMENT"
	ReasonFee        TransactionReason = "TRANSACTION_REASON_FEE"
	ReasonAdjustment TransactionReason = "TRANSACTION_REASON_ADJUSTMENT"
	ReasonConversion TransactionReason = "TRANSACTION_REASON_CONVERSION"
)

// ── Messages ──────────────────────────────────────────────────────────

// Wallet — money fields are minor units as string to avoid JS precision
// loss when the same JSON crosses the Kiro frontend boundary.
type Wallet struct {
	WalletID     string       `json:"wallet_id"`
	TenantID     string       `json:"tenant_id"`
	SubjectID    string       `json:"subject_id"`
	SubjectType  string       `json:"subject_type"` // "user" | "merchant"
	WalletType   WalletType   `json:"wallet_type"`
	Status       WalletStatus `json:"status"`
	Currency     string       `json:"currency"`
	Balance      string       `json:"balance"`
	Pending      string       `json:"pending"`
	FreezeReason string       `json:"freeze_reason"`
	FrozenBy     string       `json:"frozen_by"`
	CreatedAt    string       `json:"created_at"`
	UpdatedAt    string       `json:"updated_at"`
	FrozenAt     *string      `json:"frozen_at,omitempty"`
}

type WalletTransaction struct {
	TransactionID string            `json:"transaction_id"`
	WalletID      string            `json:"wallet_id"`
	TenantID      string            `json:"tenant_id"`
	Type          TransactionType   `json:"type"`
	Status        TransactionStatus `json:"status"`
	Reason        TransactionReason `json:"reason"`
	Amount        string            `json:"amount"`
	Currency      string            `json:"currency"`
	BalanceAfter  string            `json:"balance_after"`
	ReferenceID   string            `json:"reference_id"`
	ReferenceType string            `json:"reference_type"`
	Description   string            `json:"description"`
	ExternalRef   string            `json:"external_ref"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	CreatedAt     string            `json:"created_at"`
	SettledAt     *string           `json:"settled_at,omitempty"`
}

type DepositAddress struct {
	WalletID  string  `json:"wallet_id"`
	Currency  string  `json:"currency"`
	Network   string  `json:"network"`
	Address   string  `json:"address"`
	Memo      *string `json:"memo,omitempty"`
	ExpiresAt *string `json:"expires_at,omitempty"`
}

// Request types

type CreateWalletRequest struct {
	TenantID       string     `json:"tenant_id"`
	SubjectID      string     `json:"subject_id"`
	SubjectType    string     `json:"subject_type"`
	WalletType     WalletType `json:"wallet_type"`
	Currency       string     `json:"currency"`
	IdempotencyKey string     `json:"idempotency_key,omitempty"`
}

type CreditRequest struct {
	TenantID       string            `json:"tenant_id"`
	WalletID       string            `json:"wallet_id"`
	Amount         string            `json:"amount"`
	Reason         TransactionReason `json:"reason"`
	ReferenceID    string            `json:"reference_id"`
	ReferenceType  string            `json:"reference_type"`
	Description    string            `json:"description"`
	IdempotencyKey string            `json:"idempotency_key,omitempty"`
}

type DebitRequest = CreditRequest

type WithdrawRequest struct {
	TenantID        string `json:"tenant_id"`
	WalletID        string `json:"wallet_id"`
	Amount          string `json:"amount"`
	DestinationType string `json:"destination_type"`
	DestinationID  string `json:"destination_id"`
	Network        string `json:"network,omitempty"`
	Description    string `json:"description,omitempty"`
	IdempotencyKey string `json:"idempotency_key,omitempty"`
}

type ListTransactionsResponse struct {
	Transactions []WalletTransaction `json:"transactions"`
	NextCursor   *string             `json:"next_cursor,omitempty"`
}

type ListWalletsResponse struct {
	Wallets    []Wallet `json:"wallets"`
	NextCursor *string  `json:"next_cursor,omitempty"`
}

// ── RPCs ──────────────────────────────────────────────────────────────

func (s *WalletService) Create(ctx context.Context, req CreateWalletRequest, idempotencyKey string) (*Wallet, error) {
	req.TenantID = s.c.tenantID
	var out Wallet
	if err := s.c.do(ctx, http.MethodPost, "/v1/wallets", req, idempotencyKey, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *WalletService) Get(ctx context.Context, walletID string) (*Wallet, error) {
	path := fmt.Sprintf("/v1/wallets/%s?tenant_id=%s", walletID, s.c.tenantID)
	var out Wallet
	if err := s.c.do(ctx, http.MethodGet, path, nil, "", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *WalletService) List(ctx context.Context, subjectID string, limit int, cursor string) (*ListWalletsResponse, error) {
	qs := url.Values{}
	qs.Set("tenant_id", s.c.tenantID)
	if subjectID != "" {
		qs.Set("subject_id", subjectID)
	}
	if limit > 0 {
		qs.Set("limit", fmt.Sprintf("%d", limit))
	}
	if cursor != "" {
		qs.Set("cursor", cursor)
	}
	var out ListWalletsResponse
	if err := s.c.do(ctx, http.MethodGet, "/v1/wallets?"+qs.Encode(), nil, "", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *WalletService) Credit(ctx context.Context, req CreditRequest, idempotencyKey string) (*WalletTransaction, error) {
	req.TenantID = s.c.tenantID
	path := fmt.Sprintf("/v1/wallets/%s/credit", req.WalletID)
	var out WalletTransaction
	if err := s.c.do(ctx, http.MethodPost, path, req, idempotencyKey, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *WalletService) Debit(ctx context.Context, req DebitRequest, idempotencyKey string) (*WalletTransaction, error) {
	req.TenantID = s.c.tenantID
	path := fmt.Sprintf("/v1/wallets/%s/debit", req.WalletID)
	var out WalletTransaction
	if err := s.c.do(ctx, http.MethodPost, path, req, idempotencyKey, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *WalletService) Withdraw(ctx context.Context, req WithdrawRequest, idempotencyKey string) (*WalletTransaction, error) {
	req.TenantID = s.c.tenantID
	path := fmt.Sprintf("/v1/wallets/%s/withdraw", req.WalletID)
	var out WalletTransaction
	if err := s.c.do(ctx, http.MethodPost, path, req, idempotencyKey, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *WalletService) ListTransactions(ctx context.Context, walletID string, limit int, cursor string) (*ListTransactionsResponse, error) {
	qs := url.Values{}
	qs.Set("tenant_id", s.c.tenantID)
	if limit > 0 {
		qs.Set("limit", fmt.Sprintf("%d", limit))
	}
	if cursor != "" {
		qs.Set("cursor", cursor)
	}
	path := fmt.Sprintf("/v1/wallets/%s/transactions?%s", walletID, qs.Encode())
	var out ListTransactionsResponse
	if err := s.c.do(ctx, http.MethodGet, path, nil, "", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *WalletService) DepositAddress(ctx context.Context, walletID, network string) (*DepositAddress, error) {
	path := fmt.Sprintf("/v1/wallets/%s/deposit-address?tenant_id=%s&network=%s", walletID, s.c.tenantID, network)
	var out DepositAddress
	if err := s.c.do(ctx, http.MethodGet, path, nil, "", &out); err != nil {
		return nil, err
	}
	return &out, nil
}
