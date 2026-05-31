# Recipe: Gate access behind 18+ / identity verification

**Goal:** before a user can enter the app (or unlock an adult-only / regulated
area), require an **18+ age + identity check**. The screening is **asynchronous**:
you request a check, the user lands in a `pending`/`in_review` state, and a
HookLine event flips them to `passed` or `failed` later.

Running example: a dating app gating sign-in behind 18+ verification. Same pattern
for any vertical that must verify identity before granting access.

> **Your responsibility, not Mashgate's.** Mashgate runs the KYC/identity
> *screening*; **deciding what to gate, and enforcing the gate, is your
> vertical's job.** A user is *not* allowed in until *your* code sees a `passed`
> result. Treat anything that isn't an explicit pass as "no access yet."

---

## What you'll use

| Concern | Mashgate module | SDK surface |
|---------|-----------------|-------------|
| Sign-up / login, the identity being verified | **Identity (mgID)** | `client.Login` / `client.Register` |
| Age / identity / sanctions / PEP screening | **KYC / Compliance** (Fintech Pack) | `fintech.Client.KYC.*` |
| Async result delivery (pending → passed/failed) | **Events / HookLine** | `client.ConstructEvent` + handler |

**What *your* vertical owns** (per [building a vertical](../guides/building-a-vertical.md)):
the **gate itself**. Mashgate owns the `KycCheck` record and its status; *you*
own a `gate_status` row per user (`unverified` / `pending` / `verified` /
`rejected`) keyed by the mgID `user_id`, and the middleware that blocks routes
until that row says `verified`. See
[data modeling & identity](../guides/data-modeling-and-identity.md): your
`gate_status.mg_user_id` is a `text` link to identity, not a foreign key into
Mashgate's DB.

---

## Flow

```
register ──► login ──► no verified gate? ──► KYC.Request(IDENTITY, idempotencyKey)
                                                     │
                                          response.RedirectURL?  ── yes ──► send user to provider-hosted flow
                                                     │
                                          your gate_status = "pending"  (BLOCK app access)
                                                     │
                                  ┌──────────────────┴───────────────────┐
                                  │   provider screens asynchronously     │
                                  └──────────────────┬───────────────────┘
                                                     │
                          HookLine ── kyc.check.updated ──► your /webhooks (verify + dedupe)
                                                     │
                       ┌─────────────────────────────┼─────────────────────────────┐
                       ▼                             ▼                              ▼
              status = PASSED               status = FAILED               status = IN_REVIEW
              gate_status="verified"        gate_status="rejected"        stay "pending"
              → UNLOCK app                  → keep blocked, show reason   → keep waiting
```

End-to-end: **register → login → request check → pending (blocked) → async event →
verified/rejected → unlock or stay blocked**. The user can be parked in `pending`
for seconds or hours — design the UI and the gate for that.

---

## Implementation

### 0. Init the clients

The KYC client is part of the **Fintech Pack** (`fintech.New`), separate from the
main client. The main client gives you auth + webhook verification.

```go
import (
    "github.com/saidmashhud/mashgate-public/sdk/go"          // package mashgate
    "github.com/saidmashhud/mashgate-public/sdk/go/fintech"
)

mg := mashgate.New(os.Getenv("MASHGATE_BASE_URL"), os.Getenv("MASHGATE_API_KEY"))
fin := fintech.New(os.Getenv("MASHGATE_BASE_URL"), tenantID, os.Getenv("MASHGATE_API_KEY"))
```

> **TypeScript / Python:** there is **no dedicated typed KYC resource** in the TS
> or Python SDKs today — the typed `KYCService` is **Go-only**. For the screening
> calls below, **use Go** (run them from your Go service / worker), or call the
> `/v1/kyc/checks` REST endpoints directly with an `Idempotency-Key` header. You
> can still verify and route the resulting `kyc.*` webhooks from any language.

