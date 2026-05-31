# Billing & Subscriptions

Mashgate's Billing layer handles platform-level recurring revenue: subscription
plans, plan changes with proration, billing invoices, payment methods on file,
credits & promo codes (`billing`); plus a lower-level subscription primitive for
your own per-customer recurring plans (`subscriptions`) and usage tracking that
feeds billing (`metering`).

There are two distinct surfaces here:

- **`billing`** — the *tenant's own* subscription to a Mashgate platform plan
  (one active subscription per tenant), payment methods, billing invoices, credits.
- **`subscriptions`** — recurring plans *you* define and bill *your* customers on.

## When to use

- You manage the tenant's platform subscription, billing invoices, and payment
  methods on file (`billing`).
- You preview proration before a plan change (`billing.PreviewPlanChange`).
- You sell recurring plans to *your own* customers — create plans, subscribe
  customers, pause/resume/cancel (`subscriptions`).
- You report custom usage (e.g. AI tokens) that feeds billing (`metering`).

## Key operations

### Platform billing (`billing`)

Read plans, manage the tenant's single active subscription, and preview proration
before applying a change.

**Go**

```go
plans, err := client.Billing.ListPlans(ctx)
sub, err := client.Billing.GetSubscription(ctx)

// Preview proration without applying.
preview, err := client.Billing.PreviewPlanChange(ctx, mashgate.PreviewPlanChangeRequest{
    NewPlanID: "plan_pro",
    Effective: "immediate", // or "period_end"
})

// Apply the change.
sub, err = client.Billing.ChangePlan(ctx, mashgate.ChangePlanRequest{
    NewPlanID: "plan_pro",
    Prorate:   true,
    Effective: "immediate",
})

sub, err = client.Billing.CancelPlan(ctx, mashgate.CancelPlanRequest{
    Reason:            "downgrading",
    CancelImmediately: false, // effective at period end
})
```

**TypeScript**

```ts
const plans = await client.billing.listPlans();
const sub = await client.billing.getSubscription();

const preview = await client.billing.previewPlanChange({ planId: "plan_pro" });
await client.billing.changePlan({ planId: "plan_pro", immediate: true });
await client.billing.cancelPlan({ reason: "downgrading" });
```

Python: not yet available for this module — use Go or TypeScript.

### Billing payment methods, invoices & credits (`billing`)

**Go**

```go
methods, err := client.Billing.ListPaymentMethods(ctx)
m, err := client.Billing.AddPaymentMethod(ctx, mashgate.AddBillingPaymentMethodRequest{
    Type:      "card",
    Card:      &mashgate.CardPaymentMethod{Token: "tok_...", Brand: "visa", Last4: "4242"},
    IsDefault: true,
})
m, err = client.Billing.SetDefaultPaymentMethod(ctx, m.ID)
err = client.Billing.RemovePaymentMethod(ctx, m.ID)

invoices, err := client.Billing.ListInvoices(ctx)
inv, err := client.Billing.GetInvoice(ctx, invoiceID)
inv, err = client.Billing.PayInvoice(ctx, invoiceID) // immediate payment attempt

bal, err := client.Billing.GetCreditBalance(ctx)
redeemed, err := client.Billing.RedeemPromoCode(ctx, "LAUNCH50")
```

**TypeScript**

```ts
const methods = await client.billing.listPaymentMethods();
await client.billing.addPaymentMethod({ token: "tok_...", brand: "visa", last4: "4242", setDefault: true });
await client.billing.setDefaultPaymentMethod(methodId);

const invoices = await client.billing.listInvoices();
await client.billing.payInvoice(invoiceId);

const balance = await client.billing.getCreditBalance();
await client.billing.redeemPromoCode("LAUNCH50");
```

Python: not yet available for this module — use Go or TypeScript.

### Customer subscriptions (`subscriptions`)

Define recurring plans and subscribe *your* customers. A `paymentMethodToken` binds
the billing source. Amounts are minor-unit `int64`.

**Go**

```go
plan, err := client.Subscriptions.CreatePlan(ctx, mashgate.CreatePlanRequest{
    TenantID:  tenantID,
    Name:      "Pro Monthly",
    Amount:    9900,      // minor units
    Currency:  "UZS",
    Interval:  "monthly", // "monthly" | "yearly" | "weekly"
    TrialDays: 14,
})

sub, err := client.Subscriptions.Create(ctx, mashgate.CreateSubscriptionRequest{
    TenantID:           tenantID,
    CustomerID:         customerID,
    PlanID:             plan.ID,
    PaymentMethodToken: "tok_...",
})

subs, err := client.Subscriptions.List(ctx, tenantID)
sub, err = client.Subscriptions.Pause(ctx, sub.ID)
sub, err = client.Subscriptions.Resume(ctx, sub.ID)
sub, err = client.Subscriptions.Cancel(ctx, sub.ID)
```

**TypeScript**

```ts
const plan = await client.subscriptions.createPlan({
  tenantId,
  name: "Pro Monthly",
  amount: 9900,
  currency: "UZS",
  interval: "monthly",
  trialDays: 14,
});

const sub = await client.subscriptions.create({
  tenantId,
  customerId,
  planId: plan.id,
  paymentMethodToken: "tok_...",
});
await client.subscriptions.pause(sub.id);
await client.subscriptions.resume(sub.id);
await client.subscriptions.cancel(sub.id);
```

Python: not yet available for this module — use Go or TypeScript.

### Usage metering (`metering`)

The platform records billable events automatically on Payment/Storage/Chain calls.
Use `metering` only for custom meters (e.g. AI tokens consumed). `RecordUsage` is
**idempotent** — pass a stable key (auto-generated if empty in the Go SDK).

**Go**

```go
rec, err := client.Metering.RecordUsage(ctx, mashgate.RecordUsageRequest{
    TenantID:       tenantID,
    MeterCode:      "ai_tokens", // platform-known or tenant-custom meter
    Quantity:       1280,
    IdempotencyKey: "req:" + requestID + ":tokens",
})

records, err := client.Metering.ListUsage(ctx, mashgate.ListUsageParams{
    TenantID:  tenantID,
    MeterCode: "ai_tokens",
    From:      from,
    To:        to,
})

summary, err := client.Metering.GetUsageSummary(ctx, tenantID, from, to)
```

**TypeScript**

The TS `metering` resource is read-oriented (usage summary, time series, quota
status by resource):

```ts
const summary = await client.metering.getUsageSummary(tenantId);
const series = await client.metering.getUsageTimeSeries(tenantId, "API_CALLS", start, end);
const quota = await client.metering.getQuotaStatus(tenantId);
```

Python: not yet available for this module — use Go or TypeScript.

## Best practices

- **Pass a stable idempotency key when recording usage** so a retried request never
  double-counts (`req:<id>:tokens`, not `now()`).
- **Preview plan changes** (`PreviewPlanChange`) before applying so the UI can show
  the prorated charge.
- **Let the platform meter built-in resources** (payments, storage, chain) — only
  call `RecordUsage` for genuinely custom meters.
- **Carry `tenant_id` through your own subscription records** and make it part of
  your uniqueness constraints.

See [Best practices](../best-practices.md) (§2 idempotency, §4 multi-tenancy,
§1 money ownership).

## See also

- [Building a vertical](../guides/building-a-vertical.md)
- [Best practices](../best-practices.md)
- [Service catalog](./service-catalog.md)
- [Payments & checkout](./payments-checkout.md)
