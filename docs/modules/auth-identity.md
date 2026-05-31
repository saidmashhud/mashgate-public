# Auth & Identity

Mashgate's Identity layer (mgID) owns users, sessions, and access control for your
tenant. This guide covers the SDK resources for authentication (`auth`), API keys
and permission checks (`iam`), and per-tenant configuration (`settings`). Identity
is the source of truth — key your own tables by the `user_id` Mashgate hands you,
don't mirror credentials.

## When to use

- You need sign-up, login, and session management (JWT access + refresh tokens).
- You authenticate users by **phone + OTP** (the primary flow for Central-Asia
  verticals) or by email/password.
- You need to issue machine-to-machine **API keys** for backend services.
- You enforce per-request permission checks (`can this principal do X?`).
- You store lightweight per-tenant configuration (branding, notification prefs,
  simple feature toggles).

## Key operations

### Phone-OTP login (primary flow)

The recommended authentication flow for verticals is passwordless phone-OTP:
`SendOtp(phone, purpose="login")` delivers a code, `VerifyOtp` validates it. Provide
a `phone` for registration / passwordless flows, or a `user_id` for an existing user.

**Go**

```go
// 1. Send the code.
if err := client.SendOtp(ctx, mashgate.SendOtpRequest{
    Phone:   "+992900000000",
    Purpose: "login", // "login" | "password_reset" | "phone_verify"
}); err != nil {
    return err
}

// 2. Verify the code the user entered.
ok, err := client.VerifyOtp(ctx, mashgate.VerifyOtpRequest{
    Phone:   "+992900000000",
    Code:    "123456",
    Purpose: "login",
})
if err != nil || !ok {
    return fmt.Errorf("invalid otp")
}
```

> Note: real SMS delivery depends on the platform's SMS provider being provisioned.
> See the [service catalog](./service-catalog.md) for current channel status.

TypeScript: the published TS `auth` resource exposes email/password login and
profile management; phone-OTP send/verify is currently Go-only.

### Email + password login

`Login` returns a `TokenPair` (access + refresh). The TS client also stores the
access token on the client instance for subsequent calls.

**Go**

```go
pair, err := client.Login(ctx, "merchant@example.com", "secret")
if err != nil {
    return err
}
// pair.AccessToken is always set; pair.User may be nil if upstream omits it.
exp := pair.ExpiresAtUnix()
```

**TypeScript**

```ts
const result = await client.auth.login({
  email: "merchant@example.com",
  password: "secret",
});
// login() also calls client.setAccessToken(result.accessToken) for you.
```

**Python**

```python
data = client.auth.login(email="merchant@example.com", password="secret")
# stores accessToken on the client automatically
```

### Register a user

`Register` creates the user record only — no tokens. Chain `Login` (or a phone-OTP
flow) afterwards to obtain a session.

> Security: default `Role` to `"merchant"` for self-service merchant flows. The
> upstream register endpoint does not yet whitelist roles, so never pass a
> caller-supplied role straight through — issue admin-grade roles via IAM instead.

**Go**

```go
user, err := client.Register(ctx, mashgate.RegisterRequest{
    Email:    "merchant@example.com",
    Password: "secret",
    FullName: "Merchant",
    TenantID: tenantID,
    Role:     "merchant",
})
if err == nil {
    pair, _ := client.Login(ctx, user.Email, "secret")
    _ = pair
}
```

**TypeScript**

```ts
const user = await client.auth.register({
  email: "merchant@example.com",
  password: "secret",
  tenantId,
  role: "merchant",
});
```

**Python**

```python
user = client.auth.register(
    email="merchant@example.com",
    password="secret",
    tenant_id=tenant_id,
    role="merchant",
)
```

### Refresh & logout

**Go**

```go
pair, err := client.RefreshToken(ctx, oldRefreshToken)
// ...
err = client.Logout(ctx, refreshToken) // invalidates the refresh token server-side
```

**TypeScript**

```ts
await client.auth.refreshToken(oldRefreshToken);
await client.auth.logout();
```

**Python**

```python
client.auth.refresh_token(old_refresh_token)
client.auth.logout()
```

### Password & profile lifecycle (Go)

The Go client exposes the full credential lifecycle. Password reset and phone/email
changes are gated by an OTP you obtain first via `SendOtp`.

