package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	mashgate "github.com/saidmashhud/mashgate-public/sdk/go"
)

func main() {
	apiURL, token := os.Getenv("MASHGATE_API_URL"), os.Getenv("MASHGATE_ACCESS_TOKEN")
	if apiURL == "" || token == "" {
		log.Fatal("set MASHGATE_API_URL and MASHGATE_ACCESS_TOKEN")
	}
	client := mashgate.NewClient(
		os.Getenv("MASHGATE_API_KEY"),
		mashgate.WithBaseURL(apiURL),
		mashgate.WithAccessToken(token),
	)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	markets, err := client.Exchange.ListMarkets(ctx)
	if err != nil {
		log.Fatal(err)
	}
	balances, err := client.Exchange.ListBalances(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("markets=%d balances=%d\n", len(markets), len(balances))

	if os.Getenv("PLACE_TEST_ORDER") != "1" {
		return
	}
	order, err := client.Exchange.PlaceOrder(ctx, mashgate.PlaceExchangeOrderRequest{
		Market: "BTC/USDT", Side: mashgate.ExchangeOrderSideBuy,
		Kind: mashgate.ExchangeOrderKindLimit, TimeInForce: mashgate.ExchangeTimeInForceGTC,
		Price: "1.00", Quantity: "0.00001", PostOnly: true,
	}, "exchange-example-order-001")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("accepted order=%s status=%s\n", order.OrderID, order.Status)
}
