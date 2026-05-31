# Payments & Checkout

Mashgate's Payments layer moves money for your tenant: hosted checkout sessions, a
low-level payment-intent API (authorize / capture / void / refund), shareable
payment links, country-specific local rails (Click, Payme, Uzcard, Humo, …), and
merchant invoices. This guide covers the `checkout`, `payments`, `payment_links`,
`local_payments`, and `invoices` resources.

> Money is owned by the ledger — treat any local balance you keep as a cache and
> reconcile against Mashgate via events.

## When to use

- You want a **hosted checkout page** — create a session and redirect the customer.
- You need fine-grained control over the payment lifecycle (auth-then-capture, voids,
  partial refunds) via the **payment-intent** API.
- You want to send a customer a **shareable payment link**.
- You accept **local payment methods** (TJ: Tcell, Korti Milli, Alif, Eskhata;
  UZ: Click, Payme, Apelsin / Uzcard / Humo).
- You issue **invoices** to customers.

## Key operations

### Hosted checkout

`CreateCheckout` returns a session with a `CheckoutURL` — redirect the customer
there. Creation is **idempotent**: the Go/Python SDKs auto-generate an
`Idempotency-Key` header if you don't supply one, but you should pass a stable key
derived from your order so retries never create a duplicate session.

**Go**

```go
session, err := client.CreateCheckout(ctx, mashgate.CreateCheckoutRequest{
    TotalAmount:    mashgate.Money{Amount: "2500000.00", Currency: "UZS"},
    SuccessURL:     "https://zist.uz/booking/success",
    CancelURL:      "https://zist.uz/booking/cancel",
    CustomerEmail:  "guest@example.com",
    IdempotencyKey: "order:" + orderID + ":checkout", // stable per operation
})
if err != nil {
    return err
}
// redirect customer to session.CheckoutURL

got, err := client.GetCheckout(ctx, session.SessionID)
err = client.ExpireCheckout(ctx, session.SessionID) // invalidate before use
```

**TypeScript**

```ts
const session = await client.checkout.createSession({
  amount: "2500000",
  currency: "UZS",
  successUrl: "https://zist.uz/booking/success",
  cancelUrl: "https://zist.uz/booking/cancel",
});
// redirect to session.checkoutUrl
const got = await client.checkout.getSession(session.sessionId);
```

**Python**

```python
session = client.checkout.create_session(
    success_url="https://zist.uz/booking/success",
    cancel_url="https://zist.uz/booking/cancel",
    line_items=[{"name": "Booking", "quantity": 1, "unitPrice": {"amount": "2500000.00", "currency": "UZS"}}],
    currency="UZS",
)
got = client.checkout.get_session(session["sessionId"])
```

`CompleteCheckout` is normally invoked by Mashgate's hosted checkout page, not by
your backend directly.

### Payment intents (authorize / capture / void / refund)

Use the payment-intent API when you need control over the lifecycle. Set
`CaptureMode: "MANUAL"` to authorize first and capture later; otherwise `"AUTO"`
captures on authorize. `CreatePayment` and `RefundPayment` are idempotent — pass a
stable key (auto-generated if empty in the Go SDK).

**Go**

```go
p, err := client.CreatePayment(ctx, mashgate.CreatePaymentRequest{
    Amount:         mashgate.Money{Amount: "150000.00", Currency: "UZS"},
    OrderID:        orderID,
    CaptureMode:    "MANUAL",
    IdempotencyKey: "order:" + orderID + ":charge",
})

p, err = client.AuthorizePayment(ctx, p.PaymentID) // pending -> authorized
p, err = client.CapturePayment(ctx, p.PaymentID)   // authorized -> captured
// or release the hold:
p, err = client.VoidPayment(ctx, p.PaymentID)      // authorized -> voided

// Partial refund on a captured payment:
p, err = client.RefundPayment(ctx, p.PaymentID, mashgate.RefundRequest{
    Amount:         mashgate.Money{Amount: "50000.00", Currency: "UZS"},
    Reason:         "partial_cancel",
    IdempotencyKey: "order:" + orderID + ":refund:1",
})

list, err := client.ListPayments(ctx, mashgate.ListPaymentsParams{
    Status: "captured", Page: 1, PageSize: 20,
})
```

**TypeScript**

```ts
const intent = await client.payments.create({
  amount: "150000",
  currency: "UZS",
  orderId,
  idempotencyKey: `order:${orderId}:charge`,
});

await client.payments.authorize(intent.paymentId, { idempotencyKey: `order:${orderId}:auth` });
await client.payments.capture(intent.paymentId, { idempotencyKey: `order:${orderId}:capture` });
await client.payments.void(intent.paymentId);

await client.payments.refund(intent.paymentId, {
  amount: "50000",
  currency: "UZS",
  idempotencyKey: `order:${orderId}:refund:1`,
});
```

**Python**

```python
intent = client.payments.create(
    amount="150000", currency="UZS", order_id=order_id,
    idempotency_key=f"order:{order_id}:charge",
)
client.payments.authorize(intent["paymentId"], idempotency_key=f"order:{order_id}:auth")
client.payments.capture(intent["paymentId"], idempotency_key=f"order:{order_id}:capture")
client.payments.refund(
    intent["paymentId"], amount="50000", currency="UZS",
    idempotency_key=f"order:{order_id}:refund:1",
)
```

