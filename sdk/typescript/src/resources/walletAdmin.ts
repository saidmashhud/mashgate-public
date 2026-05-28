// Admin/merchant-side WalletService client.
//
// Mirrors `wallet.v1.WalletService` from
// mashgate/contracts/proto/v1/wallet.proto, exposed over the gateway as
// REST via google.api.http transcoding. End-user wallet operations
// (balance, saved payment methods) live in `WalletResource` / wallet.ts.
//
// Auth here is tenant-scoped — pass an admin JWT or service-account API
// key when constructing the client.

import type { MashgateClient } from "../client.js";

// ── Typed constants ──────────────────────────────────────────────────────────
//
// Pattern: const-as-object + derived union type. Gives autocomplete on
// known values, compile-time type check, and a wire format identical to
// plain strings ("USDC" not {value:"USDC"}). Untyped string literals
// stay assignable, so existing callers using "USDC" don't break.

export const Currency = {
  // Fiat (ISO 4217)
  UZS: "UZS",
  KZT: "KZT",
  KGS: "KGS",
  TJS: "TJS",
  RUB: "RUB",
  USD: "USD",
  EUR: "EUR",
  // Crypto / stablecoin tickers
  USDT: "USDT",
  USDC: "USDC",
  SOL: "SOL",
  ETH: "ETH",
  TRX: "TRX",
  BNB: "BNB",
  TON: "TON",
} as const;
export type Currency = (typeof Currency)[keyof typeof Currency];

export const Network = {
  Solana: "SOLANA",
  Ethereum: "ETHEREUM",
  Base: "BASE",
  Polygon: "POLYGON",
  BSC: "BSC",
  Tron: "TRON",
  TON: "TON",
} as const;
export type Network = (typeof Network)[keyof typeof Network];

// Token contract / mint addresses keyed by (network, asset). Empty `Mint`
// = native asset path. Solana entries are SPL token mints; TRON entries
// are TRC-20 contract addresses (base58check). The field name стало
// chain-agnostic: ledger-core picks the right interpretation based on
// the wallet's `network` column.
export const Mint = {
  USDCSolanaMainnet: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
  USDTSolanaMainnet: "Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB",
  USDTTronMainnet: "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",
  // EVM mainnet token contracts (0x-hex). ledger-core branches on the
  // wallet's `network` to pick the EIP-1559 transfer path.
  USDTEthereumMainnet: "0xdAC17F958D2ee523a2206206994597C13D831ec7",
  USDCEthereumMainnet: "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",
  USDTBscMainnet: "0x55d398326f99059fF775485246999027B3197955",
  USDCBscMainnet: "0x8AC76a51cc950d9822D68b83fE1Ad97B32Cd580d",
  USDTPolygonMainnet: "0xc2132D05D31c914a87C6611C10748AEb04B58e8F",
  USDCPolygonMainnet: "0x3c499c542cEF5E3811e1192ce70d8cC03d5c3359",
  USDCBaseMainnet: "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
} as const;
export type Mint = (typeof Mint)[keyof typeof Mint] | "" | string;

// ── Enum-like string types from wallet.proto ────────────────────────────────

export const WalletType = {
  Fiat: "WALLET_TYPE_FIAT",
  Crypto: "WALLET_TYPE_CRYPTO",
  Omnibus: "WALLET_TYPE_OMNIBUS",
} as const;
export type WalletType = (typeof WalletType)[keyof typeof WalletType];

export const WalletStatus = {
  Active: "WALLET_STATUS_ACTIVE",
  Frozen: "WALLET_STATUS_FROZEN",
  Closed: "WALLET_STATUS_CLOSED",
} as const;
export type WalletStatus = (typeof WalletStatus)[keyof typeof WalletStatus];

export const TransactionType = {
  Credit: "TRANSACTION_TYPE_CREDIT",
  Debit: "TRANSACTION_TYPE_DEBIT",
} as const;
export type TransactionType = (typeof TransactionType)[keyof typeof TransactionType];

