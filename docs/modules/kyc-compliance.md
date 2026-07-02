# KYC & Compliance

Identity verification, AML/sanctions/PEP screening, merchant onboarding, and
fraud scoring — the Mashgate **fintech pack** plus the `risk` and `guard`
resources. Use these to gate access, onboard merchants, and assess transactions
before they touch the ledger.

## When to use

- You must verify a customer's or business's identity before they transact
  (KYC), or run AML / sanctions / PEP checks.
- You onboard merchants through an accept/reject/suspend lifecycle gated on KYC.
- You need fraud risk scores on transactions, or a per-tenant blocklist.
- You raise, triage, escalate, or resolve compliance alerts (and gate
  withdrawals on open alerts).
- You enforce per-tenant rate limits and IP blocklists at the edge (`guard`).

Identity and age gating (18+, jurisdiction rules) are **your** responsibility —
Mashgate provides the verification primitives and screening signals; your
vertical decides what to allow. See [Best practices](#best-practices).

## Key operations

KYC, Compliance, and Merchant are exposed by the Go SDK on the unified
tenant client (`mashgate.NewWithTenant(baseURL, tenantID, apiKey)`). Import
`sdk/go/fintech` for request and enum types. The `risk` and `guard` resources
are available as normal client resources across SDKs.

### KYC checks

`Request` submits a check; if the provider needs a hosted flow, the response
carries a `RedirectURL` you must surface to the user. Screening is **async** —
poll with `Get` / `List` or react to `kyc.*` events; a fresh check starts
`KYC_STATUS_PENDING`.

**Go**

```go
mg := mashgate.NewWithTenant(baseURL, tenantID, apiKey)

res, err := mg.KYC.Request(ctx, fintech.RequestCheckRequest{
    SubjectID:   "user_123",
    SubjectType: fintech.KycSubjectIndividual,
    CheckType:   fintech.KycCheckFull, // IDENTITY | AML | SANCTIONS | PEP | FULL
}, "kyc:user_123:onboard")
if res.RedirectURL != nil {
    redirectUser(*res.RedirectURL) // provider-hosted verification
}

check, err := mg.KYC.Get(ctx, res.Check.CheckID)
if check.Status == fintech.KycStatusPassed {
    // proceed
}

list, err := mg.KYC.List(ctx, "user_123", fintech.KycStatusPassed, 20, "")

// Operator-only manual override.
_, err = mg.KYC.Override(ctx, fintech.OverrideCheckRequest{
    CheckID: check.CheckID, Status: fintech.KycStatusOverridden,
    OverrideNote: "manual review cleared",
})
```

**TypeScript / Python:** use the REST endpoints directly for KYC until typed
KYC/Compliance/Merchant resources are added to those SDKs.

### Compliance alerts

Raise alerts (AML, sanctions, fraud, watchlist, …), triage by status and
severity, and run the `HasOpenAlerts` gate before sensitive operations such as
merchant withdrawals.

**Go**

```go
alert, err := mg.Compliance.Raise(ctx, fintech.RaiseAlertRequest{
    SubjectID: "user_123", SubjectType: "user",
    Category: fintech.AlertCategorySanctions,
    Severity: fintech.AlertSeverityHigh,
    Source:   "screening", Description: "OFAC name match",
}, "alert:user_123:ofac")

open, err := mg.Compliance.HasOpenAlerts(ctx, "user_123")
if open {
    // block withdrawal
}

_, err = mg.Compliance.Escalate(ctx, alert.AlertID, "compliance-team", "needs SAR review")
_, err = mg.Compliance.Resolve(ctx, alert.AlertID, "false positive — verified")
list, err := mg.Compliance.List(ctx, "user_123", fintech.AlertStatusOpen, "", 20, "")
```

**TypeScript / Python:** use the REST endpoints directly until typed resources
land in those SDKs.

### Merchant onboarding

A merchant moves through `PENDING → KYC_REQUIRED → UNDER_REVIEW → ACCEPTED`
(or `REJECTED` / `SUSPENDED` / `OFFBOARDED`). `Onboard` captures the profile and
config (accepted currencies, limits, fiat/crypto toggles); accept/reject/suspend
/reinstate drive the lifecycle.

**Go**

```go
m, err := mg.Merchant.Onboard(ctx, fintech.OnboardMerchantRequest{
    SubjectID:   "user_123",
    MerchantType: fintech.MerchantTypeBusiness,
    DisplayName: "Acme Store", LegalName: "Acme LLC",
    CountryCode: "UZ",
    Config: fintech.MerchantConfig{
        AcceptedCurrencies: []string{"UZS", "USD"},
        PrimaryCurrency:    "UZS",
        FiatEnabled:        true,
    },
}, "merchant:user_123:onboard")

_, err = mg.Merchant.Accept(ctx, m.MerchantID, "KYC passed")
// or Reject(id, reason) / Suspend(id, reason) / Reinstate(id, note)

list, err := mg.Merchant.List(ctx, fintech.MerchantStatusUnderReview, 20, "")
```

**TypeScript / Python:** use the REST endpoints directly until typed resources
land in those SDKs.

### Risk scoring & blocklist

`risk` is advisory: it returns a score (0–100), a risk level, a
`recommended_action` (`approve` | `review` | `decline`), and the triggered
rules. Your payments flow decides whether to block. The blocklist holds tenant
entries by email / phone / card BIN / IP / country.

**Go**

```go
a, err := mg.Risk.AssessTransaction(ctx, mashgate.AssessTransactionRequest{
    TenantID: tenantID, OrderID: "order_789",
    Amount: mashgate.Money{ /* amount + currency */ },
    CustomerEmail: "buyer@example.com", CustomerIP: "203.0.113.7",
})
switch a.RecommendedAction {
case "decline":
    // block (your decision)
}

_, err = mg.Risk.AddBlocklistEntry(ctx, mashgate.AddBlocklistEntryRequest{
    TenantID: tenantID, EntryType: "email", Value: "fraud@example.com",
})
profile, err := mg.Risk.GetRiskProfile(ctx, tenantID, "email", "buyer@example.com")
```

**TypeScript** (`risk` resource — endpoints differ slightly:
`assessPayment` / `assessRefund` / `investigatePayment`, plus rule CRUD):

```typescript
const result = await mg.risk.assessPayment({ /* RiskAssessmentRequest */ });
await mg.risk.addBlocklistEntry({ entry_type: "email", value: "fraud@example.com" });
const { entries } = await mg.risk.listBlocklist("email");
```

**Python**

```python
result = mg.risk.assess(amount="100.00", currency="UZS", customer_id="user_123")
mg.risk.add_blocklist_entry(entry_type="email", value="fraud@example.com")
mg.risk.list_blocklist()
```

### Guard: rate limits & IP blocklist

`guard` enforces per-tenant rate limits and IP blocklists at the edge.

**Go**

```go
res, err := mg.Guard.Check(ctx, mashgate.GuardCheckRequest{
    TenantID: tenantID, Path: "/v1/checkout", Method: "POST", IP: "203.0.113.7",
})
if !res.Allowed {
    // throttle / reject
}
_, err = mg.Guard.UpsertRateLimit(ctx, mashgate.UpsertRateLimitRequest{
    TenantID: tenantID, Path: "/v1/checkout", Method: "POST", RPM: 60,
})
_, err = mg.Guard.BlockIP(ctx, mashgate.BlockIPRequest{TenantID: tenantID, IP: "203.0.113.7"})
```

**TypeScript**

```typescript
const res = await mg.guard.check({ tenantId, path: "/v1/checkout", method: "POST", ip });
await mg.guard.upsertRateLimit({ tenantId, path: "/v1/checkout", method: "POST", rpm: 60 });
await mg.guard.blockIp({ tenantId, ip: "203.0.113.7" });
```

**Python:** not yet available for `guard` — use Go or TypeScript.

## Events

KYC and compliance state changes are published to HookLine. Subscribe via the
events/webhooks SDK and verify signatures — see
[Events & webhooks](./events-webhooks.md). Relevant topics include
`kyc.*` (check lifecycle), `compliance.*` (alert lifecycle), and `merchant.*`
(onboarding transitions). Confirm the exact set against your contract snapshot
([Service catalog](./service-catalog.md)) — both legacy dotted and envelope-v1
`<product>.<resource>.<verb>` forms may be emitted during migration.

## Best practices

- **18+ / identity gating is the consumer's responsibility.** Mashgate verifies
  and screens; your vertical decides what a given KYC status is allowed to do.
- **Screening is async.** Don't block the request thread waiting for a result —
  start the check, then react to `kyc.*` events or poll `Get` / `List`.
- **Risk is advisory.** Treat `recommended_action` as input to your own
  decision, not a hard gate, and keep the final say with your payments flow.
- **Gate sensitive actions on open alerts** (`HasOpenAlerts`) before merchant
  withdrawals or payouts.
- Make `Request` / `Raise` / `Onboard` idempotent with a domain-derived key, and
  scope every subject to `tenant_id`.

More: [Best practices](../best-practices.md).

## See also

- [Wallets & ledger](./wallets-ledger.md) — freeze / withdrawal gates.
- [Events & webhooks](./events-webhooks.md) — react to `kyc.*` / `compliance.*`.
- [Service catalog](./service-catalog.md) — KycService, ComplianceService, RiskService.
- [Building a vertical](../guides/building-a-vertical.md)
- [Data modeling & identity](../guides/data-modeling-and-identity.md)
