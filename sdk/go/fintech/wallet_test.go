package fintech

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// helper: spin up a mock Mashgate gateway, asserting `wantMethod` + `wantPath`,
// returning `respBody` (JSON) with status 200. The handler also captures the
// last request body so tests can assert on it.
type capture struct {
	body            []byte
	authHeader      string
	tenantHeader    string
	idempotencyKey  string
	contentType     string
	rawQuery        string
	method          string
	path            string
	traceparent     string
}

func mockServer(t *testing.T, status int, respBody string, cap *capture) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		cap.body = body
		cap.authHeader = r.Header.Get("Authorization")
		cap.tenantHeader = r.Header.Get("X-Tenant-ID")
		cap.idempotencyKey = r.Header.Get("Idempotency-Key")
		cap.contentType = r.Header.Get("Content-Type")
		cap.traceparent = r.Header.Get("traceparent")
		cap.rawQuery = r.URL.RawQuery
		cap.method = r.Method
		cap.path = r.URL.Path
		w.WriteHeader(status)
		_, _ = w.Write([]byte(respBody))
	}))
}

func TestWalletService_CreateChain_SendsExpectedShape(t *testing.T) {
	cap := &capture{}
	respWallet := Wallet{WalletID: "w-1", Currency: "USDC", Status: WalletStatusActive}
	respJSON := `{"wallet":` + mustJSON(t, respWallet) + `,"mnemonic":"abandon ability ..."}`
	srv := mockServer(t, http.StatusOK, respJSON, cap)
	defer srv.Close()

	c := New(srv.URL, "tenant-A", "key-xyz")
	out, err := c.Wallet.CreateChain(context.Background(), CreateChainWalletRequest{
		SubjectID:   "user-1",
		SubjectType: "user",
		Currency:    "USDC",
		Network:     "SOLANA",
	}, "idem-create-1")
	if err != nil {
		t.Fatalf("CreateChain returned error: %v", err)
	}
	if out.Mnemonic == "" {
		t.Errorf("expected mnemonic in response, got empty")
	}
	if out.Wallet.WalletID != "w-1" {
		t.Errorf("expected wallet_id=w-1, got %q", out.Wallet.WalletID)
	}

	// Wire shape
	if cap.method != http.MethodPost || cap.path != "/v1/wallets/chain" {
		t.Errorf("expected POST /v1/wallets/chain, got %s %s", cap.method, cap.path)
	}
	if cap.idempotencyKey != "idem-create-1" {
		t.Errorf("expected Idempotency-Key=idem-create-1, got %q", cap.idempotencyKey)
	}
	if cap.tenantHeader != "tenant-A" {
		t.Errorf("expected X-Tenant-ID=tenant-A, got %q", cap.tenantHeader)
	}

	// Body should carry tenant_id from client config, not whatever caller passed.
	var sent CreateChainWalletRequest
	if err := json.Unmarshal(cap.body, &sent); err != nil {
		t.Fatalf("unmarshal sent body: %v", err)
	}
	if sent.TenantID != "tenant-A" {
		t.Errorf("expected tenant_id=tenant-A in body, got %q", sent.TenantID)
	}
	if sent.Network != "SOLANA" || sent.Currency != "USDC" {
		t.Errorf("unexpected payload: %+v", sent)
	}
}

func TestWalletService_DepositAddress_PassesMint(t *testing.T) {
	// L3 — when a mint is supplied, gateway routes to ledger-core which
	// derives the SPL Associated Token Account.
	cap := &capture{}
	respJSON := `{"wallet_id":"w-1","currency":"USDC","network":"solana","address":"AtaAddrBase58Here"}`
	srv := mockServer(t, http.StatusOK, respJSON, cap)
	defer srv.Close()

	c := New(srv.URL, "tenant-A", "key-xyz")
	out, err := c.Wallet.DepositAddress(
		context.Background(),
		"w-1",
		"SOLANA",
		"EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
	)
	if err != nil {
		t.Fatalf("DepositAddress error: %v", err)
	}
	if out.Address != "AtaAddrBase58Here" {
		t.Errorf("unexpected address: %q", out.Address)
	}

	if !strings.Contains(cap.rawQuery, "mint=EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v") {
		t.Errorf("expected mint param in query, got %q", cap.rawQuery)
	}
	if !strings.Contains(cap.rawQuery, "network=SOLANA") {
		t.Errorf("expected network=SOLANA in query, got %q", cap.rawQuery)
	}
	if cap.path != "/v1/wallets/w-1/deposit-address" {
		t.Errorf("unexpected path: %q", cap.path)
	}
}

