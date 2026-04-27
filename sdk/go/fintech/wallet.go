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
	Currency     Currency     `json:"currency"`
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
	Currency      Currency          `json:"currency"`
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
	WalletID  string   `json:"wallet_id"`
	Currency  Currency `json:"currency"`
	Network   Network  `json:"network"`
	Address   string   `json:"address"`
	Memo      *string  `json:"memo,omitempty"`
	ExpiresAt *string  `json:"expires_at,omitempty"`
}

// Request types

type CreateWalletRequest struct {
	TenantID       string     `json:"tenant_id"`
	SubjectID      string     `json:"subject_id"`
	SubjectType    string     `json:"subject_type"`
	WalletType     WalletType `json:"wallet_type"`
	Currency       Currency   `json:"currency"`
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
	TenantID        string  `json:"tenant_id"`
	WalletID        string  `json:"wallet_id"`
	Amount          string  `json:"amount"`
	DestinationType string  `json:"destination_type"`
	DestinationID   string  `json:"destination_id"`
	Network         Network `json:"network,omitempty"`
	Description     string  `json:"description,omitempty"`
	IdempotencyKey  string  `json:"idempotency_key,omitempty"`
	// SPL token mint (base58). Empty = native SOL. Required for SPL token
	// withdrawals; ignored for bank_account destinations. L2 of ADR-0016.
	Mint Mint `json:"mint,omitempty"`
}

// CreateChainWalletRequest creates a non-custodial on-chain wallet — BIP-39
// mnemonic is generated server-side and returned ONCE in the response.
// Caller MUST surface the mnemonic to the end user and never persist it.
// Currently SOLANA only.
type CreateChainWalletRequest struct {
	TenantID       string   `json:"tenant_id"`
	SubjectID      string   `json:"subject_id"`
	SubjectType    string   `json:"subject_type"` // "user" | "merchant"
	Currency       Currency `json:"currency"`
	Network        Network  `json:"network"`
	IdempotencyKey string   `json:"idempotency_key,omitempty"`
}

type CreateChainWalletResponse struct {
	Wallet   Wallet `json:"wallet"`
	Mnemonic string `json:"mnemonic"` // shown to user once; do NOT persist
}

type FreezeWalletRequest struct {
	TenantID     string `json:"tenant_id"`
	WalletID     string `json:"wallet_id"`
	FreezeReason string `json:"freeze_reason"`
}

type UnfreezeWalletRequest struct {
	TenantID string `json:"tenant_id"`
	WalletID string `json:"wallet_id"`
	Note     string `json:"note,omitempty"`
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

// DepositAddress returns the on-chain deposit target for a wallet. Pass
// a non-empty `mint` (SPL token mint) on Solana to get the Associated
// Token Account derived from (wallet_owner, mint); pass empty `mint` for
// the native asset (SOL etc.) — the wallet owner address is returned.
// L3 of ADR-0016.
func (s *WalletService) DepositAddress(ctx context.Context, walletID string, network Network, mint Mint) (*DepositAddress, error) {
	qs := url.Values{}
	qs.Set("tenant_id", s.c.tenantID)
	qs.Set("network", string(network))
	if mint != "" {
		qs.Set("mint", string(mint))
	}
	path := fmt.Sprintf("/v1/wallets/%s/deposit-address?%s", walletID, qs.Encode())
	var out DepositAddress
	if err := s.c.do(ctx, http.MethodGet, path, nil, "", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateChain generates a non-custodial on-chain wallet (BIP-39 mnemonic +
// SLIP-0010 derivation, AES-encrypted private key stored server-side). The
// returned mnemonic is shown to the end user ONCE — caller MUST display it
// and never persist it; the server keeps only a SHA-256 hash for duplicate
// detection. Currently SOLANA only. L1 of ADR-0016.
func (s *WalletService) CreateChain(ctx context.Context, req CreateChainWalletRequest, idempotencyKey string) (*CreateChainWalletResponse, error) {
	req.TenantID = s.c.tenantID
	var out CreateChainWalletResponse
	if err := s.c.do(ctx, http.MethodPost, "/v1/wallets/chain", req, idempotencyKey, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Freeze halts all debits on a wallet (compliance / fraud trigger).
// `reason` is required and recorded in the wallet for audit.
func (s *WalletService) Freeze(ctx context.Context, walletID, reason string) (*Wallet, error) {
	req := FreezeWalletRequest{
		TenantID:     s.c.tenantID,
		WalletID:     walletID,
		FreezeReason: reason,
	}
	path := fmt.Sprintf("/v1/wallets/%s/freeze", walletID)
	var out Wallet
	if err := s.c.do(ctx, http.MethodPost, path, req, "", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Unfreeze restores a frozen wallet. `note` is optional — recorded for audit.
func (s *WalletService) Unfreeze(ctx context.Context, walletID, note string) (*Wallet, error) {
	req := UnfreezeWalletRequest{
		TenantID: s.c.tenantID,
		WalletID: walletID,
		Note:     note,
	}
	path := fmt.Sprintf("/v1/wallets/%s/unfreeze", walletID)
	var out Wallet
	if err := s.c.do(ctx, http.MethodPost, path, req, "", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetTransaction fetches a single wallet transaction by ID.
func (s *WalletService) GetTransaction(ctx context.Context, walletID, transactionID string) (*WalletTransaction, error) {
	path := fmt.Sprintf("/v1/wallets/%s/transactions/%s?tenant_id=%s", walletID, transactionID, s.c.tenantID)
	var out WalletTransaction
	if err := s.c.do(ctx, http.MethodGet, path, nil, "", &out); err != nil {
		return nil, err
	}
	return &out, nil
}
