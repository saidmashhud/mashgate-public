# Flags & Observability

This guide covers runtime control and per-tenant visibility:

- **`flags`** — feature flags with percentage rollout, user/group targeting, and runtime evaluation (A/B testing, gradual rollouts, kill switches).
- **`analytics`** — read-only payment and customer analytics.
- **`logs`** — audit, activity, payment, and webhook log streams, plus custom event ingestion.
- **`metering`** — usage tracking and quota status that feed billing.
- **`developer`** — self-service API keys / applications, webhook-endpoint inspection, and integration health.

> SDK coverage differs by language for some resources — noted per operation below. `analytics` and `metering` are not yet available in the Python SDK; use Go or TypeScript.

## When to use

- **flags** — ship a feature to 10% of users, target a beta group, or flip a kill switch without redeploying.
- **analytics** — dashboards for payment volume, transaction counts, method/geo breakdowns, and customer cohorts.
- **logs** — surface an audit trail, debug webhook deliveries, or ingest your own product events.
- **metering** — show usage/quota in a billing or admin UI; record custom metered actions (e.g. AI tokens).
- **developer** — power a developer portal: manage keys/apps and monitor integration health.

## Feature flags

### Create, list, get

`rolloutPct` (0–100) plus `targetUsers` / `targetGroups` drive A/B and targeted rollouts.

**Go**

```go
flag, err := client.Flags.Create(ctx, mashgate.CreateFlagRequest{
    TenantID:     tenantID,
    FlagKey:      "new-checkout",
    Enabled:      true,
    RolloutPct:   10,
    TargetGroups: []string{"beta"},
    Description:  "New checkout flow",
})
flags, err := client.Flags.List(ctx, tenantID)
one, err := client.Flags.Get(ctx, "new-checkout", tenantID)
```

**TypeScript**

```ts
const flag = await client.flags.create({
  tenantId,
  flagKey: "new-checkout",
  enabled: true,
  rolloutPct: 10,
  targetGroups: ["beta"],
  description: "New checkout flow",
});
const flags = await client.flags.list(tenantId);
const one = await client.flags.get("new-checkout", tenantId);
```

**Python**

```python
flag = client.flags.create(
    tenant_id=tenant_id,
    key="new-checkout",
    enabled=True,
    rollout_percentage=10,
    description="New checkout flow",
)
flags = client.flags.list(tenant_id)
one = client.flags.get("new-checkout", tenant_id=tenant_id)
```

### Update

Go and TypeScript take pointer/optional fields so you patch only what you pass.

**Go**

```go
enabled, pct := true, 50
_, err := client.Flags.Update(ctx, "new-checkout", tenantID, mashgate.UpdateFlagRequest{
    Enabled:    &enabled,
    RolloutPct: &pct,
})
```

**TypeScript**

```ts
await client.flags.update("new-checkout", tenantId, { enabled: true, rolloutPct: 50 });
```

**Python**

```python
client.flags.update("new-checkout", tenant_id=tenant_id, enabled=True, rollout_percentage=50)
```

### Evaluate (runtime decision)

Evaluate a flag for a specific user/group. The result carries `enabled` and a `reason` (why it resolved that way — useful for debugging targeting).

**Go**

```go
ev, err := client.Flags.Evaluate(ctx, mashgate.EvaluateFlagRequest{
    TenantID: tenantID,
    FlagKey:  "new-checkout",
    UserID:   userID,
    Groups:   []string{"beta"},
})
if ev.Enabled { /* serve new flow */ }
```

**TypeScript**

```ts
const ev = await client.flags.evaluate({
  tenantId,
  flagKey: "new-checkout",
  userId,
  groups: ["beta"],
});
if (ev.enabled) { /* serve new flow */ }
```

**Python**

```python
ev = client.flags.evaluate(
    tenant_id=tenant_id,
    flag_key="new-checkout",
    entity_id=user_id,
    context={"group": "beta"},
)
```

## Analytics (read-only)

Period is one of `1d` / `7d` / `30d` / `90d` / `365d` / `custom`; granularity is `hour` / `day` / `week` / `month`. Not available in Python — use Go or TypeScript.

**Go**

```go
m, err := client.Analytics.GetPaymentMetrics(ctx, tenantID, "30d")
vol, err := client.Analytics.GetVolumeTimeSeries(ctx, tenantID, "30d", "day")
methods, err := client.Analytics.GetPaymentMethodBreakdown(ctx, tenantID, "30d")
geo, err := client.Analytics.GetGeoDistribution(ctx, tenantID, "30d")
fails, err := client.Analytics.GetFailureAnalysis(ctx, tenantID, "30d")

cm, err := client.Analytics.GetCustomerMetrics(ctx, tenantID, "90d")
cohorts, err := client.Analytics.GetCohortAnalysis(ctx, tenantID, "90d")
segments, err := client.Analytics.GetCustomerSegments(ctx, tenantID)
top, err := client.Analytics.GetTopCustomers(ctx, tenantID, "90d", 10)
```

**TypeScript**

