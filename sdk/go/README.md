# Mashgate Go SDK

Go client for the [Mashgate](https://github.com/saidmashhud/mashgate) BaaS platform — backend infrastructure for Central Asia.

Covers the full Mashgate product suite: Payments, Identity, Webhooks, Chat, Notify, Storage, Feature Flags, Guard, Invoices, Logs, and Subscriptions.

## Installation

```sh
go get github.com/saidmashhud/mashgate-public/sdk/go@main
```

No registry needed. Import directly from GitHub.

## Quick start — rental marketplace (Zist)

```go
package main

import (
    "context"
    "io"
    "log"
    "net/http"
    "os"

    mg "github.com/saidmashhud/mashgate-public/sdk/go"
)

func main() {
    client := mg.New(
        "https://api.mashgate.uz",
        os.Getenv("MASHGATE_API_KEY"),
    )
    ctx := context.Background()

    // Guest books an apartment — create hosted checkout session
    session, err := client.CreateCheckout(ctx, mg.CreateCheckoutRequest{
        TotalAmount:   mg.Money{Amount: "2500000.00", Currency: "UZS"},
        SuccessURL:    "https://zist.uz/booking/success?session={sessionId}",
        CancelURL:     "https://zist.uz/booking/cancel",
        CustomerEmail: "guest@example.com",
        Items: []mg.LineItem{
            {
                Name:      "Apartment in Tashkent",
                Quantity:  3,
                UnitPrice: mg.Money{Amount: "833333.33", Currency: "UZS"},
            },
        },
    })
    if err != nil {
        log.Fatalf("create checkout: %v", err)
    }
    log.Printf("redirect guest to: %s", session.CheckoutURL)

    // Receive payment events via webhook
    http.HandleFunc("/webhooks/mashgate", func(w http.ResponseWriter, r *http.Request) {
        body, _ := io.ReadAll(r.Body)

        err := mg.VerifySignature(
            os.Getenv("MASHGATE_WEBHOOK_SECRET"),
            r.Header.Get("x-hl-timestamp"),
            string(body),
            r.Header.Get("x-hl-signature"),
        )
        if err != nil {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }

        event, _ := mg.ParseEvent(body)
        switch event.EventType {
        case mg.EventPaymentCaptured:
            // confirm booking, notify host
            log.Printf("payment captured: %s", event.AggregateID)
        case mg.EventPaymentFailed:
            // release hold, notify guest
            log.Printf("payment failed: %s", event.AggregateID)
        case mg.EventCheckoutExpired:
            // cancel pending booking
            log.Printf("checkout expired: %s", event.AggregateID)
        }
        w.WriteHeader(http.StatusOK)
    })

    log.Fatal(http.ListenAndServe(":8080", nil))
    _ = ctx
}
```

## Usage in a service

```go
import mg "github.com/saidmashhud/mashgate-public/sdk/go"

type PaymentService struct {
    mg *mg.Client
}

func NewPaymentService() *PaymentService {
    return &PaymentService{
        mg: mg.New(
            os.Getenv("MASHGATE_URL"),
            os.Getenv("MASHGATE_API_KEY"),
        ),
    }
}

func (s *PaymentService) CreateBookingPayment(ctx context.Context, booking Booking) (string, error) {
    session, err := s.mg.CreateCheckout(ctx, mg.CreateCheckoutRequest{
        TotalAmount:   mg.Money{Amount: booking.TotalUZS, Currency: "UZS"},
        SuccessURL:    "https://zist.uz/booking/" + booking.ID + "/success",
        CancelURL:     "https://zist.uz/booking/" + booking.ID + "/cancel",
        CustomerEmail: booking.GuestEmail,
        Items: []mg.LineItem{{
            Name:      booking.PropertyName,
            Quantity:  booking.Nights,
            UnitPrice: mg.Money{Amount: booking.PricePerNightUZS, Currency: "UZS"},
        }},
    })
    if err != nil {
        return "", err
    }
    return session.CheckoutURL, nil
}
```

## API reference

### Client

```go
client := mg.New(baseURL, apiKey)
```

| Parameter | Description |
|-----------|-------------|
| `baseURL` | `https://api.mashgate.uz` in production, `http://localhost:9661` for local dev |
| `apiKey`  | `mg_test_...` for sandbox, `mg_live_...` for production |

### Payments

```go
// Create
payment, err := client.CreatePayment(ctx, mg.CreatePaymentRequest{
    Amount:  mg.Money{Amount: "150000.00", Currency: "UZS"},
    OrderID: "order_001",
})

// Lifecycle
payment, err = client.AuthorizePayment(ctx, payment.PaymentID)
payment, err = client.CapturePayment(ctx, payment.PaymentID)
payment, err = client.VoidPayment(ctx, payment.PaymentID)

// Partial refund
payment, err = client.RefundPayment(ctx, payment.PaymentID, mg.RefundRequest{
    Amount: mg.Money{Amount: "50000.00", Currency: "UZS"},
    Reason: "customer_request",
})

// List
payments, err := client.ListPayments(ctx, mg.ListPaymentsParams{Status: "captured", PageSize: 50})
```

### Checkout (hosted payment page)

```go
session, err := client.CreateCheckout(ctx, mg.CreateCheckoutRequest{...})
// → redirect customer to session.CheckoutURL

session, err = client.GetCheckout(ctx, sessionID)
err = client.ExpireCheckout(ctx, sessionID)
```

### Webhooks

```go
// Register endpoint
endpoint, err := client.CreateWebhookEndpoint(ctx, mg.CreateWebhookEndpointRequest{
    URL:         "https://zist.uz/webhooks/mashgate",
    Description: "Zist booking payments",
    EventTypes:  []string{mg.EventPaymentCaptured, mg.EventPaymentFailed},
})
// Store endpoint.SigningSecret securely — only returned once

// Verify incoming webhook
err = mg.VerifySignature(secret, timestamp, body, signature)

// Parse event
event, err := mg.ParseEvent(body)
```

### Wallet (end-user view, top-level package)

```go
balance, err := client.GetWalletBalance(ctx, "UZS")
methods, err := client.ListSavedPaymentMethods(ctx)
err = client.SetDefaultPaymentMethod(ctx, paymentMethodID)
err = client.RemoveSavedPaymentMethod(ctx, paymentMethodID)
```

### Wallet (admin/merchant API, `fintech` subpackage)

The full `wallet.v1.WalletService` from
[`mashgate/contracts/proto/v1/wallet.proto`](https://github.com/saidmashhud/mashgate/blob/main/contracts/proto/v1/wallet.proto),
including on-chain (mgCrypto) flows. Auth is tenant-scoped — pass an admin
JWT or service-account API key.

`Currency`, `Network`, `Mint` are typed string aliases with predefined
constants (`fintech.CurrencyUSDC`, `fintech.NetworkSolana`,
`fintech.MintUSDCSolanaMainnet`, …). They serialize as plain JSON strings
on the wire, so callers using string literals stay compatible.

```go
import "github.com/saidmashhud/mashgate-public/sdk/go/fintech"

c := fintech.New("https://api.mashgate.io", tenantID, apiKey)

// Off-chain wallet
w, err := c.Wallet.Create(ctx, fintech.CreateWalletRequest{
    SubjectID: "user-123", SubjectType: "user",
    WalletType: fintech.WalletTypeFiat,
    Currency:   fintech.CurrencyUZS,
}, "idem-create-1")

// On-chain wallet (BIP-39 mnemonic returned ONCE — surface to user, never persist)
chain, err := c.Wallet.CreateChain(ctx, fintech.CreateChainWalletRequest{
    SubjectID: "user-123", SubjectType: "user",
    Currency: fintech.CurrencyUSDC,
    Network:  fintech.NetworkSolana,
}, "idem-chain-1")
showOnceToEndUser(chain.Mnemonic)

// Deposit address — pass `Mint` to get an SPL Associated Token Account.
// Empty mint returns the wallet owner address (native SOL).
ata, err := c.Wallet.DepositAddress(ctx, w.WalletID,
    fintech.NetworkSolana, fintech.MintUSDCSolanaMainnet)
sol, err := c.Wallet.DepositAddress(ctx, w.WalletID,
    fintech.NetworkSolana, "")

// Withdraw — `Mint` selects SPL token, empty = native SOL.
tx, err := c.Wallet.Withdraw(ctx, fintech.WithdrawRequest{
    WalletID: w.WalletID, Amount: "10.50",
    DestinationType: "crypto_address",
    DestinationID:   "RecipientSolanaAddr",
    Network:         fintech.NetworkSolana,
    Mint:            fintech.MintUSDCSolanaMainnet,
}, "idem-w-1")

// Compliance / fraud
_, err = c.Wallet.Freeze(ctx, w.WalletID, "fraud-investigation")
_, err = c.Wallet.Unfreeze(ctx, w.WalletID, "case-resolved")

// Pagination — opaque cursor, empty page_token = first page.
resp, err := c.Wallet.List(ctx, "user-123", 50, "")
for resp.NextCursor != nil && *resp.NextCursor != "" {
    resp, err = c.Wallet.List(ctx, "user-123", 50, *resp.NextCursor)
    // ...
}

// Single transaction lookup
single, err := c.Wallet.GetTransaction(ctx, w.WalletID, "tx-xxx")
```

### Error handling

```go
payment, err := client.GetPayment(ctx, paymentID)
if err != nil {
    switch e := err.(type) {
    case *mg.NotFoundError:
        // payment does not exist
    case *mg.UnauthorizedError:
        // bad or expired API key
    case *mg.ValidationError:
        // invalid request field: e.Field, e.Message
    case *mg.APIError:
        // other 4xx/5xx: e.StatusCode, e.Body
    default:
        // network error
    }
}
```

## Supported currencies

| Code | Name |
|------|------|
| `UZS` | Uzbekistani So'm |
| `KZT` | Kazakhstani Tenge |
| `TJS` | Tajikistani Somoni |
| `KGS` | Kyrgyzstani Som |
| `USD` | US Dollar |
| `EUR` | Euro |

## Supported payment methods

| Method | BIN / Identifier | Region |
|--------|-----------------|--------|
| Uzcard | `8600xxxx` | Uzbekistan |
| Humo | `9860xxxx` | Uzbekistan |
| Click | wallet provider | Uzbekistan |
| Payme | wallet provider | Uzbekistan |
| Oson | wallet provider | Uzbekistan |
| Visa | `4xxxxxxx` | International |
| Mastercard | `51–55`, `2221–2720` | International |

## Event types

```go
mg.EventPaymentCreated             // "payment.created"
mg.EventPaymentAuthorized          // "payment.authorized"
mg.EventPaymentAuthorizationFailed // "payment.authorization_failed"
mg.EventPaymentCaptured            // "payment.captured"
mg.EventPaymentCaptureFailed       // "payment.capture_failed"
mg.EventPaymentVoided              // "payment.voided"
mg.EventPaymentFailed              // "payment.failed"
mg.EventRefundRequested            // "refund.requested"
mg.EventRefundSettled              // "refund.settled"
mg.EventRefundFailed               // "refund.failed"
mg.EventCheckoutCompleted          // "checkout.completed"
mg.EventCheckoutExpired            // "checkout.expired"
```

## BaaS Services

### Chat

```go
// Create a channel
channel, err := client.Chat.CreateChannel(ctx, mg.CreateChannelRequest{
    TenantID:  "tenant-uuid",
    ChannelID: "booking-123",
    Name:      "Booking Discussion",
})

// Send a message
msg, err := client.Chat.SendMessage(ctx, channel.ChannelID, mg.SendMessageRequest{
    TenantID: "tenant-uuid",
    SenderID: "user-uuid",
    Content:  "Hello, is the apartment available?",
})

// List messages (cursor-paginated)
msgs, err := client.Chat.ListMessages(ctx, channel.ChannelID, mg.ListMessagesParams{
    TenantID: "tenant-uuid",
    Limit:    50,
})
```

### Notify

```go
// Send SMS
err := client.Notify.SendSMS(ctx, mg.SendSMSRequest{
    TenantID: "tenant-uuid",
    To:       "+998901234567",
    Text:     "Your booking is confirmed!",
})

// Send email via template
err = client.Notify.SendEmail(ctx, mg.SendEmailRequest{
    TenantID:    "tenant-uuid",
    To:          "guest@example.com",
    TemplateKey: "booking_confirmed",
    Vars:        map[string]string{"name": "Said", "checkin": "2026-04-01"},
})
```

### Storage

```go
// Generate presigned upload URL
upload, err := client.Storage.GenerateUploadURL(ctx, mg.GenerateUploadURLRequest{
    TenantID: "tenant-uuid",
    Filename: "listing-photo.jpg",
    MimeType: "image/jpeg",
})
// → upload.UploadUrl, upload.FileID

// Get download URL
download, err := client.Storage.GetDownloadURL(ctx, fileID, "tenant-uuid")
```

### Feature Flags

```go
// Evaluate a flag for a user
result, err := client.Flags.Evaluate(ctx, mg.EvaluateFlagRequest{
    TenantID: "tenant-uuid",
    FlagKey:  "instant_book_v2",
    UserID:   "user-123",
})
if result.Enabled {
    // new feature path
}
```

### Guard (Rate Limiting)

```go
// Check rate limit
check, err := client.Guard.Check(ctx, mg.GuardCheckRequest{
    TenantID: "tenant-uuid",
    Path:     "/v1/payments",
    Method:   "POST",
    IP:       "203.0.113.42",
})
if !check.Allowed {
    // rate limited — check.Reason, check.ResetAt
}
```

### Invoices

```go
invoice, err := client.Invoices.Create(ctx, mg.CreateInvoiceRequest{
    TenantID:   "tenant-uuid",
    CustomerID: "cus-123",
    Amount:     "500000.00",
    Currency:   "UZS",
    LineItems: []mg.InvoiceLineItem{
        {Description: "3 nights — Apartment", Quantity: 3, UnitAmount: "166666.67"},
    },
})
err = client.Invoices.Send(ctx, invoice.ID)
```

### Logs

```go
// Query audit logs
page, err := client.Logs.QueryAudit(ctx, mg.LogsQuery{
    TenantID: "tenant-uuid",
    From:     "2026-02-01T00:00:00Z",
    To:       "2026-02-25T00:00:00Z",
    Limit:    100,
})
```

## Full API documentation

See [`docs/api/README.md`](../../docs/api/README.md) in the Mashgate repository.

## Running tests

```sh
cd packages/sdk-go
go test ./... -v -race
```
