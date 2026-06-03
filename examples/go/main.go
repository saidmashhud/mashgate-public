// Mashgate Go SDK quickstart.
//
// This single file shows the three things every integration needs:
//
//  1. Initialize the client from environment variables.
//  2. Create a hosted checkout session (and, optionally, a single payment)
//     and print the redirect URL + ids.
//  3. Receive webhooks and verify their signature with the SDK helper.
//
// Run the checkout demo (default):
//
//	export MASHGATE_BASE_URL="https://api.mashgate.uz"   # or http://localhost:9661 for local dev
//	export MASHGATE_API_KEY="mg_test_..."
//	go run .
//
// Run the webhook receiver instead:
//
//	export MASHGATE_WEBHOOK_SECRET="whsec_..."           # SigningSecret from your endpoint
//	go run . serve                                       # listens on :8080/webhooks
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	mashgate "github.com/saidmashhud/mashgate-public/sdk/go"
)

func main() {
	// `go run . serve` starts the webhook receiver; anything else runs the
	// checkout + payment demo.
	if len(os.Args) > 1 && os.Args[1] == "serve" {
		serveWebhooks()
		return
	}
	runCheckoutDemo()
}

// newClient builds a Mashgate client from the environment. The base URL is read
// from MASHGATE_BASE_URL (falling back to MASHGATE_API_URL), the key from
// MASHGATE_API_KEY. Construct the client once and reuse it — it holds a pooled
// *http.Client.
func newClient() *mashgate.Client {
	baseURL := firstEnv("MASHGATE_BASE_URL", "MASHGATE_API_URL")
	if baseURL == "" {
		baseURL = "https://api.mashgate.uz" // sensible default
	}
	apiKey := os.Getenv("MASHGATE_API_KEY")
	if apiKey == "" {
		log.Fatal("set MASHGATE_API_KEY (mg_test_... or mg_live_...)")
	}
	return mashgate.New(baseURL, apiKey)
}

// runCheckoutDemo creates a hosted checkout session, prints the redirect URL,
// then creates a standalone payment intent and prints its id.
func runCheckoutDemo() {
	client := newClient()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// 1. Hosted checkout: Mashgate renders the payment page; you redirect the
	//    customer to CheckoutURL and they come back to SuccessURL / CancelURL.
	//    CreateCheckout auto-generates an idempotency key, so a retried request
	//    never creates a duplicate session.
	session, err := client.CreateCheckout(ctx, mashgate.CreateCheckoutRequest{
		TotalAmount:   mashgate.Money{Amount: "150000.00", Currency: "UZS"},
		SuccessURL:    "https://example.com/success?session={sessionId}",
		CancelURL:     "https://example.com/cancel",
		CustomerEmail: "customer@example.com",
		Items: []mashgate.LineItem{
			{
				Name:      "Pro plan",
				Quantity:  1,
				UnitPrice: mashgate.Money{Amount: "150000.00", Currency: "UZS"},
			},
		},
		Metadata: map[string]string{"order_ref": "demo-001"},
	})
	if err != nil {
		fatalAPIError("create checkout", err)
	}
	fmt.Println("Checkout session created:")
	fmt.Printf("  session id   : %s\n", session.SessionID)
	fmt.Printf("  status       : %s\n", session.Status)
	fmt.Printf("  redirect URL : %s\n", session.CheckoutURL)

	// 2. Single payment intent (server-to-server). Use this when you collect
	//    the order amount yourself rather than via the hosted page. OrderID is
	//    your own reference; CaptureMode "AUTO" captures immediately (use
	//    "MANUAL" to authorize now and capture later).
	payment, err := client.CreatePayment(ctx, mashgate.CreatePaymentRequest{
		Amount:      mashgate.Money{Amount: "150000.00", Currency: "UZS"},
		OrderID:     "demo-001",
		CaptureMode: "AUTO",
		Metadata:    map[string]string{"order_ref": "demo-001"},
	})
	if err != nil {
		fatalAPIError("create payment", err)
	}
	fmt.Println("Payment intent created:")
	fmt.Printf("  payment id : %s\n", payment.PaymentID)
	fmt.Printf("  status     : %s\n", payment.Status)
}

// serveWebhooks starts a minimal HTTP server that receives Mashgate events and
// verifies their signature before acting on them.
func serveWebhooks() {
	secret := os.Getenv("MASHGATE_WEBHOOK_SECRET")
	if secret == "" {
		log.Fatal("set MASHGATE_WEBHOOK_SECRET (the SigningSecret returned when you created the endpoint)")
	}

	http.HandleFunc("/webhooks", func(w http.ResponseWriter, r *http.Request) {
		// Read the RAW body before any JSON parsing — the signature is computed
		// over the exact bytes Mashgate sent.
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "cannot read body", http.StatusBadRequest)
			return
		}

		// Verify HMAC-SHA256 over "{timestamp}.{body}". The helper also rejects
		// timestamps older than 5 minutes to block replays. The signature header
		// is "x-hl-signature" (format "v1=<hex>"), the timestamp "x-hl-timestamp".
		err = mashgate.VerifySignature(
			secret,
			r.Header.Get("x-hl-timestamp"),
			string(body),
			r.Header.Get("x-hl-signature"),
		)
		if err != nil {
			// Distinguish the common failure modes for clearer logs.
			switch {
			case errors.Is(err, mashgate.ErrInvalidSignature):
				log.Printf("rejected webhook: bad signature")
			case errors.Is(err, mashgate.ErrTimestampTooOld):
				log.Printf("rejected webhook: stale timestamp (possible replay)")
			default:
				log.Printf("rejected webhook: %v", err)
			}
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Signature is valid — parse the verified envelope.
		event, err := mashgate.ParseEvent(body)
		if err != nil {
			http.Error(w, "bad payload", http.StatusBadRequest)
			return
		}

		// Event type lives in Topic (envelope v1) or EventType (legacy emitters).
		eventType := event.Topic
		if eventType == "" {
			eventType = event.EventType
		}
		log.Printf("event %s: %s", event.CanonicalID(), eventType)

		switch eventType {
		case mashgate.EventPaymentCaptured:
			// Funds settled — fulfil the order. event.PayloadBytes() returns the
			// payload regardless of envelope vs. legacy format.
			log.Printf("payment captured, payload: %s", event.PayloadBytes())
		case mashgate.EventPaymentFailed:
			log.Printf("payment failed, payload: %s", event.PayloadBytes())
		case mashgate.EventCheckoutCompleted:
			log.Printf("checkout completed, payload: %s", event.PayloadBytes())
		default:
			log.Printf("unhandled event type: %s", eventType)
		}

		// Acknowledge quickly with 2xx so Mashgate stops retrying. Do heavy work
		// asynchronously, keyed by event id for idempotency.
		w.WriteHeader(http.StatusOK)
	})

	addr := ":8080"
	log.Printf("webhook receiver listening on %s/webhooks", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

// fatalAPIError logs a Mashgate *Error with the fields you want in support
// tickets (Code, RequestID, DocURL) and exits.
func fatalAPIError(action string, err error) {
	var mgErr *mashgate.Error
	if errors.As(err, &mgErr) {
		log.Fatalf("%s failed: code=%s request_id=%s doc=%s",
			action, mgErr.Code, mgErr.RequestID, mgErr.DocURL())
	}
	log.Fatalf("%s: %v", action, err)
}

// firstEnv returns the first non-empty environment variable from keys.
func firstEnv(keys ...string) string {
	for _, k := range keys {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	return ""
}
