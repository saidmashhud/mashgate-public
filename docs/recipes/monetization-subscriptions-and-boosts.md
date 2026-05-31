# Recipe: Premium subscriptions + one-off boosts

**Goal:** sell a recurring *premium* tier **and** one-off purchases (a "boost"
that pushes your profile to the top, or "super-likes" you spend from a coin
balance) — composing checkout, billing, the wallet ledger, and HookLine events.

Running example: a dating app. The same pattern fits any vertical that mixes a
subscription with consumable credits.

---

## What you'll use

| Concern | Mashgate module | SDK surface |
|---------|-----------------|-------------|
| Sign-up / login, identity | **Identity (mgID)** | `client.Login` / `client.Register` |
| One-off purchase (buy a boost / a coin pack) | **Checkout** | `client.CreateCheckout` |
| Recurring premium tier | **Billing** (platform plans) or **Subscriptions** (per-customer plans) | `client.Billing.*` / `client.Subscriptions.*` |
| In-app coins / consumable balance | **Wallets & Ledger** (the ledger is the money SoT) | `fintech.Client.Wallet.*` |
| React to a completed payment, reconcile coins | **Events / HookLine** | `client.ConstructEvent` + your handler |

**What *your* vertical owns** (per [building a vertical](../guides/building-a-vertical.md)):
the *meaning* of premium and boosts. Mashgate moves the money and holds the
balance; you own `entitlements` (is this user premium right now?),
`boost_orders`, and what a "super-like" *does*. Key every one of those rows by
the mgID `user_id` — never copy Mashgate's authoritative balance into your DB as
a source of truth. See [data modeling & identity](../guides/data-modeling-and-identity.md).

> **Money SoT:** the ledger is authoritative for coin balances and movements.
> Your `entitlements`/`boost_orders` tables are projections you reconcile from
> `payment.captured` / `wallet.credit` events — not the place balances live.

---

## Flow

```
                         ┌─────────────── one-off "boost" / coin pack ───────────────┐
                         │                                                            │
signup ──► login ──► [user picks]                                                     │
                         │                                                            ▼
                         │                                          CreateCheckout(amount, idempotencyKey)
                         │                                                            │
                         │                                          redirect → hosted checkout → pays
                         │                                                            │
                         ▼                                                            ▼
              premium tier?                                          HookLine ── payment.captured ──►
                         │                                          your /webhooks (verify + dedupe)
            Billing.ChangePlan / Subscriptions.Create                            │
                         │                                          ┌─────────────┴─────────────┐
                         ▼                                          ▼                           ▼
            HookLine ── payment.captured ──►              grant entitlement            Wallet.Credit coins
            mark user premium                            (mark boost active)        (ledger = money SoT)
                                                                    │                           │
                                                                    └────────► your DB row reconciled ◄
```

End-to-end: **signup → login → purchase (checkout or subscription) → webhook →
unlock**. Coins are spent later via `Wallet.Debit`; the user's balance view is
always derived from the ledger, reconciled by `wallet.credit` / `wallet.debit`
events — never poll.

---

## Implementation

### 0. Init the clients

Two clients: the main client (auth, checkout, billing, subscriptions, webhooks)
and the **fintech** client (the wallet ledger lives in the Fintech Pack).

```go
import (
    "github.com/saidmashhud/mashgate-public/sdk/go"          // package mashgate
    "github.com/saidmashhud/mashgate-public/sdk/go/fintech"
)

mg := mashgate.New(os.Getenv("MASHGATE_BASE_URL"), os.Getenv("MASHGATE_API_KEY"))
fin := fintech.New(os.Getenv("MASHGATE_BASE_URL"), tenantID, os.Getenv("MASHGATE_API_KEY"))
```

```ts
import { MashgateClient } from "@mashgate/sdk";

const mg = new MashgateClient({
  baseUrl: process.env.MASHGATE_BASE_URL!,
  apiKey:  process.env.MASHGATE_API_KEY!,
});
// TS: checkout / billing / subscriptions are first-class resources.
// The wallet *ledger* (Wallet.Credit/Debit) is Fintech-Pack and Go-only today —
// call it from your Go service, or hit the REST endpoints directly from TS.
```

> **Python:** `checkout` and `wallet` (saved cards / balance read) exist, but
> there is no `billing`, `subscriptions`, or wallet-**ledger** resource. For the
> subscription and coin-ledger steps below, **use Go/TS**.

Take `userId`/`tenantId` from the validated token, not the request body
([best practices §5](../best-practices.md)).

### 1. One-off purchase — a boost or a coin pack (Checkout)

`CreateCheckout` returns a session with a `CheckoutURL`; redirect the buyer there.
**Pass an idempotency key derived from your order id** so a retry never creates a
second session ([best practices §2](../best-practices.md)).

