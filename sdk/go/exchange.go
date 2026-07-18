package mashgate

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// ExchangeClient exposes custodial spot trading and funding operations.
// Ownership is always derived from the bearer token configured with
// WithAccessToken; request types intentionally contain no account ID.
type ExchangeClient struct{ c *Client }

type ExchangeOrderSide string

const (
	ExchangeOrderSideBuy  ExchangeOrderSide = "ORDER_SIDE_BUY"
	ExchangeOrderSideSell ExchangeOrderSide = "ORDER_SIDE_SELL"
)

type ExchangeOrderKind string

const (
	ExchangeOrderKindLimit  ExchangeOrderKind = "ORDER_KIND_LIMIT"
	ExchangeOrderKindMarket ExchangeOrderKind = "ORDER_KIND_MARKET"
)

type ExchangeTimeInForce string

const (
	ExchangeTimeInForceGTC ExchangeTimeInForce = "TIME_IN_FORCE_GTC"
	ExchangeTimeInForceIOC ExchangeTimeInForce = "TIME_IN_FORCE_IOC"
)

type ExchangeMarket struct {
	Market        string `json:"market"`
	BaseAsset     string `json:"baseAsset"`
	QuoteAsset    string `json:"quoteAsset"`
	PriceScale    uint32 `json:"priceScale"`
	QuantityScale uint32 `json:"quantityScale"`
	MinQuantity   string `json:"minQuantity"`
	MinNotional   string `json:"minNotional"`
	Status        string `json:"status"`
}

type ExchangePriceLevel struct {
	Price      string `json:"price"`
	Quantity   string `json:"quantity"`
	OrderCount uint32 `json:"orderCount"`
}

type ExchangeOrderBook struct {
	Market   string               `json:"market"`
	Sequence uint64               `json:"sequence,string"`
	Bids     []ExchangePriceLevel `json:"bids"`
	Asks     []ExchangePriceLevel `json:"asks"`
}

type ExchangeOrder struct {
	OrderID           string              `json:"orderId"`
	AccountID         string              `json:"accountId"`
	Market            string              `json:"market"`
	Side              ExchangeOrderSide   `json:"side"`
	Kind              ExchangeOrderKind   `json:"kind"`
	TimeInForce       ExchangeTimeInForce `json:"timeInForce"`
	Price             string              `json:"price"`
	Quantity          string              `json:"quantity"`
	FilledQuantity    string              `json:"filledQuantity"`
	RemainingQuantity string              `json:"remainingQuantity"`
	Status            string              `json:"status"`
	PostOnly          bool                `json:"postOnly"`
	CreatedAt         string              `json:"createdAt"`
	UpdatedAt         string              `json:"updatedAt"`
}

type ExchangeTrade struct {
	TradeID       string            `json:"tradeId"`
	Market        string            `json:"market"`
	Price         string            `json:"price"`
	Quantity      string            `json:"quantity"`
	QuoteQuantity string            `json:"quoteQuantity"`
	TakerSide     ExchangeOrderSide `json:"takerSide"`
	ExecutedAt    string            `json:"executedAt"`
}

type ExchangeBalance struct {
	Asset     string `json:"asset"`
	Available string `json:"available"`
	Held      string `json:"held"`
	Total     string `json:"total"`
}

type ExchangeDepositAddress struct {
	Asset   string `json:"asset"`
	Network string `json:"network"`
	Address string `json:"address"`
	Memo    string `json:"memo"`
}

type ExchangeDeposit struct {
	DepositID     string `json:"depositId"`
	Asset         string `json:"asset"`
	Network       string `json:"network"`
	Amount        string `json:"amount"`
	Address       string `json:"address"`
	TxHash        string `json:"txHash"`
	Confirmations uint32 `json:"confirmations"`
	Status        string `json:"status"`
	DetectedAt    string `json:"detectedAt"`
	CreditedAt    string `json:"creditedAt"`
}