Always take `userID` from the validated token, never from the request body
([best practices §5](../best-practices.md)).

### 1. Request the check (the gate starts closed)

On the first authenticated request from an unverified user, request an identity
check and set your local gate to `pending`. **Pass an idempotency key derived
from the user id** so a retry doesn't open a second check
([best practices §2](../best-practices.md)).

```go
// your DB: UPSERT gate_status (mg_user_id, tenant_id, status) VALUES (?, ?, 'unverified') ...
resp, err := fin.KYC.Request(ctx, fintech.RequestCheckRequest{
    SubjectID:   userID,                         // mgID user_id
    SubjectType: fintech.KycSubjectIndividual,
    CheckType:   fintech.KycCheckIdentity,       // identity + age document check
    Metadata:    map[string]string{"gate": "18plus", "min_age": "18"},
}, "kyc:"+userID)                                // idempotency key
if err != nil {
    return err
}

// Persist the check id and move the gate to "pending" — access stays BLOCKED.
setGateStatus(userID, "pending", resp.Check.CheckID)

// If the provider needs the user to complete a hosted flow, send them there.
if resp.RedirectURL != nil {
    redirect(*resp.RedirectURL)
}
```

`KYC.Request` returns the `KycCheck` (its `Status` typically starts at
`KYC_STATUS_PENDING` or `KYC_STATUS_IN_REVIEW`) and an optional `RedirectURL`.
**Surface the redirect to the end user** when present — that's the
provider-hosted identity capture.

> **Check types** available: `KycCheckIdentity`, `KycCheckAML`,
> `KycCheckSanctions`, `KycCheckPEP`, `KycCheckFull`. For an 18+ gate,
> `KycCheckIdentity` (document + age) is the usual choice; use `KycCheckFull` if
> you also need AML/sanctions/PEP in one shot.

### 2. Enforce the gate (everything blocked until `verified`)

This is the part Mashgate does **not** do for you. Your middleware reads *your*
`gate_status` row — not Mashgate — on every gated request:

```go
func require18Plus(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        userID := userIDFromToken(r) // from the validated JWT, never the body
        switch gateStatus(userID) {
        case "verified":
            next.ServeHTTP(w, r)                       // allow
        case "pending":
            http.Error(w, "verification pending", http.StatusForbidden)
        default: // "unverified" | "rejected"
            http.Error(w, "verification required", http.StatusForbidden)
        }
    })
}
```

Anything that is not an explicit `verified` is **no access**. Don't fail open.

### 3. Receive the async result (HookLine) — verify, dedupe, flip the gate

Screening finishes out of band. React to the `kyc.*` event: verify the signature,
dedupe on the event id, return 2xx fast ([best practices §3](../best-practices.md)).

```go
func handleKycWebhook(mg *mashgate.Client) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        body, _ := io.ReadAll(r.Body)
        event, err := mg.ConstructEvent(
            body,
            r.Header.Get("x-hl-signature"),
            r.Header.Get("x-hl-timestamp"),
            os.Getenv("MASHGATE_WEBHOOK_SECRET"),
        )
        if err != nil {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }
        if seenBefore(event.CanonicalID()) {          // idempotent: no-op
            w.WriteHeader(http.StatusOK)
            return
        }

        // kyc.* topics carry the check + subject + status in the payload.
        var p struct {
            CheckID   string `json:"check_id"`
            SubjectID string `json:"subject_id"` // = mg user_id
            Status    string `json:"status"`     // KYC_STATUS_*
        }
        _ = json.Unmarshal(event.PayloadBytes(), &p)

        switch fintech.KycStatus(p.Status) {
        case fintech.KycStatusPassed:
            setGateStatus(p.SubjectID, "verified", p.CheckID) // → UNLOCK
        case fintech.KycStatusFailed, fintech.KycStatusExpired:
            setGateStatus(p.SubjectID, "rejected", p.CheckID) // stay blocked
        case fintech.KycStatusInReview, fintech.KycStatusPending:
            // still screening — leave the gate "pending"
        }

        markSeen(event.CanonicalID())
        w.WriteHeader(http.StatusOK) // ack fast; HookLine retries non-2xx
    }
}
```