```go
// your DB: INSERT boost_orders (id, mg_user_id, tenant_id, status='pending') ...
session, err := mg.CreateCheckout(ctx, mashgate.CreateCheckoutRequest{
    TotalAmount: mashgate.Money{Amount: "25000.00", Currency: "UZS"}, // one "boost"
    Items: []mashgate.LineItem{{
        Name:      "Profile Boost (30 min)",
        Quantity:  1,
        UnitPrice: mashgate.Money{Amount: "25000.00", Currency: "UZS"},
    }},
    SuccessURL:    "https://yourapp.example/boost/success",
    CancelURL:     "https://yourapp.example/boost/cancel",
    CustomerID:    userID,                 // mgID user_id
    Metadata:      map[string]string{"boost_order_id": orderID, "kind": "boost"},
    IdempotencyKey: "boost:" + orderID,    // stable per operation, not per attempt
})
if err != nil {
    return err
}
// redirect the customer to session.CheckoutURL
```

```ts
const session = await mg.checkout.createSession({
  totalAmount: { amount: "25000.00", currency: "UZS" },
  items: [{ name: "Profile Boost (30 min)", quantity: 1,
            unitPrice: { amount: "25000.00", currency: "UZS" } }],
  successUrl: "https://yourapp.example/boost/success",
  cancelUrl:  "https://yourapp.example/boost/cancel",
  customerId: userId,
  metadata: { boost_order_id: orderId, kind: "boost" },
});
// redirect to session.checkoutUrl
```

For a **coin pack** it's the same call with `kind: "coins"` and the coin count in
metadata — the difference is what your webhook handler does on capture (below).

### 2. Premium tier — recurring (Billing or Subscriptions)

Two ways, depending on whether premium is a **platform plan** (priced in the
Mashgate console) or a **plan you define per customer**.

**Platform plans (Billing):**

```go
plans, _ := mg.Billing.ListPlans(ctx)            // discover "premium" plan id
sub, err := mg.Billing.ChangePlan(ctx, mashgate.ChangePlanRequest{
    NewPlanID: premiumPlanID,
    Prorate:   true,
    Effective: "immediate",
})
// sub.Status: "active" | "trialing" | "past_due" | "canceled"
```

```ts
const plans = await mg.billing.listPlans();
const sub = await mg.billing.changePlan({ planId: premiumPlanId, immediate: true });
```

Preview the proration before you charge, for a confirmation UI:

```go
preview, _ := mg.Billing.PreviewPlanChange(ctx, mashgate.PreviewPlanChangeRequest{
    NewPlanID: premiumPlanID, Effective: "immediate",
})
// preview.ProrationCents, preview.NextChargeAt
```

**Per-customer plans (Subscriptions):**

```go
plan, _ := mg.Subscriptions.CreatePlan(ctx, mashgate.CreatePlanRequest{
    TenantID: tenantID, Name: "Premium", Amount: 4900000, Currency: "UZS",
    Interval: "monthly", TrialDays: 7,
})
sub, _ := mg.Subscriptions.Create(ctx, mashgate.CreateSubscriptionRequest{
    TenantID: tenantID, CustomerID: userID, PlanID: plan.ID,
    PaymentMethodToken: pmToken, // a saved card token from the wallet
})
// later: mg.Subscriptions.Cancel / Pause / Resume
```

```ts
const plan = await mg.subscriptions.createPlan({
  tenantId, name: "Premium", amount: 4900000, currency: "UZS",
  interval: "monthly", trialDays: 7,
});
const sub = await mg.subscriptions.create({
  tenantId, customerId: userId, planId: plan.id, paymentMethodToken: pmToken,
});
```

> Note `Subscriptions.CreatePlan` amount is **minor units `int64`** (`4900000` =
> 49,000.00 UZS), while `Checkout`/`Money` uses a **decimal string** (`"25000.00"`).
> Don't mix the two.

### 3. The webhook handler — verify, dedupe, then unlock

Everything above only *requests* money movement. You **unlock on the event**, not
on the API response, because the hosted-checkout payment completes out of band.
Verify the signature, dedupe on the event id, return 2xx fast
([best practices §3](../best-practices.md)).