### Payment links

Shareable, single-use payment URLs.

**Go**

```go
link, err := client.PaymentLinks.Create(ctx, mashgate.CreatePaymentLinkRequest{
    TenantID:    tenantID,
    Amount:      150000, // minor units (int64)
    Currency:    "UZS",
    Description: "Invoice #42",
})
// share link.URL

links, err := client.PaymentLinks.List(ctx, tenantID)
got, err := client.PaymentLinks.Get(ctx, link.ID)
```

**TypeScript**

```ts
const link = await client.paymentLinks.create({
  tenantId,
  amount: 150000,
  currency: "UZS",
  description: "Invoice #42",
});
const links = await client.paymentLinks.list(tenantId);
```

Python: not yet available for this module — use Go or TypeScript.

### Local payments

Country-specific rails. `InitiatePayment` returns a provider-specific `NextStep`
(redirect URL, USSD code, or QR); `ConfirmPayment` submits the provider callback
(SMS OTP, etc.). Initiation is idempotent.

> Status: providers may run in mock mode until the platform wires real credentials
> — see the [service catalog](./service-catalog.md).

**Go**

```go
methods, err := client.LocalPayments.ListSupportedMethods(ctx, tenantID, "TJ")

initiated, err := client.LocalPayments.InitiatePayment(ctx, mashgate.InitiateLocalPaymentRequest{
    TenantID:       tenantID,
    MethodID:       "korti-milli",
    Amount:         mashgate.Money{Amount: "150000.00", Currency: "TJS"},
    OrderID:        orderID,
    IdempotencyKey: "order:" + orderID + ":local",
})
// act on initiated.NextStep (redirect / ussd / qr / otp)

confirmed, err := client.LocalPayments.ConfirmPayment(ctx, initiated.PaymentID,
    mashgate.ConfirmLocalPaymentRequest{OTP: "123456"})

status, err := client.LocalPayments.GetPaymentStatus(ctx, initiated.PaymentID)
err = client.LocalPayments.CancelPayment(ctx, initiated.PaymentID, "customer_abandoned")
```

**TypeScript**

The TS `localPayments` resource models card (Uzcard / Humo, OTP) and wallet
(Click / Payme / Oson, redirect) flows directly:

```ts
const pay = await client.localPayments.payByCard({
  provider: "uzcard",
  cardNumber: "8600...",
  expiryDate: "12/27",
  amount: "150000",
  currency: "UZS",
  orderId,
});
await client.localPayments.confirmOtp({ paymentId: pay.paymentId, otpCode: "123456" });

const wallet = await client.localPayments.payByWallet({
  provider: "click",
  amount: "150000",
  orderId,
  returnUrl: "https://zist.uz/return",
});
// redirect to wallet.redirectUrl
```

Python: not yet available for this module — use Go or TypeScript.

### Invoices

**Go**

```go
inv, err := client.Invoices.Create(ctx, mashgate.CreateInvoiceRequest{
    TenantID: tenantID,
    Amount:   150000, // minor units
    Currency: "UZS",
    LineItems: []mashgate.InvoiceLineItem{
        {Description: "Consulting", Quantity: 1, UnitAmount: 150000},
    },
})

invoices, err := client.Invoices.List(ctx, tenantID, "open")
sent, err := client.Invoices.Send(ctx, inv.ID)   // email via notify-service
voided, err := client.Invoices.Void(ctx, inv.ID)
```

**TypeScript**

```ts
const inv = await client.invoices.create({
  tenantId,
  amount: 150000,
  currency: "UZS",
  lineItems: [{ description: "Consulting", quantity: 1, unitAmount: 150000 }],
});
await client.invoices.send(inv.id);
```

Python: not yet available for this module — use Go or TypeScript.

## Events

Payments emit HookLine events you can subscribe to (envelope v1 topics, with legacy
dotted forms also emitted during migration):

- `payments.payment.created` / `payment.created`
- `payments.payment.authorized` / `payment.authorized`
- `payments.payment.completed`, `payment.captured`
- `payments.payment.failed` / `payment.failed`, `payment.voided`
- `payments.refund.created` / `refund.requested`, `payments.refund.completed` /
  `refund.settled`, `payments.refund.failed`
- `payments.checkout_session.created`, `payments.checkout_session.completed` /
  `checkout.completed`, `checkout.expired`

Use the `Event*` / `Topic*` constants in the Go SDK instead of raw strings.
Make handlers idempotent and dedupe on the event id.

## Best practices

- **Pass a stable idempotency key on every money operation** (create, refund,
  capture) — derive it from your domain (`order:123:charge`), not from `now()`.
- **Don't keep your own authoritative balance** — read from Mashgate and reconcile
  on events; the ledger is the source of truth.
- **Use amounts as decimal strings** (`Money{Amount, Currency}`) for orchestrated
  flows; payment-link / invoice amounts are minor-unit `int64`.
- **React to events, don't poll** — subscribe to `payment.*` / `checkout.*` and
  return 2xx fast.

See [Best practices](../best-practices.md) (§2 idempotency, §1 money ownership,
§3 events).

## See also

- [Building a vertical](../guides/building-a-vertical.md)
- [Best practices](../best-practices.md)
- [Service catalog](./service-catalog.md)
- [Billing & subscriptions](./billing-subscriptions.md)