type ExchangeWithdrawal struct {
	WithdrawalID string `json:"withdrawalId"`
	Asset        string `json:"asset"`
	Network      string `json:"network"`
	Amount       string `json:"amount"`
	Destination  string `json:"destination"`
	Status       string `json:"status"`
	TxHash       string `json:"txHash"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

type PlaceExchangeOrderRequest struct {
	Market      string              `json:"market"`
	Side        ExchangeOrderSide   `json:"side"`
	Kind        ExchangeOrderKind   `json:"kind"`
	TimeInForce ExchangeTimeInForce `json:"timeInForce"`
	Price       string              `json:"price,omitempty"`
	Quantity    string              `json:"quantity"`
	PostOnly    bool                `json:"postOnly,omitempty"`
}

type RequestExchangeWithdrawal struct {
	Asset       string `json:"asset"`
	Network     string `json:"network"`
	Amount      string `json:"amount"`
	Destination string `json:"destination"`
}

type ExchangeOrderListOptions struct {
	Market string
	Status string
	Limit  uint32
	Cursor string
}

type ExchangeTransferListOptions struct {
	Asset  string
	Status string
	Limit  uint32
	Cursor string
}

func (e *ExchangeClient) ListMarkets(ctx context.Context) ([]ExchangeMarket, error) {
	var out struct {
		Markets []ExchangeMarket `json:"markets"`
	}
	if err := e.c.do(ctx, http.MethodGet, "/v1/exchange/markets", nil, &out); err != nil {
		return nil, err
	}
	return out.Markets, nil
}

func (e *ExchangeClient) GetOrderBook(ctx context.Context, market string, depth uint32) (*ExchangeOrderBook, error) {
	query := url.Values{"market": {market}}
	if depth > 0 {
		query.Set("depth", uintString(depth))
	}
	var out ExchangeOrderBook
	if err := e.c.do(ctx, http.MethodGet, "/v1/exchange/order-book?"+query.Encode(), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (e *ExchangeClient) ListTrades(ctx context.Context, market string, limit uint32, cursor string) ([]ExchangeTrade, string, error) {
	query := url.Values{"market": {market}}
	if limit > 0 {
		query.Set("limit", uintString(limit))
	}
	if cursor != "" {
		query.Set("cursor", cursor)
	}
	var out struct {
		Trades     []ExchangeTrade `json:"trades"`
		NextCursor string          `json:"nextCursor"`
	}
	if err := e.c.do(ctx, http.MethodGet, "/v1/exchange/trades?"+query.Encode(), nil, &out); err != nil {
		return nil, "", err
	}
	return out.Trades, out.NextCursor, nil
}

func (e *ExchangeClient) ListBalances(ctx context.Context) ([]ExchangeBalance, error) {
	var out struct {
		Balances []ExchangeBalance `json:"balances"`
	}
	if err := e.c.do(ctx, http.MethodGet, "/v1/exchange/balances", nil, &out); err != nil {
		return nil, err
	}
	return out.Balances, nil
}

func (e *ExchangeClient) PlaceOrder(ctx context.Context, request PlaceExchangeOrderRequest, idempotencyKey string) (*ExchangeOrder, error) {
	var out ExchangeOrder
	if err := e.c.doWithHeader(ctx, http.MethodPost, "/v1/exchange/orders", requiredIdempotencyHeader(idempotencyKey), request, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (e *ExchangeClient) ListOrders(ctx context.Context, options ExchangeOrderListOptions) ([]ExchangeOrder, string, error) {
	query := url.Values{}
	if options.Market != "" {
		query.Set("market", options.Market)
	}
	if options.Status != "" {
		query.Set("status", options.Status)
	}
	if options.Limit > 0 {
		query.Set("limit", uintString(options.Limit))
	}
	if options.Cursor != "" {
		query.Set("cursor", options.Cursor)
	}
	var out struct {
		Orders     []ExchangeOrder `json:"orders"`
		NextCursor string          `json:"nextCursor"`
	}
	if err := e.c.do(ctx, http.MethodGet, "/v1/exchange/orders?"+query.Encode(), nil, &out); err != nil {
		return nil, "", err
	}
	return out.Orders, out.NextCursor, nil
}

func (e *ExchangeClient) GetOrder(ctx context.Context, orderID string) (*ExchangeOrder, error) {
	var out ExchangeOrder
	if err := e.c.do(ctx, http.MethodGet, "/v1/exchange/orders/"+url.PathEscape(orderID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (e *ExchangeClient) CancelOrder(ctx context.Context, orderID, idempotencyKey string) (*ExchangeOrder, error) {
	var out ExchangeOrder
	if err := e.c.doWithHeader(ctx, http.MethodPost, "/v1/exchange/orders/"+url.PathEscape(orderID)+"/cancel", requiredIdempotencyHeader(idempotencyKey), struct{}{}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (e *ExchangeClient) GetDepositAddress(ctx context.Context, asset, network string) (*ExchangeDepositAddress, error) {
	query := url.Values{"asset": {asset}, "network": {network}}
	var out ExchangeDepositAddress
	if err := e.c.do(ctx, http.MethodGet, "/v1/exchange/deposits/address?"+query.Encode(), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (e *ExchangeClient) ListDeposits(ctx context.Context, options ExchangeTransferListOptions) ([]ExchangeDeposit, string, error) {
	query := exchangeTransferQuery(options)
	var out struct {
		Deposits   []ExchangeDeposit `json:"deposits"`
		NextCursor string            `json:"nextCursor"`
	}
	if err := e.c.do(ctx, http.MethodGet, "/v1/exchange/deposits?"+query.Encode(), nil, &out); err != nil {
		return nil, "", err
	}
	return out.Deposits, out.NextCursor, nil
}

func (e *ExchangeClient) RequestWithdrawal(ctx context.Context, request RequestExchangeWithdrawal, idempotencyKey string) (*ExchangeWithdrawal, error) {
	var out ExchangeWithdrawal
	if err := e.c.doWithHeader(ctx, http.MethodPost, "/v1/exchange/withdrawals", requiredIdempotencyHeader(idempotencyKey), request, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (e *ExchangeClient) ListWithdrawals(ctx context.Context, options ExchangeTransferListOptions) ([]ExchangeWithdrawal, string, error) {
	query := exchangeTransferQuery(options)
	var out struct {
		Withdrawals []ExchangeWithdrawal `json:"withdrawals"`
		NextCursor  string               `json:"nextCursor"`
	}
	if err := e.c.do(ctx, http.MethodGet, "/v1/exchange/withdrawals?"+query.Encode(), nil, &out); err != nil {
		return nil, "", err
	}
	return out.Withdrawals, out.NextCursor, nil
}

func (e *ExchangeClient) GetWithdrawal(ctx context.Context, withdrawalID string) (*ExchangeWithdrawal, error) {
	var out ExchangeWithdrawal
	if err := e.c.do(ctx, http.MethodGet, "/v1/exchange/withdrawals/"+url.PathEscape(withdrawalID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func exchangeTransferQuery(options ExchangeTransferListOptions) url.Values {
	query := url.Values{}
	if options.Asset != "" {
		query.Set("asset", options.Asset)
	}
	if options.Status != "" {
		query.Set("status", options.Status)
	}
	if options.Limit > 0 {
		query.Set("limit", uintString(options.Limit))
	}
	if options.Cursor != "" {
		query.Set("cursor", options.Cursor)
	}
	return query
}

func requiredIdempotencyHeader(key string) map[string]string {
	return map[string]string{"Idempotency-Key": key}
}

func uintString(value uint32) string {
	return fmt.Sprintf("%d", value)
}
