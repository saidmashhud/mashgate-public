package mashgate

import (
	"context"
	"fmt"
	"net/url"
)

// ────────────────────────────────────────────────────────────────────────────
// ChainClient — crypto rails (mgChain). Wraps ChainService in chain.proto.
// 21 RPCs (subset implemented here — most-used 14).
//
// Backend: chain-rpc (Rust) + chain-indexer (Rust) + mgchain-orchestrator (Scala).
// Tenant isolation enforced via JWT — ChainService never trusts body tenant_id.
// ────────────────────────────────────────────────────────────────────────────

type ChainClient struct {
	c *Client
}

// ── Address management ───────────────────────────────────────────────────

// CreateAddress derives a new deposit address for tenant on given network.
func (ch *ChainClient) CreateAddress(ctx context.Context, req CreateAddressRequest) (*ChainAddress, error) {
	var out ChainAddress
	if err := ch.c.do(ctx, "POST", "/v1/chain/addresses", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListAddresses returns all tenant addresses on optional network filter.
func (ch *ChainClient) ListAddresses(ctx context.Context, tenantID, network string) ([]*ChainAddress, error) {
	path := fmt.Sprintf("/v1/chain/addresses?tenantId=%s", url.QueryEscape(tenantID))
	if network != "" {
		path += "&network=" + url.QueryEscape(network)
	}
	var out struct {
		Addresses []*ChainAddress `json:"addresses"`
	}
	if err := ch.c.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return out.Addresses, nil
}

// GetAddress retrieves a single address by id.
func (ch *ChainClient) GetAddress(ctx context.Context, addressID string) (*ChainAddress, error) {
	var out ChainAddress
	if err := ch.c.do(ctx, "GET", "/v1/chain/addresses/"+addressID, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ── Balance ──────────────────────────────────────────────────────────────

// GetBalance returns balance for an address on its network.
func (ch *ChainClient) GetBalance(ctx context.Context, addressID string) (*ChainBalance, error) {
	var out ChainBalance
	if err := ch.c.do(ctx, "GET", "/v1/chain/addresses/"+addressID+"/balance", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ── Transactions ─────────────────────────────────────────────────────────

// ListTransactions returns transaction history for a tenant (optionally per address).
func (ch *ChainClient) ListTransactions(ctx context.Context, tenantID, addressID string, page, pageSize int) ([]*ChainTransaction, error) {
	path := fmt.Sprintf("/v1/chain/transactions?tenantId=%s", url.QueryEscape(tenantID))
	if addressID != "" {
		path += "&addressId=" + url.QueryEscape(addressID)
	}
	if page > 0 {
		path += fmt.Sprintf("&page=%d", page)
	}
	if pageSize > 0 {
		path += fmt.Sprintf("&pageSize=%d", pageSize)
	}
	var out struct {
		Transactions []*ChainTransaction `json:"transactions"`
	}
	if err := ch.c.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return out.Transactions, nil
}

// GetTransaction retrieves a single tx by id or hash.
func (ch *ChainClient) GetTransaction(ctx context.Context, txID string) (*ChainTransaction, error) {
	var out ChainTransaction
	if err := ch.c.do(ctx, "GET", "/v1/chain/transactions/"+txID, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// EstimateFee returns estimated network fee for a hypothetical send.
func (ch *ChainClient) EstimateFee(ctx context.Context, req EstimateFeeRequest) (*FeeEstimate, error) {
	var out FeeEstimate
	if err := ch.c.do(ctx, "POST", "/v1/chain/fee-estimate", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SendTransaction submits a transaction to the network.
// Idempotent via IdempotencyKey header.
func (ch *ChainClient) SendTransaction(ctx context.Context, req SendTransactionRequest) (*ChainTransaction, error) {
	headers := map[string]string{}
	if req.IdempotencyKey != "" {
		headers["Idempotency-Key"] = req.IdempotencyKey
	}
	var out ChainTransaction
	if err := ch.c.doWithHeader(ctx, "POST", "/v1/chain/transactions/send", headers, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ── Blocks ───────────────────────────────────────────────────────────────

// GetBlock retrieves a block by hash or height on a given network.
func (ch *ChainClient) GetBlock(ctx context.Context, network, hashOrHeight string) (*ChainBlock, error) {
	path := fmt.Sprintf("/v1/chain/networks/%s/blocks/%s", url.PathEscape(network), url.PathEscape(hashOrHeight))
	var out ChainBlock
	if err := ch.c.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetLatestBlock returns the most recent block on a network.
func (ch *ChainClient) GetLatestBlock(ctx context.Context, network string) (*ChainBlock, error) {
	return ch.GetBlock(ctx, network, "latest")
}

// ── Networks ─────────────────────────────────────────────────────────────

// ListNetworks returns chains supported by this Mashgate instance.
func (ch *ChainClient) ListNetworks(ctx context.Context) ([]*ChainNetwork, error) {
	var out struct {
		Networks []*ChainNetwork `json:"networks"`
	}
	if err := ch.c.do(ctx, "GET", "/v1/chain/networks", nil, &out); err != nil {
		return nil, err
	}
	return out.Networks, nil
}