func TestWalletService_DepositAddress_OmitsEmptyMint(t *testing.T) {
	// Native SOL deposit — empty mint should NOT appear in query string.
	cap := &capture{}
	srv := mockServer(t, http.StatusOK, `{"address":"OwnerPubkey"}`, cap)
	defer srv.Close()

	c := New(srv.URL, "tenant-A", "key-xyz")
	_, err := c.Wallet.DepositAddress(context.Background(), "w-1", "SOLANA", "")
	if err != nil {
		t.Fatalf("DepositAddress error: %v", err)
	}
	if strings.Contains(cap.rawQuery, "mint=") {
		t.Errorf("expected NO mint param in query for native asset, got %q", cap.rawQuery)
	}
}

func TestWalletService_Withdraw_IncludesMint(t *testing.T) {
	// L2 — explicit mint field replaces the older `mint=...;` description hack.
	cap := &capture{}
	respJSON := `{"transaction_id":"tx-1","wallet_id":"w-1","status":"TRANSACTION_STATUS_PENDING"}`
	srv := mockServer(t, http.StatusOK, respJSON, cap)
	defer srv.Close()

	c := New(srv.URL, "tenant-A", "key-xyz")
	_, err := c.Wallet.Withdraw(context.Background(), WithdrawRequest{
		WalletID:        "w-1",
		Amount:          "10.5",
		DestinationType: "crypto_address",
		DestinationID:   "DestSolanaAddr",
		Network:         "SOLANA",
		Mint:            "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
	}, "idem-w-1")
	if err != nil {
		t.Fatalf("Withdraw error: %v", err)
	}

	var sent WithdrawRequest
	if err := json.Unmarshal(cap.body, &sent); err != nil {
		t.Fatalf("unmarshal sent body: %v", err)
	}
	if sent.Mint != "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v" {
		t.Errorf("expected mint in body, got %q", sent.Mint)
	}
	if sent.Description != "" {
		t.Errorf("description should not be touched, got %q", sent.Description)
	}
}

func TestWalletService_FreezeAndUnfreeze(t *testing.T) {
	cap := &capture{}
	srv := mockServer(t, http.StatusOK, `{"wallet_id":"w-1","status":"WALLET_STATUS_FROZEN"}`, cap)
	defer srv.Close()

	c := New(srv.URL, "tenant-A", "key-xyz")
	w, err := c.Wallet.Freeze(context.Background(), "w-1", "fraud-investigation")
	if err != nil {
		t.Fatalf("Freeze error: %v", err)
	}
	if w.Status != WalletStatusFrozen {
		t.Errorf("expected frozen status, got %q", w.Status)
	}
	if cap.path != "/v1/wallets/w-1/freeze" || cap.method != http.MethodPost {
		t.Errorf("unexpected request: %s %s", cap.method, cap.path)
	}

	var sent FreezeWalletRequest
	if err := json.Unmarshal(cap.body, &sent); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if sent.FreezeReason != "fraud-investigation" {
		t.Errorf("expected freeze_reason in body, got %+v", sent)
	}

	// Unfreeze
	*cap = capture{}
	srv2 := mockServer(t, http.StatusOK, `{"wallet_id":"w-1","status":"WALLET_STATUS_ACTIVE"}`, cap)
	defer srv2.Close()
	c2 := New(srv2.URL, "tenant-A", "key-xyz")
	w2, err := c2.Wallet.Unfreeze(context.Background(), "w-1", "case-resolved")
	if err != nil {
		t.Fatalf("Unfreeze error: %v", err)
	}
	if w2.Status != WalletStatusActive {
		t.Errorf("expected active status, got %q", w2.Status)
	}
	if cap.path != "/v1/wallets/w-1/unfreeze" {
		t.Errorf("unexpected path: %q", cap.path)
	}
}

func TestWalletService_GetTransaction(t *testing.T) {
	cap := &capture{}
	srv := mockServer(t, http.StatusOK, `{"transaction_id":"tx-99","wallet_id":"w-1"}`, cap)
	defer srv.Close()

	c := New(srv.URL, "tenant-A", "key-xyz")
	tx, err := c.Wallet.GetTransaction(context.Background(), "w-1", "tx-99")
	if err != nil {
		t.Fatalf("GetTransaction error: %v", err)
	}
	if tx.TransactionID != "tx-99" {
		t.Errorf("expected tx-99, got %q", tx.TransactionID)
	}
	if cap.path != "/v1/wallets/w-1/transactions/tx-99" {
		t.Errorf("unexpected path: %q", cap.path)
	}
	if !strings.Contains(cap.rawQuery, "tenant_id=tenant-A") {
		t.Errorf("expected tenant_id in query: %q", cap.rawQuery)
	}
}

