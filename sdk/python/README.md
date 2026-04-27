# mashgate (Python SDK)

Official Python SDK for the Mashgate Payment Gateway API. Mirrors the
gRPC contracts in
[mashgate/contracts/proto/v1/](https://github.com/saidmashhud/mashgate/tree/main/contracts/proto/v1)
and reaches the gateway over REST (gRPC-JSON transcoding via Envoy).

**Repo:** <https://github.com/saidmashhud/mashgate-public>

```bash
pip install mashgate
```

Requires Python 3.9+.

## Quick start

```python
from mashgate import MashgateClient

mg = MashgateClient(base_url="https://api.mashgate.uz", api_key="mg_test_key")
payment = mg.payments.create(amount="100.00", currency="UZS")
```

## Wallet APIs

Two distinct wallet surfaces:

- **`mg.wallet`** — end-user view (saved payment methods, balance,
  movements). Use with a customer JWT.
- **`mg.wallet_admin`** — admin / merchant view (full
  `wallet.v1.WalletService`). Use with an admin JWT or service account
  API key.

### `wallet_admin` — full WalletService

```python
from mashgate import (
    MashgateClient,
    Currency, Network, Mint,
    WalletType, WalletStatus, TransactionReason,
)

mg = MashgateClient(base_url="https://api.mashgate.uz", api_key="mg_admin_key")

# Off-chain wallet
w = mg.wallet_admin.create(
    subject_id="user-123",
    subject_type="user",
    wallet_type=WalletType.FIAT,
    currency=Currency.UZS,
    idempotency_key="idem-create-1",
)

# On-chain wallet — mnemonic returned ONCE; surface to user, never persist.
chain = mg.wallet_admin.create_chain(
    subject_id="user-123",
    subject_type="user",
    currency=Currency.USDC,
    network=Network.SOLANA,
    idempotency_key="idem-chain-1",
)
show_once_to_end_user(chain["mnemonic"])

# Deposit address
#   - SPL: pass `mint` → returns the Associated Token Account.
#   - Native asset: leave `mint=None` → returns the wallet owner address.
ata = mg.wallet_admin.deposit_address(
    chain["wallet"]["wallet_id"],
    network=Network.SOLANA,
    mint=Mint.USDC_SOLANA_MAINNET,
)
sol = mg.wallet_admin.deposit_address(
    chain["wallet"]["wallet_id"],
    network=Network.SOLANA,
)

# Withdraw — pass `mint` for SPL, leave None for native SOL.
tx = mg.wallet_admin.withdraw(
    chain["wallet"]["wallet_id"],
    amount="10.50",
    destination_type="crypto_address",
    destination_id="RecipientSolanaAddr",
    network=Network.SOLANA,
    mint=Mint.USDC_SOLANA_MAINNET,
    idempotency_key="idem-w-1",
)

# Compliance
mg.wallet_admin.freeze(w["wallet_id"], reason="fraud-investigation")
mg.wallet_admin.unfreeze(w["wallet_id"], note="case-resolved")

# Pagination — opaque cursor; empty cursor = first page.
resp = mg.wallet_admin.list(limit=50)
while resp.get("next_cursor"):
    resp = mg.wallet_admin.list(limit=50, cursor=resp["next_cursor"])

# Money movement on an existing wallet
mg.wallet_admin.credit(
    w["wallet_id"],
    amount="100.00",
    reason=TransactionReason.DEPOSIT,
    idempotency_key="idem-credit-1",
)
```

### Typed constants

`Currency`, `Network`, `Mint`, plus enum-strings from `wallet.proto`
(`WalletStatus`, `WalletType`, `TransactionType`, `TransactionStatus`,
`TransactionReason`) are `(str, Enum)` — they serialise transparently
to JSON, compare equal to their string values, and accept either the
enum member or a plain string at call sites:

```python
Currency.USDC.value           # "USDC"
Currency.USDC == "USDC"       # True
Network.SOLANA                # "SOLANA"
Mint.USDC_SOLANA_MAINNET      # "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"
```

`Mint` accepts arbitrary base58 strings outside the listed mainnet
whitelist — server-side (`chain-rpc.derive_spl_ata`) is the
authoritative validator.

## Errors

Non-2xx responses raise `MashgateError`:

```python
from mashgate import MashgateError

try:
    mg.wallet_admin.get("missing")
except MashgateError as e:
    print(e.status, e.code, e)
```

## Webhooks

```python
from mashgate import verify_webhook_signature

@app.post("/webhooks/mashgate")
async def webhook(request):
    body = await request.body()
    sig = request.headers["X-Mashgate-Signature"]
    if not verify_webhook_signature(body=body, signature=sig, secret=SECRET):
        return Response(status_code=401)
    # handle event
```

## Development

```bash
python -m venv .venv && source .venv/bin/activate
pip install -e '.[dev]'
PYTHONPATH=. pytest
```
