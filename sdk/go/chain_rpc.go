package mashgate

import (
	"context"
	"net/http"
	"net/url"
)

// ChainRPCClient exposes stateless chain transaction construction and
// broadcast primitives. It never generates, receives, or stores private keys.
type ChainRPCClient struct {
	c *Client
}

type ChainRPCBalance struct {
	Balance  string `json:"balance"`
	USDValue string `json:"usdValue"`
	Asset    string `json:"asset"`
	Network  string `json:"network"`
	Error    string `json:"error,omitempty"`
}

type RecentBlockhash struct {
	Blockhash       string `json:"blockhash"`
	LastValidHeight string `json:"lastValidHeight"`
	Error           string `json:"error,omitempty"`
}

type BuildSolanaTransferRequest struct {
	FromPubkeyBase58      string `json:"fromPubkeyBase58"`
	ToPubkeyBase58        string `json:"toPubkeyBase58"`
	MintBase58            string `json:"mintBase58,omitempty"`
	Amount                string `json:"amount"`
	RecentBlockhashBase58 string `json:"recentBlockhashBase58"`
	FeePayerPubkeyBase58  string `json:"feePayerPubkeyBase58,omitempty"`
}

type BuildSolanaTransferResponse struct {
	UnsignedMessageBase64 string `json:"unsignedMessageBase64"`
	MessageHashHex        string `json:"messageHashHex"`
	FeeLamports           string `json:"feeLamports"`
	Error                 string `json:"error,omitempty"`
}

type BroadcastTransactionRequest struct {
	Network       string `json:"network"`
	SignedPayload string `json:"signedTxHex"`
}

type BroadcastTransactionResponse struct {
	TxHash string `json:"txHash"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type ChainTransactionStatus struct {
	Status        string `json:"status"`
	Confirmations string `json:"confirmations"`
	BlockNumber   string `json:"blockNumber"`
	BlockHash     string `json:"blockHash"`
	Timestamp     string `json:"timestamp"`
	Error         string `json:"error,omitempty"`
}

func (ch *ChainRPCClient) GetBalance(ctx context.Context, network, address, asset string) (*ChainRPCBalance, error) {
	query := url.Values{}
	query.Set("network", network)
	query.Set("address", address)
	if asset != "" {
		query.Set("asset", asset)
	}
	var out ChainRPCBalance
	if err := ch.c.do(ctx, http.MethodGet, "/v1/chain/balance?"+query.Encode(), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (ch *ChainRPCClient) GetRecentBlockhash(ctx context.Context) (*RecentBlockhash, error) {
	var out RecentBlockhash
	if err := ch.c.do(ctx, http.MethodGet, "/v1/chain/solana/blockhash", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (ch *ChainRPCClient) BuildSolanaTransfer(ctx context.Context, req BuildSolanaTransferRequest) (*BuildSolanaTransferResponse, error) {
	var out BuildSolanaTransferResponse
	if err := ch.c.do(ctx, http.MethodPost, "/v1/chain/solana/build-transfer", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (ch *ChainRPCClient) BroadcastTransaction(ctx context.Context, req BroadcastTransactionRequest, idempotencyKey string) (*BroadcastTransactionResponse, error) {
	headers := map[string]string{}
	if idempotencyKey != "" {
		headers["Idempotency-Key"] = idempotencyKey
	}
	var out BroadcastTransactionResponse
	if err := ch.c.doWithHeader(ctx, http.MethodPost, "/v1/chain/broadcast", headers, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (ch *ChainRPCClient) GetTransactionStatus(ctx context.Context, network, txHash string) (*ChainTransactionStatus, error) {
	query := url.Values{}
	query.Set("network", network)
	query.Set("txHash", txHash)
	var out ChainTransactionStatus
	if err := ch.c.do(ctx, http.MethodGet, "/v1/chain/tx-status?"+query.Encode(), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