```ts
const m = await client.analytics.getPaymentMetrics(tenantId, "30d");
const vol = await client.analytics.getVolumeTimeSeries(tenantId, "30d", "day");
const methods = await client.analytics.getPaymentMethodBreakdown(tenantId, "30d");
const geo = await client.analytics.getGeoDistribution(tenantId, "30d");
const fails = await client.analytics.getFailureAnalysis(tenantId, "30d");

const cm = await client.analytics.getCustomerMetrics(tenantId, "90d");
const cohorts = await client.analytics.getCohortAnalysis(tenantId, "90d");
const segments = await client.analytics.getCustomerSegments(tenantId);
const top = await client.analytics.getTopCustomers(tenantId, "90d", 10);
```

## Logs

Four read streams — `audit`, `activity`, `payments`, `webhooks` — share common pagination, plus custom event ingestion (`Track`, Go) into mgLogs.

**Go**

```go
params := mashgate.LogsQueryParams{TenantID: tenantID, Limit: 50}
audit, err := client.Logs.Audit(ctx, params, /*actor*/ "", /*action*/ "user.login")
acts, err := client.Logs.Activity(ctx, params, /*logType*/ "")
pays, err := client.Logs.Payments(ctx, params, /*status*/ "failed")
hooks, err := client.Logs.Webhooks(ctx, params, /*endpointID*/ "")
// audit.NextCursor for the next page

err = client.Logs.Track(ctx, mashgate.TrackEvent{
    TenantID: tenantID,
    Event:    "listing_viewed",
    Props:    map[string]any{"listingId": "4821"},
})
```

**TypeScript**

```ts
const audit = await client.logs.audit({ tenantId, action: "user.login", limit: 50 });
const acts = await client.logs.activity({ tenantId });
const pays = await client.logs.payments({ tenantId, status: "failed" });
const hooks = await client.logs.webhooks({ tenantId, endpointId });
```

**Python** (page-based; uses `from_ms`/`to_ms` epoch-millis windows)

```python
audit = client.logs.audit(tenant_id=tenant_id, action="user.login", page=1, page_size=50)
acts = client.logs.activity(tenant_id=tenant_id)
pays = client.logs.payments(tenant_id=tenant_id, status="failed")
hooks = client.logs.webhooks(tenant_id=tenant_id, endpoint_id=endpoint_id)
```

## Metering

The platform records billable events automatically on payment/storage/chain calls — you only call metering for custom meters (e.g. AI tokens) or to read usage. SDK surface differs by language; Python has no metering resource.

**Go** — record + read (`RecordUsage` is idempotent via the `Idempotency-Key` header, auto-generated if blank):

```go
rec, err := client.Metering.RecordUsage(ctx, mashgate.RecordUsageRequest{ /* TenantID, MeterCode, ... */ })

records, err := client.Metering.ListUsage(ctx, mashgate.ListUsageParams{
    TenantID: tenantID, MeterCode: "api_call", From: from, To: to,
})
summary, err := client.Metering.GetUsageSummary(ctx, tenantID, from, to)
```

**TypeScript** — usage summary, per-resource time series, and quota status:

```ts
const summary = await client.metering.getUsageSummary(tenantId);
const ts = await client.metering.getUsageTimeSeries(tenantId, "API_CALLS", start, end);
const quota = await client.metering.getQuotaStatus(tenantId);
if (quota.anyNearLimit) { /* warn the tenant */ }
```

## Developer

The developer surface differs by language. **Go** is API-key + integration-health focused; **TypeScript/Python** manage developer applications.

**Go** — keys, webhook-endpoint inspection, and health (secret material is returned only once, on create):

```go
keys, err := client.Developer.ListAPIKeys(ctx, tenantID)
created, err := client.Developer.CreateAPIKey(ctx, req) // created.Secret available once
err = client.Developer.RevokeAPIKey(ctx, keyID)

endpoints, err := client.Developer.ListWebhookEndpoints(ctx, tenantID)
activity, err := client.Developer.GetRecentActivity(ctx, tenantID, 20)
health, err := client.Developer.GetIntegrationHealth(ctx, tenantID)
```

**TypeScript**

```ts
const app = await client.developer.createApplication({ /* CreateApplicationRequest */ });
const { applications } = await client.developer.listApplications();
const one = await client.developer.getApplication(appId);
await client.developer.deleteApplication(appId);
```

**Python**

```python
app = client.developer.create_application(name="My App", app_type="web")
apps = client.developer.list_applications()
one = client.developer.get_application(app_id)
client.developer.delete_application(app_id)
```

## Best practices

- Use flags for gradual rollouts and kill switches; evaluate at request time and read the `reason` when targeting behaves unexpectedly.
- Scope every flags/analytics/logs/metering/developer call to the tenant. See [Best practices §4](../best-practices.md).
- Let the platform meter built-in actions automatically; only `RecordUsage` for custom meters, and make those calls idempotent. See [Best practices §2](../best-practices.md).
- Store the API-key secret server-side at creation — it is shown only once. See [Best practices §7](../best-practices.md).
- Page through logs with the returned cursor; don't request unbounded windows.

## See also

- [Notifications](./notifications.md) — delivery logs and mail events
- [Service catalog](./service-catalog.md) — `flags-service`, `analytics-service`, `logs-service`, `metering-service`
- [Building a vertical](../guides/building-a-vertical.md)
- [Best practices](../best-practices.md)