export const TransactionStatus = {
  Pending: "TRANSACTION_STATUS_PENDING",
  Settled: "TRANSACTION_STATUS_SETTLED",
  Failed: "TRANSACTION_STATUS_FAILED",
  Reversed: "TRANSACTION_STATUS_REVERSED",
} as const;
export type TransactionStatus = (typeof TransactionStatus)[keyof typeof TransactionStatus];

export const TransactionReason = {
  Deposit: "TRANSACTION_REASON_DEPOSIT",
  Withdrawal: "TRANSACTION_REASON_WITHDRAWAL",
  Payment: "TRANSACTION_REASON_PAYMENT",
  Refund: "TRANSACTION_REASON_REFUND",
  Settlement: "TRANSACTION_REASON_SETTLEMENT",
  Fee: "TRANSACTION_REASON_FEE",
  Adjustment: "TRANSACTION_REASON_ADJUSTMENT",
  Conversion: "TRANSACTION_REASON_CONVERSION",
} as const;
export type TransactionReason = (typeof TransactionReason)[keyof typeof TransactionReason];

// ── Domain types ────────────────────────────────────────────────────────────

export interface Wallet {
  wallet_id: string;
  tenant_id: string;
  subject_id: string;
  subject_type: "user" | "merchant";
  wallet_type: WalletType;
  status: WalletStatus;
  currency: Currency;
  // Money fields are decimal strings to avoid JS Number precision loss.
  balance: string;
  pending: string;
  freeze_reason?: string;
  frozen_by?: string;
  created_at: string;
  updated_at: string;
  frozen_at?: string;
}

export interface WalletTransaction {
  transaction_id: string;
  wallet_id: string;
  tenant_id: string;
  type: TransactionType;
  status: TransactionStatus;
  reason: TransactionReason;
  amount: string;
  currency: Currency;
  balance_after: string;
  reference_id?: string;
  reference_type?: string;
  description?: string;
  external_ref?: string;
  metadata?: Record<string, string>;
  created_at: string;
  settled_at?: string;
}

export interface DepositAddress {
  wallet_id: string;
  currency: Currency;
  network: Network;
  address: string;
  memo?: string;
  expires_at?: string;
}

// ── Request shapes ──────────────────────────────────────────────────────────

export interface CreateWalletRequest {
  subject_id: string;
  subject_type: "user" | "merchant";
  wallet_type: WalletType;
  currency: Currency;
  idempotency_key?: string;
}

// L1 of ADR-0016. Returns `{wallet, mnemonic}` — `mnemonic` is shown to
// the end user ONCE; never persisted server-side beyond a SHA-256 hash.
export interface CreateChainWalletRequest {
  subject_id: string;
  subject_type: "user" | "merchant";
  currency: Currency;
  network: Network;
  idempotency_key?: string;
}

export interface CreateChainWalletResponse {
  wallet: Wallet;
  mnemonic: string;
}

// L1 of ADR-0016 recovery flow. Re-import the same mnemonic for the same
// subject within a tenant returns the existing wallet с `was_existing=true`
// (idempotent). Re-import под другим subject_id returns
// `PERMISSION_DENIED` — credential-stuffing protection.
export interface ImportChainWalletRequest {
  subject_id: string;
  subject_type: "user" | "merchant";
  currency: Currency;
  network: Network;
  // BIP-39 phrase (12 or 24 words). Validated server-side; SHA-256-hashed
  // before storage. **Clear from caller memory immediately после успешного
  // ответа** — phrase is never persisted plaintext.
  mnemonic: string;
  idempotency_key?: string;
}

export interface ImportChainWalletResponse {
  wallet: Wallet;
  // True when import resolved to a pre-existing row (same mnemonic + same
  // subject in the same tenant) — recovery rather than fresh creation.
  // Frontends use this to render "✓ wallet recovered" vs "✓ wallet imported".
  was_existing: boolean;
  // ISO-8601. Mirror of wallet.created_at on fresh import or wallet.updated_at
  // on recovery. May be omitted by older servers.
  recovered_at?: string;
}

