# Data modeling & identity

The most common question when starting a vertical: *"I want to extend the users
table — how?"* You don't extend Mashgate's users table. You model your own data
**next to** it and join by id. This guide shows the pattern.

---

## Who owns identity

**Mashgate (mgID) owns identity.** Users, credentials, sessions, roles, email,
phone — these live in Mashgate, are multi-tenant, and are the **source of truth**.
You do not own that table, cannot add columns to it, and should not duplicate its
authoritative fields.

Your vertical owns **everything domain-specific** about a user: a dating profile,
a seller's storefront settings, preferences, etc. You keep that in **your own
database**, in a row keyed to the Mashgate user id.

```
mgID:  user_id ─────────────┐  (identity SoT: phone, email, roles)
                            │
your DB: profiles.mg_user_id┘  (your data: bio, photos, preferences, geo)
```

## The Mashgate user id

`user_id` is an **opaque string** (the API types it as `string`; a base64url form
is also exposed as `user_id_b64`). It is **not** a database auto-increment integer,
and you should **not** assume it is a UUID or parse its structure. Treat it as an
opaque token issued by mgID.

You obtain it from the **authenticated request context** — the validated JWT `sub`
is the user id. Don't trust a user id sent in a request body; take it from the token
that the gateway validated.

## Pattern: a profile table keyed to the Mashgate id

You have two good options.

### Option A — `mg_user_id` as the primary key

Cleanest conceptually: the local row *is* the global identity.

```sql
CREATE TABLE profiles (
  mg_user_id   text PRIMARY KEY,        -- = mgID user_id
  tenant_id    text NOT NULL,
  display_name text,
  bio          text,
  photos       jsonb,
  created_at   timestamptz NOT NULL DEFAULT now(),
  updated_at   timestamptz NOT NULL DEFAULT now()
);
```

### Option B — local surrogate key + unique `mg_user_id` (recommended when you have many FKs)

A stable local `bigint` primary key, with `mg_user_id` as a unique anchor:

```sql
CREATE TABLE profiles (
  id          bigserial PRIMARY KEY,    -- internal FKs point here (short, fast, stable)
  mg_user_id  text NOT NULL,
  tenant_id   text NOT NULL,
  display_name text,
  bio         text,
  photos      jsonb,
  created_at  timestamptz NOT NULL DEFAULT now(),
  updated_at  timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, mg_user_id)
);
```

**Choosing:** if your schema has many tables referencing the profile
(`swipes`, `matches`, `orders`, …), prefer **Option B** — internal foreign keys on
a `bigint` are smaller, faster, and decoupled from the external id format, while
`mg_user_id` stays a unique link to identity. If references are few, **Option A**
is fine and simpler.

## Rules

1. **Column type is `text`/`varchar`** — never `int`, never assume `uuid`.
2. **Don't store identity attributes as source of truth.** Email / phone / name /
   roles belong to mgID. Keep only domain data locally. If you cache a display name
   for speed, treat it as a cache and refresh it on a `user.updated` event.
3. **No foreign keys into Mashgate's database.** That's a service boundary. Enforce
   the link to `mg_user_id` in application code (identity is proven by the validated
   token), not by a cross-database FK.
4. **Scope by `tenant_id`.** The same person can exist under different tenants;
   uniqueness is `(tenant_id, mg_user_id)`, not `mg_user_id` alone.
5. **Provision the row, don't insert into a foreign table.** Create the profile:
   - **lazily** — on first authenticated request, `UPSERT … ON CONFLICT (mg_user_id)`; or
   - **reactively** — subscribe to the `user.created` webhook (HookLine) and create it.

## Joining back at request time

```ts
// gateway has authenticated the caller; take the id from the verified context
const userId = ctx.user.id;                       // mgID user_id
const profile = await db.profiles.upsertByMgUserId(userId, ctx.tenantId);
// ... your domain logic keyed by profile.id (Option B) or profile.mg_user_id (Option A)
```

For money, the same idea applies one level up: the **ledger** is the source of
truth for balances — don't keep your own authoritative balance column, read/derive
from Mashgate and reconcile via events. See [Best practices](../best-practices.md).