```go
func handleWebhook(mg *mashgate.Client, fin *fintech.Client) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        body, _ := io.ReadAll(r.Body)
        event, err := mg.ConstructEvent(
            body,
            r.Header.Get("x-hl-signature"),
            r.Header.Get("x-hl-timestamp"),
            os.Getenv("MASHGATE_WEBHOOK_SECRET"),
        )
        if err != nil { // bad/old signature → reject
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }

        // Idempotent: an already-processed event is a no-op, not an error.
        if seenBefore(event.CanonicalID()) {
            w.WriteHeader(http.StatusOK)
            return
        }

        switch event.EventType { // tolerate legacy + envelope-v1 topics
        case mashgate.EventPaymentCaptured, mashgate.TopicPaymentCompleted:
            var p struct {
                PaymentID string            `json:"paymentId"`
                Metadata  map[string]string `json:"metadata"`
            }
            _ = json.Unmarshal(event.PayloadBytes(), &p)

            switch p.Metadata["kind"] {
            case "boost":
                // grant the entitlement in YOUR db, keyed by mg user_id
                activateBoost(p.Metadata["boost_order_id"])
            case "coins":
                // credit coins into the LEDGER (money SoT), idempotent on payment id
                _, _ = fin.Wallet.Credit(r.Context(), fintech.CreditRequest{
                    WalletID:      walletIDFor(p.Metadata["mg_user_id"]),
                    Amount:        p.Metadata["coin_amount"], // decimal string
                    Reason:        fintech.ReasonDeposit,
                    ReferenceID:   p.PaymentID,
                    ReferenceType: "checkout_payment",
                    Description:   "coin pack purchase",
                }, "credit:"+p.PaymentID) // idempotency key = payment id
            }
        }
        markSeen(event.CanonicalID())
        w.WriteHeader(http.StatusOK) // ack fast; do slow work async
    }
}
```

```ts
// mg.webhooks.on(...) + handleRequest verifies the signature and routes by topic.
mg.webhooks.on("payment.captured", async (event) => {
  const p = eventPayload(event) as { paymentId: string; metadata: Record<string,string> };
  if (alreadyProcessed(event.id ?? event.eventId)) return;     // dedupe
  if (p.metadata.kind === "boost") await activateBoost(p.metadata.boost_order_id);
  // coin credit (ledger) is Go-only today — enqueue a job your Go worker drains,
  // or call the /v1/wallets/{id}/credit endpoint directly with an Idempotency-Key.
});

app.post("/webhooks", express.raw({ type: "*/*" }), async (req, res) => {
  await mg.webhooks.handleRequest(
    req.body.toString(),
    req.headers["x-hl-signature"] as string,
    process.env.MASHGATE_WEBHOOK_SECRET!,
    req.headers["x-hl-timestamp"] as string,
  );
  res.json({ ok: true });   // 2xx fast — HookLine retries on non-2xx
});
```

### 4. Spend coins later — `Wallet.Debit`

When the user spends a super-like, debit the ledger with an idempotency key tied
to the action so a double-tap can't double-spend:

```go
tx, err := fin.Wallet.Debit(ctx, fintech.DebitRequest{
    WalletID:      walletID,
    Amount:        "1",                         // 1 super-like
    Reason:        fintech.ReasonPayment,
    ReferenceID:   superLikeActionID,           // your domain id
    ReferenceType: "super_like",
    Description:   "super-like spend",
}, "debit:"+superLikeActionID)
// tx.BalanceAfter is the ledger's post-debit balance — render from this, not a local column.
```

### 5. Read the balance from the ledger (don't cache as SoT)

```go
wallet, _ := fin.Wallet.Get(ctx, walletID) // wallet.Balance / wallet.Pending (minor-unit strings)
```

The end-user balance view (`client.GetWalletBalance`) and the ledger
(`fintech.Wallet.Get` + `ListTransactions`) both derive from the same source of
truth. If you keep a local "coins" number for UI speed, treat it strictly as a
cache and reconcile it from `wallet.credit` / `wallet.debit` events.

---

## Gotchas / best practices

- **Unlock on the event, not the API response.** Hosted checkout completes out of
  band; the `CreateCheckout` return only means the session was created.
- **Idempotency keys on every money op** — derive from your domain id
  (`boost:<orderId>`, `credit:<paymentId>`), never from `now()`/random. → [§2](../best-practices.md)
- **Idempotent webhook handlers** — dedupe on `CanonicalID()`/`event.id`; the same
  event can arrive more than once and out of order. → [§3](../best-practices.md)
- **Ledger is the money SoT** — never keep your own authoritative balance column;
  reconcile coins from `wallet.credit`/`wallet.debit`. → [§1](../best-practices.md)
- **Mind the amount units** — `Money`/`Checkout` use decimal strings (`"25000.00"`);
  `Subscriptions.CreatePlan`/billing cents use minor-unit `int64`.
- **Tolerate both event forms** — match legacy `EventPaymentCaptured` *and*
  envelope-v1 `TopicPaymentCompleted`; use `PayloadBytes()` to read either shape.
- **Keep the signing secret server-side**; it's returned only once on
  `CreateWebhookEndpoint`. → [§7](../best-practices.md)

---

## See also

- [Building a vertical](../guides/building-a-vertical.md) — the module/domain split.
- [Data modeling & identity](../guides/data-modeling-and-identity.md) — key your
  `entitlements`/`boost_orders` by `mg_user_id`.
- [Best practices](../best-practices.md) — idempotency, money, webhooks.
- [Service catalog](../modules/service-catalog.md) — checkout, billing,
  subscription-service, ledger-core.
- [KYC gate (18+) recipe](./kyc-gate-18plus.md) — gate purchases behind age/identity.