export interface CreditDebitRequest {
  amount: string; // decimal string — e.g. "100.50"
  reason: TransactionReason;
  reference_id?: string;
  reference_type?: string;
  description?: string;
  idempotency_key?: string;
  merchant_id?: string;
}

export interface InitiateWithdrawalRequest {
  amount: string;
  destination_type: "crypto_address" | "bank_account";
  destination_id: string;
  network?: Network;
  description?: string;
  idempotency_key?: string;
  // SPL token mint. Empty / undefined = native SOL. L2 of ADR-0016.
  mint?: Mint;
  // Optional sponsor wallet. When provided, the platform sponsor pays the
  // chain fee + any SPL ATA rent instead of the source — used for gasless
  // withdrawals. Must be in the same tenant + same network, active, on-chain.
  sponsor_wallet_id?: string;
}

export interface ListWalletsQuery {
  subject_id?: string;
  status?: WalletStatus;
  wallet_type?: WalletType;
  limit?: number;
  cursor?: string; // opaque, from a previous response
}

export interface ListWalletsResponse {
  wallets: Wallet[];
  next_cursor?: string;
}

export interface ListTransactionsQuery {
  type?: TransactionType;
  status?: TransactionStatus;
  reason?: TransactionReason;
  limit?: number;
  cursor?: string;
}

export interface ListTransactionsResponse {
  transactions: WalletTransaction[];
  next_cursor?: string;
}

// Atomic inter-wallet transfer (same tenant, same currency). v1 of
// `wallet.v1.WalletService.TransferBetweenWallets`. The server commits
// a single Postgres tx: balance delta on both wallets + two
// wallet_transactions rows + three outbox events (wallet.debit,
// wallet.credit, wallet.transfer). Idempotent via `idempotency_key`.
export interface TransferBetweenWalletsRequest {
  to_wallet_id: string;
  // Decimal string. Must be > 0.
  amount: string;
  // Defaults to TRANSACTION_REASON_ADJUSTMENT if omitted.
  reason?: TransactionReason;
  description?: string;
  idempotency_key?: string;
  // Surfaced in the paired wallet.{credit,debit} envelope events. Empty
  // skips merchant_id on those events (money state still commits).
  merchant_id?: string;
  // Optional free-text on the wallet.transfer envelope.
  note?: string;
}

export interface TransferBetweenWalletsResponse {
  transfer_id: string;
  debit: WalletTransaction;
  credit: WalletTransaction;
}

// ── Resource ────────────────────────────────────────────────────────────────

export class WalletAdminResource {
  constructor(private readonly client: MashgateClient) {}

  /** Create an off-chain wallet (fiat / omnibus). */
  create(req: CreateWalletRequest): Promise<Wallet> {
    return this.client.request<Wallet>("POST", "/v1/wallets", { body: req });
  }

  /**
   * Create a non-custodial on-chain wallet. The returned `mnemonic` is
   * shown to the end user ONCE; surface it immediately and never persist
   * it. Currently SOLANA only. L1 of ADR-0016.
   */
  createChain(req: CreateChainWalletRequest): Promise<CreateChainWalletResponse> {
    return this.client.request<CreateChainWalletResponse>("POST", "/v1/wallets/chain", {
      body: req,
    });
  }

  /**
   * Recover / import a non-custodial wallet from a user-provided BIP-39
   * mnemonic. The mnemonic must be cleared from caller memory immediately
   * after the call returns — it touches process memory briefly but is
   * never persisted plaintext server-side. Currently SOLANA only.
   *
   * - Re-importing the same phrase for the same subject returns the
   *   existing wallet с `was_existing=true` (idempotent recovery).
   * - Re-importing under a different subject_id within the tenant returns
   *   `PERMISSION_DENIED` (credential-stuffing protection).
   */
  importChain(req: ImportChainWalletRequest): Promise<ImportChainWalletResponse> {
    return this.client.request<ImportChainWalletResponse>(
      "POST",
      "/v1/wallets/chain/import",
      { body: req },
    );
  }

  get(walletId: string): Promise<Wallet> {
    return this.client.request<Wallet>("GET", `/v1/wallets/${walletId}`);
  }

