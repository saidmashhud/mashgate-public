# Wallets & Ledger

Multi-currency wallets backed by an authoritative double-entry ledger. The
ledger is the source of truth for balances and money movements — your vertical
reads from it and reconciles via events rather than holding its own balance.
Covers the SDK resources `wallet` (end-user), `wallet_admin` (tenant-scoped
WalletService), the fintech-pack `fintech.Wallet` service, and `chain`
(on-chain crypto rails).

## When to use

- A user or merchant needs a balance you can credit, debit, hold, or settle.
- You move funds between accounts inside a tenant (escrow release, payout,
  marketplace settlement) and need it to be atomic and auditable.
- You accept crypto deposits / send crypto withdrawals on Solana, TRON, or an
  EVM chain, or generate per-user deposit addresses.
- A support operator needs to freeze, adjust, or audit a wallet.

Don't use it as a generic key-value store for app state, and don't keep a
parallel "authoritative" balance column in your own database — see
[Best practices](#best-practices).

## Key operations

### End-user wallet (balance + saved cards)

Top-level client methods for the authenticated end user — balance is derived
from the ledger (available vs. holds).

**Go**

```go
bal, err := mg.GetWalletBalance(ctx, "UZS") // ISO 4217; "" = default currency
// bal.AvailableBalance / bal.HoldsBalance are decimal strings

methods, err := mg.ListSavedPaymentMethods(ctx)
err = mg.SetDefaultPaymentMethod(ctx, methods[0].PaymentMethodID)
err = mg.RemoveSavedPaymentMethod(ctx, paymentMethodID)
```

**TypeScript**

```typescript
const bal = await mg.wallet.getBalance("UZS");
const { paymentMethods } = await mg.wallet.listPaymentMethods();
await mg.wallet.setDefaultPaymentMethod(paymentMethods[0].paymentMethodId);
await mg.wallet.removePaymentMethod(paymentMethodId);
```

**Python**

```python
bal = mg.wallet.get_balance()
methods = mg.wallet.list_payment_methods()
mg.wallet.set_default_payment_method(payment_method_id)
mg.wallet.remove_payment_method(payment_method_id)
```

### Tenant wallets, credit / debit, transfers

The fintech-pack `WalletService` (Go: `fintech`, TS: `walletAdmin`, Python:
`wallet_admin`) is the full WalletService surface: create wallets, move money,
and transfer between accounts. Money fields are **decimal strings** in minor or
major units (never floats) to avoid precision loss. Mutating calls take an
**idempotency key** — pass a stable key derived from your domain id.

**Go** (fintech pack — construct with `fintech.New(baseURL, tenantID, apiKey)`)

```go
fc := fintech.New(baseURL, tenantID, apiKey)

w, err := fc.Wallet.Create(ctx, fintech.CreateWalletRequest{
    SubjectID:   "user_123",
    SubjectType: "user",
    WalletType:  fintech.WalletTypeFiat,
    Currency:    fintech.CurrencyUZS,
}, "wallet:user_123:create")

tx, err := fc.Wallet.Credit(ctx, fintech.CreditRequest{
    WalletID: w.WalletID,
    Amount:   "100.50",
    Reason:   fintech.ReasonDeposit,
}, "credit:order_789")

// Atomic same-tenant, same-currency transfer. Server commits one Postgres
// tx: both balances + two wallet_transactions rows + three outbox events.
res, err := fc.Wallet.Transfer(ctx, fintech.TransferRequest{
    FromWalletID: w.WalletID,
    ToWalletID:   "wal_merchant",
    Amount:       "25.00",
    Reason:       fintech.ReasonSettlement,
}, "transfer:order_789")
// res.TransferID, res.Debit, res.Credit
```

**TypeScript** (`walletAdmin` resource)

```typescript
const w = await mg.walletAdmin.create({
  subject_id: "user_123",
  subject_type: "user",
  wallet_type: WalletType.Fiat,
  currency: Currency.UZS,
  idempotency_key: "wallet:user_123:create",
});

const tx = await mg.walletAdmin.credit(w.wallet_id, {
  amount: "100.50",
  reason: TransactionReason.Deposit,
  idempotency_key: "credit:order_789",
});

const res = await mg.walletAdmin.transfer(w.wallet_id, {
  to_wallet_id: "wal_merchant",
  amount: "25.00",
  reason: TransactionReason.Settlement,
  idempotency_key: "transfer:order_789",
});
```

**Python** (`wallet_admin` resource)

```python
w = mg.wallet_admin.create(
    subject_id="user_123", subject_type="user",
    wallet_type=WalletType.FIAT, currency=Currency.UZS,
    idempotency_key="wallet:user_123:create",
)
tx = mg.wallet_admin.credit(
    w["wallet_id"], amount="100.50",
    reason=TransactionReason.DEPOSIT, idempotency_key="credit:order_789",
)
res = mg.wallet_admin.transfer(
    w["wallet_id"], to_wallet_id="wal_merchant", amount="25.00",
    reason=TransactionReason.SETTLEMENT, idempotency_key="transfer:order_789",
)
```

`Debit` mirrors `Credit`. List with `Wallet.List(...)`,
`Wallet.ListTransactions(...)`, and fetch one with
`Wallet.GetTransaction(walletID, txID)`.

### On-chain crypto wallets and deposit/withdraw

> **Crypto is testnet-only.** On-chain support targets Solana, TRON, and EVM
> chains (Ethereum, Base, Polygon, BSC). Treat balances as testnet, and never
> ship mnemonics to your backend store.

`CreateChain` generates a **non-custodial** wallet — the BIP-39 mnemonic is
returned **once**. Surface it to the end user and never persist it; the server
keeps only a SHA-256 hash for duplicate detection. `ImportChain` recovers from a
user-supplied phrase (idempotent per subject; cross-subject re-import is
`PERMISSION_DENIED`). Currently SOLANA only for create/import.

**Go**

```go
res, err := fc.Wallet.CreateChain(ctx, fintech.CreateChainWalletRequest{
    SubjectID: "user_123", SubjectType: "user",
    Currency: fintech.CurrencySOL, Network: fintech.NetworkSolana,
}, "chain:user_123:create")
showOnce(res.Mnemonic) // display to user, then drop from memory

// Deposit address: pass a non-empty mint for SPL tokens (returns the ATA),
// empty mint for the native asset.
addr, err := fc.Wallet.DepositAddress(ctx, res.Wallet.WalletID,
    fintech.NetworkSolana, fintech.MintUSDCSolanaMainnet)

// Withdraw. mint selects SPL token vs native; sponsor_wallet_id enables
// gasless withdrawals (sponsor pays chain fee + ATA rent).
tx, err := fc.Wallet.Withdraw(ctx, fintech.WithdrawRequest{
    WalletID: res.Wallet.WalletID, Amount: "10.00",
    DestinationType: "crypto_address", DestinationID: "9xQe...",
    Network: fintech.NetworkSolana, Mint: fintech.MintUSDCSolanaMainnet,
}, "withdraw:payout_42")
```

**TypeScript**

```typescript
const res = await mg.walletAdmin.createChain({
  subject_id: "user_123", subject_type: "user",
  currency: Currency.SOL, network: Network.Solana,
  idempotency_key: "chain:user_123:create",
});
showOnce(res.mnemonic);

const addr = await mg.walletAdmin.depositAddress(
  res.wallet.wallet_id, Network.Solana, Mint.USDCSolanaMainnet);

const tx = await mg.walletAdmin.withdraw(res.wallet.wallet_id, {
  amount: "10.00",
  destination_type: "crypto_address", destination_id: "9xQe...",
  network: Network.Solana, mint: Mint.USDCSolanaMainnet,
  idempotency_key: "withdraw:payout_42",
});
```

**Python**

```python
res = mg.wallet_admin.create_chain(
    subject_id="user_123", subject_type="user",
    currency=Currency.SOL, network=Network.SOLANA,
    idempotency_key="chain:user_123:create",
)
show_once(res["mnemonic"])
addr = mg.wallet_admin.deposit_address(
    res["wallet"]["wallet_id"], network=Network.SOLANA, mint=Mint.USDC_SOLANA_MAINNET)
tx = mg.wallet_admin.withdraw(
    res["wallet"]["wallet_id"], amount="10.00",
    destination_type="crypto_address", destination_id="9xQe...",
    network=Network.SOLANA, mint=Mint.USDC_SOLANA_MAINNET,
    idempotency_key="withdraw:payout_42",
)
```

### Low-level chain RPC (addresses, transactions, networks)

The `chain` resource wraps ChainService for raw chain reads/sends (deposit
address derivation, tx history, fee estimates, block/network metadata).

**Go**

```go
addr, err := mg.Chain.CreateAddress(ctx, mashgate.CreateAddressRequest{
    TenantID: tenantID, Network: "solana_mainnet", Label: "deposit",
})
bal, err := mg.Chain.GetBalance(ctx, addr.AddressID)
fee, err := mg.Chain.EstimateFee(ctx, mashgate.EstimateFeeRequest{
    TenantID: tenantID, Network: "solana_mainnet",
    From: addr.Address, To: "9xQe...", Amount: "1.0", Asset: "SOL",
})
tx, err := mg.Chain.SendTransaction(ctx, mashgate.SendTransactionRequest{
    TenantID: tenantID, Network: "solana_mainnet",
    FromAddressID: addr.AddressID, To: "9xQe...", Amount: "1.0", Asset: "SOL",
    IdempotencyKey: "send:payout_42",
})
nets, err := mg.Chain.ListNetworks(ctx)
```

**TypeScript** — the `ChainResource` exposes a higher-level surface
(`createWallet`, `pay`, `swap`, `createEscrow`, `onRamp`/`offRamp`,
`screenAddress`, `gasEstimate`, `exchangeRate`, `batchPayout`):

```typescript
const wallet = await mg.chain.createWallet({
  tenantId, userId: "user_123", walletType: "custodial",
  networks: ["SOLANA", "ETHEREUM"],
});
const { balances } = await mg.chain.getWalletBalance(wallet.walletId);
const payment = await mg.chain.pay({ amount: "10.0", asset: "USDC", network: "SOLANA" });
const screen = await mg.chain.screenAddress("9xQe...", "SOLANA");
```

**Python:** not yet available for the low-level chain / on-chain wallet APIs —
use Go or TypeScript.

### Admin: freeze, adjust, audit

The `wallet_admin` privileged RPCs (Go: `WalletAdminClient`, requires
`wallet:admin:*` permissions) cover support operations. Every action writes to
the platform audit log with operator id, reason, and before/after state.

**Go**

```go
admin := mg.WalletAdmin // *WalletAdminClient

wallets, err := admin.AdminListWallets(ctx, tenantID, "frozen", 1, 50)
w, err := admin.AdminGetWallet(ctx, walletID)

_, err = admin.AdminFreezeWallet(ctx, walletID, mashgate.FreezeWalletRequest{
    Reason: "fraud_suspicion",
})
_, err = admin.AdminUnfreezeWallet(ctx, walletID, mashgate.UnfreezeWalletRequest{
    ResolvedReason: "cleared_by_review",
})

adj, err := admin.AdminAdjustBalance(ctx, walletID, mashgate.AdjustBalanceRequest{
    AmountCents: -500, Reason: "chargeback correction",
    IdempotencyKey: "adjust:case_17",
}) // negative = debit

err = admin.AdminCloseWallet(ctx, walletID, "account closed", false)
entries, err := admin.AdminAuditLog(ctx, walletID, 1, 50)
```

**TypeScript / Python:** freeze, credit/debit, transfer, and listings are
available via the `walletAdmin` / `wallet_admin` resources shown above. The
privileged adjust/close/audit RPCs are currently Go-only — use the Go SDK.

## Events

Money movements are published to HookLine. Subscribe with the events/webhooks
SDK and verify signatures (see [Events & webhooks](./events-webhooks.md)).

- `wallet.credit` — funds credited to a wallet.
- `wallet.debit` — funds debited from a wallet.
- `wallet.transfer` — an atomic inter-wallet transfer (emitted alongside the
  paired `wallet.debit` + `wallet.credit`). Correlate via `transfer_id`.

A `Transfer` commits the balance change and emits all three events in the same
transaction, so a consumer that sees `wallet.transfer` can rely on the money
state already being final. Payment, refund, and checkout lifecycle events
(`payments.payment.*`, `payments.refund.*`, `payments.checkout_session.*`) are
emitted by the payments services and often drive the credits above.

## Best practices

- **The ledger is the money source of truth.** Read balances from Mashgate and
  reconcile via `wallet.*` events; treat any local balance column as a cache,
  never as authoritative.
- **Make every money operation idempotent.** Pass a stable idempotency key
  derived from your domain id (`credit:order_789`), not from `now()`/random, and
  reuse it on retries.
- **Use decimal strings, never floats**, for all amounts.
- **On-chain is testnet-only.** Never persist mnemonics — surface them once and
  drop them from memory; the server stores only a hash.
- Scope everything to `tenant_id`; transfers are same-tenant, same-currency.

More: [Best practices](../best-practices.md).

## See also

- [Events & webhooks](./events-webhooks.md) — react to `wallet.*` events.
- [KYC & compliance](./kyc-compliance.md) — freeze/screen flows feeding wallet holds.
- [Service catalog](./service-catalog.md) — WalletService, mgChain services.
- [Data modeling & identity](../guides/data-modeling-and-identity.md)
- [Building a vertical](../guides/building-a-vertical.md)
