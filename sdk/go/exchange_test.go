package mashgate

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExchangePlaceOrderUsesBearerAndIdempotency(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/exchange/orders" || r.Method != http.MethodPost {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer user-token" {
			t.Fatalf("authorization = %q", got)
		}
		if got := r.Header.Get("Idempotency-Key"); got != "order-1" {
			t.Fatalf("idempotency key = %q", got)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if _, exists := body["accountId"]; exists {
			t.Fatal("ownership must not be supplied by client")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"orderId":"11111111-1111-1111-1111-111111111111","market":"BTC/USDT","status":"ORDER_STATUS_ACCEPTED"}`))
	}))
	defer server.Close()

	client := NewClient("server-key", WithBaseURL(server.URL), WithAccessToken("user-token"))
	order, err := client.Exchange.PlaceOrder(context.Background(), PlaceExchangeOrderRequest{
		Market:      "BTC/USDT",
		Side:        ExchangeOrderSideBuy,
		Kind:        ExchangeOrderKindLimit,
		TimeInForce: ExchangeTimeInForceGTC,
		Price:       "50000",
		Quantity:    "0.001",
	}, "order-1")
	if err != nil {
		t.Fatal(err)
	}
	if order.Market != "BTC/USDT" {
		t.Fatalf("market = %q", order.Market)
	}
}

func TestExchangeOrderBookUsesQueryForSlashMarket(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/exchange/order-book" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("market"); got != "BTC/USDT" {
			t.Fatalf("market = %q", got)
		}
		_, _ = w.Write([]byte(`{"market":"BTC/USDT","sequence":"1","bids":[],"asks":[]}`))
	}))
	defer server.Close()

	client := NewClient("", WithBaseURL(server.URL), WithAccessToken("user-token"))
	book, err := client.Exchange.GetOrderBook(context.Background(), "BTC/USDT", 20)
	if err != nil {
		t.Fatal(err)
	}
	if book.Sequence != 1 {
		t.Fatalf("sequence = %d", book.Sequence)
	}
}