```ts
// Routing/verification works in any language even though the KYC *calls* are Go-only.
mg.webhooks.on("kyc.check.updated", async (event) => {
  const id = event.id ?? event.eventId;
  if (alreadyProcessed(id)) return;                  // dedupe
  const p = eventPayload(event) as { check_id: string; subject_id: string; status: string };
  if (p.status === "KYC_STATUS_PASSED")      await setGate(p.subject_id, "verified");
  else if (p.status === "KYC_STATUS_FAILED") await setGate(p.subject_id, "rejected");
  // IN_REVIEW / PENDING: leave the gate pending
});
```

> Match on `status` *values*, not topic strings — the same `kyc.*` topic can carry
> different statuses. Use `PayloadBytes()` / `eventPayload()` so you read the body
> whether it arrives in envelope-v1 (`payload`) or legacy (`data`) form.

### 4. Don't trust events for security-critical reads — re-fetch on the edge

Events can be delayed, retried, or (briefly) out of order. When a user hits a
gated route and your local status says `pending` but you want to be sure, you can
re-read the authoritative check:

```go
check, _ := fin.KYC.Get(ctx, checkID)
if check.Status == fintech.KycStatusPassed {
    setGateStatus(userID, "verified", check.CheckID)
}
```

You can also list a subject's checks (e.g. to show history or re-screen):

```go
res, _ := fin.KYC.List(ctx, userID, fintech.KycStatusPassed, 20, "")
```

### 5. Manual override (operator only)

A support/compliance operator can override a check (e.g. resolve a borderline
review). This is privileged — it writes an audit note and is **not** something a
regular user can trigger:

```go
_, _ = fin.KYC.Override(ctx, fintech.OverrideCheckRequest{
    CheckID:      checkID,
    Status:       fintech.KycStatusPassed,        // becomes KYC_STATUS_OVERRIDDEN-tracked
    OverrideNote: "manual review: passport verified by ops #1234",
})
```

---

## Gotchas / best practices

- **Gating is *your* responsibility.** Mashgate screens; your middleware enforces.
  Fail closed — anything not explicitly `passed` means no access.
- **Screening is async.** `KYC.Request` returns `pending`/`in_review`; the real
  outcome arrives later via `kyc.*` events. Design the UI for a waiting state.
- **Surface `RedirectURL`** when the response carries one — that's the
  provider-hosted identity capture; skipping it leaves the user stuck in pending.
- **Idempotency key on `KYC.Request`** — derive from the user id (`kyc:<userId>`),
  not random, so retries don't open duplicate checks. → [§2](../best-practices.md)
- **Idempotent webhook handlers** — dedupe on `CanonicalID()`/`event.id`; tolerate
  out-of-order delivery. → [§3](../best-practices.md)
- **Match on `KycStatus` values**, not topic names — one topic, many statuses.
- **Don't store identity attributes as SoT** — keep only your `gate_status`,
  keyed by `mg_user_id`; the `KycCheck` lives in Mashgate. → [§1](../best-practices.md)
- **Keep the webhook secret server-side**; it's returned once on
  `CreateWebhookEndpoint`. → [§7](../best-practices.md)

---

## See also

- [Building a vertical](../guides/building-a-vertical.md) — module/domain split.
- [Data modeling & identity](../guides/data-modeling-and-identity.md) — key your
  `gate_status` by `mg_user_id`.
- [Best practices](../best-practices.md) — idempotency, events, identity-from-token.
- [Service catalog](../modules/service-catalog.md) — KYC / Compliance (Fintech Pack).
- [Monetization recipe](./monetization-subscriptions-and-boosts.md) — gate paid
  features behind a verified user.