func TestWalletService_List_PassesCursorAndLimit(t *testing.T) {
	cap := &capture{}
	respJSON := `{"wallets":[{"wallet_id":"w-1"}],"next_cursor":"opaque-token"}`
	srv := mockServer(t, http.StatusOK, respJSON, cap)
	defer srv.Close()

	c := New(srv.URL, "tenant-A", "key-xyz")
	resp, err := c.Wallet.List(context.Background(), "user-1", 25, "prev-token")
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if resp.NextCursor == nil || *resp.NextCursor != "opaque-token" {
		t.Errorf("expected next_cursor=opaque-token, got %+v", resp.NextCursor)
	}
	if !strings.Contains(cap.rawQuery, "cursor=prev-token") {
		t.Errorf("expected cursor in query, got %q", cap.rawQuery)
	}
	if !strings.Contains(cap.rawQuery, "limit=25") {
		t.Errorf("expected limit=25 in query, got %q", cap.rawQuery)
	}
}

func TestWalletService_APIErrorOnNon2xx(t *testing.T) {
	cap := &capture{}
	srv := mockServer(t, http.StatusForbidden, `{"error":"insufficient scope"}`, cap)
	defer srv.Close()

	c := New(srv.URL, "tenant-A", "key-xyz")
	_, err := c.Wallet.Get(context.Background(), "w-1")
	if err == nil {
		t.Fatalf("expected error on 403, got nil")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.Status != http.StatusForbidden {
		t.Errorf("expected 403, got %d", apiErr.Status)
	}
}

func TestTypedConstantsMarshalAsPlainStrings(t *testing.T) {
	// JSON serialization of a request with typed Currency/Network/Mint must
	// produce the same wire format as plain strings — server-side parsers
	// expect "USDC" not `{"value":"USDC"}`.
	req := CreateChainWalletRequest{
		TenantID:    "t-1",
		SubjectID:   "u-1",
		SubjectType: "user",
		Currency:    CurrencyUSDC,
		Network:     NetworkSolana,
	}
	got, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	gotS := string(got)
	if !strings.Contains(gotS, `"currency":"USDC"`) {
		t.Errorf("expected currency=USDC literal in JSON, got %s", gotS)
	}
	if !strings.Contains(gotS, `"network":"SOLANA"`) {
		t.Errorf("expected network=SOLANA literal in JSON, got %s", gotS)
	}

	// Mint marshals same way.
	wd, _ := json.Marshal(WithdrawRequest{Mint: MintUSDCSolanaMainnet, Network: NetworkSolana})
	if !strings.Contains(string(wd), `"mint":"EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"`) {
		t.Errorf("USDC mint literal mismatch: %s", string(wd))
	}

	// Stringer for log fields.
	if CurrencyUSDC.String() != "USDC" || NetworkSolana.String() != "SOLANA" {
		t.Errorf("Stringer should return raw value")
	}
}

func TestWalletService_ImportChain_FreshImportSurfacesWasExistingFalse(t *testing.T) {
	cap := &capture{}
	resp := ImportChainWalletResponse{
		Wallet: Wallet{
			WalletID: "w-1",
			Currency: "USDC",
			Status:   WalletStatusActive,
		},
		WasExisting: false,
		RecoveredAt: "2026-05-19T10:15:00Z",
	}
	srv := mockServer(t, http.StatusOK, mustJSON(t, resp), cap)
	defer srv.Close()

	c := New(srv.URL, "tenant-A", "key-xyz")
	out, err := c.Wallet.ImportChain(context.Background(), ImportChainWalletRequest{
		SubjectID:   "user-1",
		SubjectType: "user",
		Currency:    "USDC",
		Network:     "SOLANA",
		Mnemonic:    "abandon ability able about above absent absorb abstract absurd abuse access accident",
	}, "idem-import-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.WasExisting {
		t.Errorf("expected was_existing=false on fresh import, got true")
	}
	if out.Wallet.WalletID != "w-1" {
		t.Errorf("wallet id mismatch: %s", out.Wallet.WalletID)
	}
	if out.RecoveredAt == "" {
		t.Errorf("recovered_at should be populated")
	}
	if cap.path != "/v1/wallets/chain/import" {
		t.Errorf("path mismatch: %s", cap.path)
	}
	if cap.idempotencyKey != "idem-import-1" {
		t.Errorf("idempotency header missing: %q", cap.idempotencyKey)
	}
}

func TestWalletService_ImportChain_RecoverySurfacesWasExistingTrue(t *testing.T) {
	cap := &capture{}
	resp := ImportChainWalletResponse{
		Wallet: Wallet{
			WalletID: "w-existing",
			Currency: "USDC",
			Status:   WalletStatusActive,
		},
		WasExisting: true,
		RecoveredAt: "2026-04-01T12:00:00Z",
	}
	srv := mockServer(t, http.StatusOK, mustJSON(t, resp), cap)
	defer srv.Close()

	c := New(srv.URL, "tenant-A", "key-xyz")
	out, err := c.Wallet.ImportChain(context.Background(), ImportChainWalletRequest{
		SubjectID: "user-1",
		Mnemonic:  "abandon ability able about above absent absorb abstract absurd abuse access accident",
	}, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !out.WasExisting {
		t.Fatalf("expected was_existing=true on recovery")
	}
	if out.Wallet.WalletID != "w-existing" {
		t.Errorf("wallet id mismatch: %s", out.Wallet.WalletID)
	}
}

func TestWalletService_ImportChain_PropagatesCrossSubject403(t *testing.T) {
	cap := &capture{}
	srv := mockServer(t, http.StatusForbidden, `{"message":"wallets belong to different tenants"}`, cap)
	defer srv.Close()

	c := New(srv.URL, "tenant-A", "key-xyz")
	_, err := c.Wallet.ImportChain(context.Background(), ImportChainWalletRequest{
		SubjectID: "user-1",
		Mnemonic:  "abandon ability able about above absent absorb abstract absurd abuse access accident",
	}, "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.Status != http.StatusForbidden {
		t.Errorf("expected 403, got %d", apiErr.Status)
	}
}

func TestWalletService_Withdraw_TronUsdtMint(t *testing.T) {
	cap := &capture{}
	srv := mockServer(
		t,
		http.StatusOK,
		`{"transaction_id":"tx-tron","wallet_id":"w-tron","type":"TRANSACTION_TYPE_DEBIT","amount":"100","currency":"USDT","external_ref":"d83f...txid"}`,
		cap,
	)
	defer srv.Close()

	c := New(srv.URL, "tenant-A", "key-xyz")
	_, err := c.Wallet.Withdraw(context.Background(), WithdrawRequest{
		WalletID:        "w-tron",
		Amount:          "100",
		DestinationType: "crypto_address",
		DestinationID:   "TXyz1234567890abcdefABCDEFghIJKLmn",
		Network:         NetworkTron,
		Mint:            MintUSDTTronMainnet,
	}, "idem-tron-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var sent map[string]any
	if err := json.Unmarshal(cap.body, &sent); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if sent["network"] != "TRON" {
		t.Errorf("network mismatch: %v", sent["network"])
	}
	if sent["mint"] != "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t" {
		t.Errorf("USDT TRC-20 mint mismatch: %v", sent["mint"])
	}
}

func TestNetworkTronWireValue(t *testing.T) {
	// Server-side parser compares against the literal "TRON".
	if string(NetworkTron) != "TRON" {
		t.Errorf("NetworkTron wire mismatch: %s", NetworkTron)
	}
	if string(MintUSDTTronMainnet) != "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t" {
		t.Errorf("USDT TRC-20 mint literal mismatch: %s", MintUSDTTronMainnet)
	}
}

func TestWalletService_Withdraw_PassesSponsorWalletID(t *testing.T) {
	cap := &capture{}
	srv := mockServer(
		t,
		http.StatusOK,
		`{"transaction_id":"tx-sp","wallet_id":"w-from","type":"TRANSACTION_TYPE_DEBIT","amount":"5.00","currency":"USDC"}`,
		cap,
	)
	defer srv.Close()

	c := New(srv.URL, "tenant-A", "key-xyz")
	_, err := c.Wallet.Withdraw(context.Background(), WithdrawRequest{
		WalletID:        "w-from",
		Amount:          "5.00",
		DestinationType: "crypto_address",
		DestinationID:   "DEsT_sOlanA",
		Network:         "SOLANA",
		Mint:            MintUSDCSolanaMainnet,
		SponsorWalletID: "spon-uuid",
	}, "idem-spon-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var sent map[string]any
	if err := json.Unmarshal(cap.body, &sent); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if sent["sponsor_wallet_id"] != "spon-uuid" {
		t.Errorf("sponsor_wallet_id not forwarded: %v", sent["sponsor_wallet_id"])
	}
	if cap.path != "/v1/wallets/w-from/withdraw" {
		t.Errorf("path mismatch: %s", cap.path)
	}
}

func TestWalletService_Transfer_SendsExpectedShape(t *testing.T) {
	cap := &capture{}
	resp := TransferResponse{
		TransferID: "xfer-uuid",
		Debit: WalletTransaction{
			TransactionID: "tx-debit",
			WalletID:      "w-from",
			Type:          TransactionDebit,
			Amount:        "25.50",
			Currency:      "USDC",
			BalanceAfter:  "74.50",
		},
		Credit: WalletTransaction{
			TransactionID: "tx-credit",
			WalletID:      "w-to",
			Type:          TransactionCredit,
			Amount:        "25.50",
			Currency:      "USDC",
			BalanceAfter:  "125.50",
		},
	}
	srv := mockServer(t, http.StatusOK, mustJSON(t, resp), cap)
	defer srv.Close()

	c := New(srv.URL, "tenant-A", "key-xyz")
	out, err := c.Wallet.Transfer(context.Background(), TransferRequest{
		FromWalletID: "w-from",
		ToWalletID:   "w-to",
		Amount:       "25.50",
		Reason:       ReasonSettlement,
		Description:  "monthly close",
		MerchantID:   "m-1",
		Note:         "Q2 settlement",
	}, "idem-xfer-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.TransferID != "xfer-uuid" {
		t.Errorf("transfer_id mismatch: %s", out.TransferID)
	}
	if out.Debit.WalletID != "w-from" || out.Credit.WalletID != "w-to" {
		t.Errorf("leg wallets mismatch: debit=%s credit=%s", out.Debit.WalletID, out.Credit.WalletID)
	}
	if cap.method != http.MethodPost {
		t.Errorf("expected POST, got %s", cap.method)
	}
	if cap.path != "/v1/wallets/w-from/transfer" {
		t.Errorf("path mismatch: %s", cap.path)
	}
	if cap.idempotencyKey != "idem-xfer-1" {
		t.Errorf("idempotency header missing: %q", cap.idempotencyKey)
	}
	// Body must echo the tenant_id (client overwrites) and carry the
	// caller-provided fields verbatim.
	var sent map[string]any
	if err := json.Unmarshal(cap.body, &sent); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if sent["tenant_id"] != "tenant-A" {
		t.Errorf("tenant_id mismatch in body: %v", sent["tenant_id"])
	}
	if sent["from_wallet_id"] != "w-from" || sent["to_wallet_id"] != "w-to" {
		t.Errorf("wallet ids mismatch: %v", sent)
	}
	if sent["amount"] != "25.50" {
		t.Errorf("amount mismatch: %v", sent["amount"])
	}
	if sent["note"] != "Q2 settlement" {
		t.Errorf("note mismatch: %v", sent["note"])
	}
}

func TestWalletService_Transfer_PropagatesServerError(t *testing.T) {
	cap := &capture{}
	srv := mockServer(t, http.StatusPreconditionFailed, `{"message":"insufficient balance on source"}`, cap)
	defer srv.Close()

	c := New(srv.URL, "tenant-A", "key-xyz")
	_, err := c.Wallet.Transfer(context.Background(), TransferRequest{
		FromWalletID: "w-from",
		ToWalletID:   "w-to",
		Amount:       "1000000",
	}, "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.Status != http.StatusPreconditionFailed {
		t.Errorf("expected 412, got %d", apiErr.Status)
	}
}

func TestWalletService_Transfer_MirrorsIdempotencyKeyIntoBody(t *testing.T) {
	// Caller passed the key via the helper arg only — Transfer should also
	// surface it in the body so server-side per-leg namespacing works
	// regardless of whether the gateway preserves the header.
	cap := &capture{}
	srv := mockServer(t, http.StatusOK, `{"transfer_id":"x","debit":{},"credit":{}}`, cap)
	defer srv.Close()

	c := New(srv.URL, "tenant-A", "key-xyz")
	_, err := c.Wallet.Transfer(context.Background(), TransferRequest{
		FromWalletID: "w-from",
		ToWalletID:   "w-to",
		Amount:       "1",
	}, "via-header-arg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var sent map[string]any
	_ = json.Unmarshal(cap.body, &sent)
	if sent["idempotency_key"] != "via-header-arg" {
		t.Errorf("idempotency_key not mirrored into body: %v", sent["idempotency_key"])
	}
	if cap.idempotencyKey != "via-header-arg" {
		t.Errorf("idempotency header missing: %q", cap.idempotencyKey)
	}
}

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return string(b)
}
