package mashgate

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChainRPCBuildAndBroadcast(t *testing.T) {
	var sawIdempotency bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/chain/solana/build-transfer":
			var req BuildSolanaTransferRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("decode build request: %v", err)
			}
			if req.FromPubkeyBase58 != "from" || req.ToPubkeyBase58 != "to" || req.Amount != "1.25" {
				t.Fatalf("unexpected build request: %+v", req)
			}
			_, _ = w.Write([]byte(`{"unsignedMessageBase64":"message","messageHashHex":"hash","feeLamports":"5000"}`))
		case "/v1/chain/broadcast":
			sawIdempotency = r.Header.Get("Idempotency-Key") == "intent-1"
			_, _ = w.Write([]byte(`{"txHash":"tx-1","status":"submitted"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewWithTenant(server.URL, "tenant-1", "test-key")
	built, err := client.ChainRPC.BuildSolanaTransfer(context.Background(), BuildSolanaTransferRequest{
		FromPubkeyBase58:      "from",
		ToPubkeyBase58:        "to",
		Amount:                "1.25",
		RecentBlockhashBase58: "blockhash",
	})
	if err != nil || built.MessageHashHex != "hash" {
		t.Fatalf("build: response=%+v err=%v", built, err)
	}
	broadcast, err := client.ChainRPC.BroadcastTransaction(context.Background(), BroadcastTransactionRequest{
		Network:       "SOLANA",
		SignedPayload: "signed-base64",
	}, "intent-1")
	if err != nil || broadcast.TxHash != "tx-1" || !sawIdempotency {
		t.Fatalf("broadcast: response=%+v idempotency=%v err=%v", broadcast, sawIdempotency, err)
	}
}
