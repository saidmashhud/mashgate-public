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
	WalletID     string       `json:"walletId"`
	TenantID     string       `json:"tenantId"`
	SubjectID    string       `json:"subjectId"`
	SubjectType  string       `json:"subjectType"` // "user" | "merchant"
	WalletType   WalletType   `json:"walletType"`
	Status       WalletStatus `json:"status"`
	Currency     Currency     `json:"currency"`
	Balance      string       `json:"balance"`
	Pending      string       `json:"pending"`
	FreezeReason string       `json:"freezeReason"`
	FrozenBy     string       `json:"frozenBy"`
	CreatedAt    string       `json:"createdAt"`
	UpdatedAt    string       `json:"updatedAt"`
	FrozenAt     *string      `json:"frozenAt,omitempty"`
}

type WalletTransaction struct {
	TransactionID string            `json:"transactionId"`
	WalletID      string            `json:"walletId"`
	TenantID      string            `json:"tenantId"`
	Type          TransactionType   `json:"type"`
	Status        TransactionStatus `json:"status"`
	Reason        TransactionReason `json:"reason"`
	Amount        string            `json:"amount"`
	Currency      Currency          `json:"currency"`
	BalanceAfter  string            `json:"balanceAfter"`
	ReferenceID   string            `json:"referenceId"`
	ReferenceType string            `json:"referenceType"`
	Description   string            `json:"description"`
	ExternalRef   string            `json:"externalRef"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	CreatedAt     string            `json:"createdAt"`
	SettledAt     *string           `json:"settledAt,omitempty"`
}

type DepositAddress struct {
	WalletID  string   `json:"walletId"`
	Currency  Currency `json:"currency"`
	Network   Network  `json:"network"`
	Address   string   `json:"address"`
	Memo      *string  `json:"memo,omitempty"`
	ExpiresAt *string  `json:"expiresAt,omitempty"`
}

// Request types

type CreateWalletRequest struct {
	TenantID       string     `json:"tenantId"`
	SubjectID      string     `json:"subjectId"`
	SubjectType    string     `json:"subjectType"`
	WalletType     WalletType `json:"walletType"`
	Currency       Currency   `json:"currency"`
	IdempotencyKey string     `json:"idempotencyKey,omitempty"`
}

type CreditRequest struct {
	TenantID       string            `json:"tenantId"`
	WalletID       string            `json:"walletId"`
	Amount         string            `json:"amount"`
	Reason         TransactionReason `json:"reason"`
	ReferenceID    string            `json:"referenceId"`
	ReferenceType  string            `json:"referenceType"`
	Description    string            `json:"description"`
	IdempotencyKey string            `json:"idempotencyKey,omitempty"`
}

type DebitRequest = CreditRequest

// TransferRequest moves funds atomically between two wallets in the same
// tenant. Same-currency only in v1 (cross-currency FX out of scope).
// The server enforces: both wallets exist, share the tenant, share the
// currency, neither is frozen, source has sufficient balance. Backed by
// a single Postgres transaction — either the entire transfer commits
// (both balances updated + two wallet_transactions rows + three outbox
// events: wallet.debit, wallet.credit, wallet.transfer) or none of it.
type TransferRequest struct {
	TenantID       string            `json:"tenantId"`
	FromWalletID   string            `json:"fromWalletId"`
	ToWalletID     string            `json:"toWalletId"`
	Amount         string            `json:"amount"`
	Reason         TransactionReason `json:"reason,omitempty"`
	Description    string            `json:"description,omitempty"`
	IdempotencyKey string            `json:"idempotencyKey,omitempty"`
	// MerchantID is included in the paired wallet.debit / wallet.credit
	// envelope events. Required if those events must be contract-valid
	// per contracts/events/wallet.{credit,debit}.json. Empty = movement
	// events are emitted without merchant_id and dropped by consumers
	// that require it (money state still commits).
	MerchantID string `json:"merchantId,omitempty"`
	// Note is optional free-text attached to the wallet.transfer envelope.
	Note string `json:"note,omitempty"`
}

// TransferResponse carries the synthetic transfer_id plus both
// wallet_transactions rows (debit on source, credit on destination).
// transfer_id correlates with `transfer_id` in the wallet.transfer
// outbox event and with `reference_id` on both wallet_transactions rows.
type TransferResponse struct {
	TransferID string            `json:"transferId"`
	Debit      WalletTransaction `json:"debit"`
	Credit     WalletTransaction `json:"credit"`
}

type WithdrawRequest struct {
	TenantID        string  `json:"tenantId"`
	WalletID        string  `json:"walletId"`
	Amount          string  `json:"amount"`
	DestinationType string  `json:"destinationType"`
	DestinationID   string  `json:"destinationId"`
	Network         Network `json:"network,omitempty"`
	Description     string  `json:"description,omitempty"`
	IdempotencyKey  string  `json:"idempotencyKey,omitempty"`
	// SPL token mint (base58). Empty = native SOL. Required for SPL token
	// withdrawals; ignored for bank_account destinations. L2 of ADR-0016.
	Mint Mint `json:"mint,omitempty"`
	// Optional sponsor wallet — when set, the platform sponsor wallet pays
	// the chain fee + ATA rent instead of the source. Used for gasless
	// withdrawals: a customer holding USDT but zero SOL can still move
	// tokens off-chain. Sponsor must be in the same tenant + same network,
	// active, and on-chain.
	SponsorWalletID string `json:"sponsorWalletId,omitempty"`
}

// CreateChainWalletRequest creates a non-custodial on-chain wallet — BIP-39
// mnemonic is generated server-side and returned ONCE in the response.
// Caller MUST surface the mnemonic to the end user and never persist it.
// Currently SOLANA only.
type CreateChainWalletRequest struct {
	TenantID       string   `json:"tenantId"`
	SubjectID      string   `json:"subjectId"`
	SubjectType    string   `json:"subjectType"` // "user" | "merchant"
	Currency       Currency `json:"currency"`
	Network        Network  `json:"network"`
	IdempotencyKey string   `json:"idempotencyKey,omitempty"`
}

type CreateChainWalletResponse struct {
	Wallet   Wallet `json:"wallet"`
	Mnemonic string `json:"mnemonic"` // shown to user once; do NOT persist
}

// ImportChainWalletRequest imports an existing non-custodial wallet from a
// caller-provided BIP-39 mnemonic. Server validates, derives address
// (SLIP-0010), AES-encrypts the private key, and stores it. Idempotent on
// (tenant, mnemonic_hash) — re-importing the same phrase for the same
// subject returns the existing wallet с WasExisting=true. Importing under a
// different subject within the tenant returns 403 PERMISSION_DENIED
// (credential-stuffing protection). Mnemonic never logged / never
// persisted plaintext.
type ImportChainWalletRequest struct {
	TenantID       string   `json:"tenantId"`
	SubjectID      string   `json:"subjectId"`
	SubjectType    string   `json:"subjectType"` // "user" | "merchant"
	Currency       Currency `json:"currency"`
	Network        Network  `json:"network"`
	Mnemonic       string   `json:"mnemonic"`
	IdempotencyKey string   `json:"idempotencyKey,omitempty"`
}

// ImportChainWalletResponse carries the resolved wallet plus a flag that
// distinguishes a fresh import from a recovery (same seed re-imported for
// the same subject).
type ImportChainWalletResponse struct {
	Wallet Wallet `json:"wallet"`
	// True when the import resolved to a pre-existing row — a recovery
	// rather than a new wallet. Frontends use this to render
	// "✓ wallet recovered" vs "✓ wallet imported".
	WasExisting bool `json:"wasExisting"`
	// RFC-3339 timestamp echoing wallet.created_at on fresh imports or
	// wallet.updated_at on recoveries. Optional — empty on older servers.
	RecoveredAt string `json:"recoveredAt,omitempty"`
}

type FreezeWalletRequest struct {
	TenantID     string `json:"tenantId"`
	WalletID     string `json:"walletId"`
	FreezeReason string `json:"freezeReason"`
}

type UnfreezeWalletRequest struct {
	TenantID string `json:"tenantId"`
	WalletID string `json:"walletId"`
	Note     string `json:"note,omitempty"`
}

type ListTransactionsResponse struct {
	Transactions []WalletTransaction `json:"transactions"`
	NextCursor   *string             `json:"nextCursor,omitempty"`
}

type ListWalletsResponse struct {
	Wallets    []Wallet `json:"wallets"`
	NextCursor *string  `json:"nextCursor,omitempty"`
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

// ImportChain restores a non-custodial wallet from a user-provided BIP-39
// mnemonic. Returns the wallet plus a WasExisting flag that distinguishes a
// fresh import from a recovery of the same phrase by the same subject.
// Idempotent on (tenant, mnemonic_hash). Currently SOLANA only.
//
// **Caller MUST clear the mnemonic from memory after this call** — the
// mnemonic touches process memory of both caller and server briefly; it is
// never stored plaintext server-side (SHA-256 hash for dedup only).
//
// Errors mapped from gRPC status:
//   - INVALID_ARGUMENT — mnemonic fails BIP-39 checksum, network/currency missing.
//   - PERMISSION_DENIED — same mnemonic_hash already belongs to a different
//     subject in this tenant (credential-stuffing protection).
//   - FAILED_PRECONDITION — WALLET_ENCRYPTION_KEY not configured server-side.
//
// BREAKING CHANGE vs sdk/go/v1.8.0: return type changed from *Wallet to
// *ImportChainWalletResponse. Callers using the wallet directly need
// `resp.Wallet` instead of `resp`.
func (s *WalletService) ImportChain(ctx context.Context, req ImportChainWalletRequest, idempotencyKey string) (*ImportChainWalletResponse, error) {
	req.TenantID = s.c.tenantID
	var out ImportChainWalletResponse
	if err := s.c.do(ctx, http.MethodPost, "/v1/wallets/chain/import", req, idempotencyKey, &out); err != nil {
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

// Transfer moves `amount` from `req.FromWalletID` to `req.ToWalletID`
// atomically. Both wallets must belong to the same tenant and share the
// same currency. Idempotent via `req.IdempotencyKey`. Returns the
// synthetic transfer_id plus both wallet_transactions rows.
//
// Errors mapped from gRPC status:
//   - INVALID_ARGUMENT — same wallet IDs, currency mismatch, non-positive amount.
//   - FAILED_PRECONDITION — source/destination frozen, insufficient balance.
//   - PERMISSION_DENIED — wallets belong to different tenants.
//   - NOT_FOUND — wallet does not exist.
func (s *WalletService) Transfer(ctx context.Context, req TransferRequest, idempotencyKey string) (*TransferResponse, error) {
	req.TenantID = s.c.tenantID
	if idempotencyKey != "" && req.IdempotencyKey == "" {
		// Mirror the header into the body so the server picks up either
		// form (existing wallet methods rely on the header alone, but
		// transfer namespaces the key per leg internally — surface both
		// to keep callers' lives easy).
		req.IdempotencyKey = idempotencyKey
	}
	path := fmt.Sprintf("/v1/wallets/%s/transfer", req.FromWalletID)
	var out TransferResponse
	if err := s.c.do(ctx, http.MethodPost, path, req, idempotencyKey, &out); err != nil {
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