  list(query?: ListWalletsQuery): Promise<ListWalletsResponse> {
    return this.client.request<ListWalletsResponse>("GET", "/v1/wallets", {
      query: {
        subject_id: query?.subject_id,
        status: query?.status,
        wallet_type: query?.wallet_type,
        limit: query?.limit,
        cursor: query?.cursor,
      },
    });
  }

  /** Halt all debits (compliance / fraud trigger). `reason` is required and audited. */
  freeze(walletId: string, reason: string): Promise<Wallet> {
    return this.client.request<Wallet>("POST", `/v1/wallets/${walletId}/freeze`, {
      body: { freeze_reason: reason },
    });
  }

  /** Restore a frozen wallet. `note` is optional, audited. */
  unfreeze(walletId: string, note?: string): Promise<Wallet> {
    return this.client.request<Wallet>("POST", `/v1/wallets/${walletId}/unfreeze`, {
      body: { note: note ?? "" },
    });
  }

  credit(walletId: string, req: CreditDebitRequest): Promise<WalletTransaction> {
    return this.client.request<WalletTransaction>("POST", `/v1/wallets/${walletId}/credit`, {
      body: req,
    });
  }

  debit(walletId: string, req: CreditDebitRequest): Promise<WalletTransaction> {
    return this.client.request<WalletTransaction>("POST", `/v1/wallets/${walletId}/debit`, {
      body: req,
    });
  }

  /**
   * Initiate a withdrawal. Pass `mint` for SPL tokens, leave empty for
   * native SOL. L2 of ADR-0016.
   */
  withdraw(walletId: string, req: InitiateWithdrawalRequest): Promise<WalletTransaction> {
    return this.client.request<WalletTransaction>("POST", `/v1/wallets/${walletId}/withdraw`, {
      body: req,
    });
  }

  listTransactions(
    walletId: string,
    query?: ListTransactionsQuery,
  ): Promise<ListTransactionsResponse> {
    return this.client.request<ListTransactionsResponse>(
      "GET",
      `/v1/wallets/${walletId}/transactions`,
      {
        query: {
          type: query?.type,
          status: query?.status,
          reason: query?.reason,
          limit: query?.limit,
          cursor: query?.cursor,
        },
      },
    );
  }

  getTransaction(walletId: string, transactionId: string): Promise<WalletTransaction> {
    return this.client.request<WalletTransaction>(
      "GET",
      `/v1/wallets/${walletId}/transactions/${transactionId}`,
    );
  }

  /**
   * Atomically transfer `req.amount` from `fromWalletId` to
   * `req.to_wallet_id`. Both wallets must belong to the same tenant and
   * share currency. Server errors mapped from gRPC status:
   * - INVALID_ARGUMENT — same wallet IDs, currency mismatch, non-positive amount.
   * - FAILED_PRECONDITION — source or destination frozen, insufficient balance.
   * - PERMISSION_DENIED — wallets belong to different tenants.
   * - NOT_FOUND — wallet does not exist.
   */
  transfer(
    fromWalletId: string,
    req: TransferBetweenWalletsRequest,
  ): Promise<TransferBetweenWalletsResponse> {
    return this.client.request<TransferBetweenWalletsResponse>(
      "POST",
      `/v1/wallets/${fromWalletId}/transfer`,
      { body: req },
    );
  }

  /**
   * Deposit address resolver. For SPL tokens pass a non-empty `mint` and
   * the gateway returns the Associated Token Account derived from
   * (wallet_owner, mint). For native assets leave `mint` empty —
   * the wallet owner address is returned. L3 of ADR-0016.
   */
  depositAddress(
    walletId: string,
    network: Network,
    mint?: Mint,
  ): Promise<DepositAddress> {
    return this.client.request<DepositAddress>(
      "GET",
      `/v1/wallets/${walletId}/deposit-address`,
      {
        query: {
          network,
          mint: mint && mint.length > 0 ? (mint as string) : undefined,
        },
      },
    );
  }
}
