# @mashgate/sdk

Official TypeScript SDK for the Mashgate BaaS platform. Mirrors
the gRPC contracts in
[mashgate/contracts/proto/v1/](https://github.com/saidmashhud/mashgate/tree/main/contracts/proto/v1)
and reaches the gateway over REST (gRPC-JSON transcoding via Envoy).

**Repo:** <https://github.com/saidmashhud/mashgate-public>

```bash
npm install @mashgate/sdk
```

## Quick start

```ts
import { MashgateClient } from "@mashgate/sdk";

const mg = new MashgateClient({
  baseUrl: "https://api.mashgate.uz",
  apiKey: "mg_test_key",
});

const payment = await mg.payments.create({
  amount: "100.00",
  currency: "UZS",
  description: "Subscription",
});
```

## Fintech resources

KYC, compliance, and merchant acceptance require a tenant-scoped client. The
SDK sends `X-Tenant-ID` and includes the tenant in the canonical request body.

```ts
import { KycCheckType, KycSubjectType } from "@mashgate/sdk";

const mg = new MashgateClient({
  baseUrl: "https://api.mashgate.uz",
  apiKey: process.env.MASHGATE_API_KEY!,
  tenantId: process.env.MASHGATE_TENANT_ID!,
});

const check = await mg.kyc.request({
  subjectId: "user-123",
  subjectType: KycSubjectType.Individual,
  checkType: KycCheckType.Full,
}, crypto.randomUUID());
```

The tenant-scoped namespaces are `mg.kyc`, `mg.compliance`,
`mg.merchant`, and `mg.walletAdmin`. Pass a stable, unique idempotency key to
KYC requests, compliance-alert creation, and merchant onboarding.

## Wallet APIs

The SDK exposes two distinct wallet surfaces:

- **`mg.wallet`** — *end-user view* (saved payment methods, balance,
  movements). Use with a customer-issued JWT.
- **`mg.walletAdmin`** — *admin / merchant view* (full
  `wallet.v1.WalletService`). Use with an admin JWT or service account
  API key.

### `walletAdmin` — full WalletService

```ts
import {
  MashgateClient,
  Currency,
  Network,
  Mint,
  WalletType,
} from "@mashgate/sdk";

const mg = new MashgateClient({
  baseUrl: "https://api.mashgate.uz",
  apiKey: process.env.MASHGATE_API_KEY!,
});

// Off-chain wallet
const w = await mg.walletAdmin.create({
  subject_id: "user-123",
  subject_type: "user",
  wallet_type: WalletType.Fiat,
  currency: Currency.UZS,
  idempotency_key: "idem-create-1",
});

// On-chain wallet (BIP-39 mnemonic returned ONCE — surface to user, never persist)
const chain = await mg.walletAdmin.createChain({
  subject_id: "user-123",
  subject_type: "user",
  currency: Currency.USDC,
  network: Network.Solana,
});
showOnceToEndUser(chain.mnemonic);

// Deposit address
//   - SPL token: pass `mint` → returns the Associated Token Account.
//   - Native asset: leave `mint` empty → returns the wallet owner address.
const ata = await mg.walletAdmin.depositAddress(
  chain.wallet.wallet_id,
  Network.Solana,
  Mint.USDCSolanaMainnet,
);
const sol = await mg.walletAdmin.depositAddress(
  chain.wallet.wallet_id,
  Network.Solana,
  "",
);

// Withdraw — `mint` selects SPL token, empty / undefined = native SOL.
const tx = await mg.walletAdmin.withdraw(chain.wallet.wallet_id, {
  amount: "10.50",
  destination_type: "crypto_address",
  destination_id: "RecipientSolanaAddr",
  network: Network.Solana,
  mint: Mint.USDCSolanaMainnet,
  idempotency_key: "idem-w-1",
});

// Compliance / fraud
await mg.walletAdmin.freeze(w.wallet_id, "fraud-investigation");
await mg.walletAdmin.unfreeze(w.wallet_id, "case-resolved");

// Pagination — opaque cursor; empty cursor = first page.
let resp = await mg.walletAdmin.list({ limit: 50 });
while (resp.next_cursor) {
  resp = await mg.walletAdmin.list({ limit: 50, cursor: resp.next_cursor });
}

const single = await mg.walletAdmin.getTransaction(w.wallet_id, "tx-xxx");
```

### Typed constants

`Currency`, `Network`, `Mint`, `WalletType`, `WalletStatus`,
`TransactionType`, `TransactionStatus`, `TransactionReason` are
const-as-object enums — IDE autocomplete on known values, literal-union
type for compile-time check, plain JSON string on the wire. Untyped
string literals (`"USDC"`, `"SOLANA"`) stay assignable, so existing
callers don't break.

```ts
Currency.USDC; // "USDC"
Network.Solana; // "SOLANA"
Mint.USDCSolanaMainnet; // "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"
```

`Mint` allows arbitrary strings outside the listed mainnet whitelist —
server-side (`chain-rpc.derive_spl_ata`) is the authoritative validator.

## AI — языковая модель платформы

Модель одна на установку и живёт на самом сервере; ключ к ней не покидает
платформу, поэтому обращаться нужно сюда, а не к провайдеру напрямую. Требуется
включённый тенанту модуль `mgAI` — иначе вызовы дают 403.

Определяющее ограничение — **префилл дороже генерации**: обработка входа
занимает больше времени, чем сам ответ, и контекст в полторы тысячи токенов
держит соединение около минуты. Отсюда два пути.

### Короткий вызов

Пока вход укладывается в сотни токенов — классификация, извлечение полей,
короткая сводка:

```ts
const res = await mg.ai.complete({
  system: "Ты классифицируешь заявки. Отвечай строго по схеме.",
  user: "Нужно 40 мешков цемента М400 на объект в Худжанде к пятнице",
  jsonSchema: JSON.stringify({
    type: "object",
    properties: {
      товар: { type: "string" },
      количество: { type: "integer" },
      срочность: { type: "string", enum: ["низкая", "обычная", "высокая"] },
    },
    required: ["товар", "количество", "срочность"],
  }),
  maxTokens: 200,
});

console.log(JSON.parse(res.text), res.cachedTokens, "токенов из кеша");
```

`jsonSchema` — единственный способ получить строгий JSON: описать схему на
стороне модели через файл грамматики в этой сборке нельзя.

### Длинное задание

Всё остальное. `submit` возвращает задание сразу, `wait` опрашивает его:

```ts
const job = await mg.ai.submit({
  request: { system: SYSTEM_PROMPT, user: longText, maxTokens: 1500 },
});

const done = await mg.ai.wait(job.jobId, { timeoutMs: 10 * 60_000 });

if (done.state === "JOB_STATE_FAILED") {
  console.error("модель отказала:", done.error);
} else {
  console.log(done.result?.text);
}
```

Отказ модели — это разрешившееся задание со `state: "JOB_STATE_FAILED"`, а не
исключение: проверяйте состояние. Исключение из `wait` означает, что ответа не
дождались или вызов отменили через `signal`.

Если опрашивать не хочется, передайте `callbackUrl` — платформа сама постучится,
когда задание будет готово.

### Кеш префилла

Системную часть промпта держите **неизменной** между вызовами одного сценария:
совпадение системного промпта экономит около трети времени. Подстановка туда
даты, имени пользователя или идентификатора заказа тихо ломает кеш — запрос
отработает, но заметно дольше. Всё переменное кладите в `user`.

Насколько кеш сработал, видно по `cachedTokens` в ответе.

### Векторы

```ts
const { embeddings } = await mg.ai.embed([
  "Цемент М400, мешок 50 кг",
  "Цемент портландский, 50 кг",
]);
```

Отдельный вызов, а не режим `complete`: у эмбеддингов другая модель, другая цена
и другой результат.

### Живость

```ts
const st = await mg.ai.status();
// { available: true, model: "Qwen3-30B-A3B", contextSize: 8192, queueDepth: 2 }
```

Модель обслуживает один запрос за раз, поэтому `queueDepth` — единственный
честный признак того, сколько ждать. Проверяйте его, чтобы отличить «модель не
ответила» от «модель ответила плохо».

## Errors

Non-2xx responses raise `MashgateError`:

```ts
import { MashgateError } from "@mashgate/sdk";

try {
  await mg.walletAdmin.get("missing");
} catch (e) {
  if (e instanceof MashgateError) {
    console.error(e.status, e.code, e.message);
  }
}
```

## Webhooks

```ts
import { verifyWebhookSignature } from "@mashgate/sdk";

app.post("/webhooks/mashgate", express.raw({ type: "application/json" }), async (req, res) => {
  const ok = await verifyWebhookSignature(
    req.body,
    req.header("x-hl-signature")!,
    process.env.WEBHOOK_SECRET!,
    req.header("x-hl-timestamp")!,
  );
  if (!ok) return res.status(401).end();
  // handle event
});
```

## Development

```bash
npm install
npm run build           # tsc → dist/
npm test                # vitest run
```