```go
// Authenticated change (user proves the current password).
err := client.ChangePassword(ctx, mashgate.ChangePasswordRequest{
    UserID:          userID,
    CurrentPassword: "old",
    NewPassword:     "new",
})

// Forgot-password completion: SendOtp(purpose="password_reset") first, then:
err = client.ResetPassword(ctx, mashgate.ResetPasswordRequest{
    Phone:       "+992900000000",
    Code:        "123456",
    NewPassword: "new",
})

// Change phone: SendOtp(phone=NEW, purpose="phone_verify") first, then:
err = client.UpdateUserPhone(ctx, mashgate.UpdateUserPhoneRequest{
    UserID:   userID,
    NewPhone: "+992900000001",
    Code:     "123456",
})
```

`DeleteAccount` soft-deletes the user (downstream products consume the resulting
account-deletion event).

### API keys & permission checks (`iam`)

Issue API keys for backend-to-backend auth, and check permissions at request time.
`CreateAPIKey` returns the secret **once** — store it securely.

**Go**

```go
created, err := client.CreateAPIKey(ctx, mashgate.CreateAPIKeyRequest{
    Name:        "zist-backend",
    Permissions: []string{"payments:write", "checkout:write"},
})
// created.Secret is shown only here.

keys, err := client.ListAPIKeys(ctx)
err = client.DeleteAPIKey(ctx, keyID)

// Gate an action on a permission (use as middleware: if !ok { return 403 }).
ok, err := client.CheckPermission(ctx, "payments:refund")
```

**TypeScript**

The TS `iam` resource is richer — it covers roles, groups, policies, tenants, and
API keys. Examples:

```ts
const { apiKey, plaintextKey } = await client.iam.createApiKey({
  tenantId,
  clientId,
  name: "zist-backend",
  scopes: ["payments:write"],
});
// plaintextKey is returned once.

// RBAC: assign / check.
await client.iam.assignRole(tenantId, principalId, roleId);
const { allow, reason } = await client.iam.evaluateAccess({
  tenantId,
  principalId,
  permission: "payments:refund",
});

const { tenants } = await client.iam.listTenants({ status: "active" });
```

Python: not yet available for this module — use Go or TypeScript.

### Tenant directory (`iam`, Go)

For cold-start backfill before subscribing to tenant events, list tenants visible
to the authenticated principal:

```go
tenants, err := client.ListTenants(ctx, &mashgate.ListTenantsOptions{
    Status:   "active",
    PageSize: 100,
})
```

### Per-tenant settings (`settings`)

The Go `settings` client stores arbitrary key-value pairs scoped to a tenant
(branding, notification preferences, lightweight flags). Filter by namespace prefix.

**Go**

```go
all, err := client.Settings.GetSettings(ctx, tenantID, "notify.")
val, found, err := client.Settings.GetSetting(ctx, tenantID, "branding.logo_url")

err = client.Settings.UpdateSetting(ctx, tenantID, "branding.logo_url", "https://...")
err = client.Settings.UpdateSettings(ctx, tenantID, map[string]string{
    "notify.email_enabled": "true",
    "notify.sms_enabled":   "false",
})
err = client.Settings.DeleteSetting(ctx, tenantID, "branding.logo_url")
```

**TypeScript**

The TS `settings` resource is a merchant-settings blob (refund policy, etc.):

```ts
const settings = await client.settings.get();
await client.settings.update({ refundEnabled: true, maxRefundAmount: "500000" });
```

**Python**

```python
settings = client.settings.get()
client.settings.update(refund_enabled=True, max_refund_amount="500000")
```

## Best practices

- **Take identity from the validated token, not from a request body.** The
  authenticated `user_id` / `tenant_id` come from the JWT the gateway validates.
- **Default registration to `role="merchant"`;** never forward a caller-supplied
  admin role — grant elevated roles via IAM.
- **Treat API key secrets as tenant credentials:** inject from a secret store,
  never commit, never ship to the browser.

See [Best practices](../best-practices.md) (§1 own-what-it-owns, §5 token identity,
§7 secrets) for the full rules.

## See also

- [Building a vertical](../guides/building-a-vertical.md)
- [Data modeling & identity](../guides/data-modeling-and-identity.md)
- [Best practices](../best-practices.md)
- [Service catalog](./service-catalog.md)
