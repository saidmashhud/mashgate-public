import type { MashgateClient } from "../client.js";

// ── Types ────────────────────────────────────────────────────────────────

export interface CryptoWallet {
  walletId: string;
  tenantId: string;
  userId: string;
  walletType: string;
  addresses: WalletAddress[];
  createdAt: string;
}

export interface WalletAddress {
  network: string;
  address: string;
  derivationPath: string;
}

export interface AssetBalance {
  asset: string;
  network: string;
  balance: string;
  usdValue: string;
}

export interface CryptoPayment {
  paymentId: string;
  tenantId: string;
  status: string;
  amount: string;
  asset: string;
  network: string;
  depositAddress: string;
  txHash?: string;
  confirmations: number;
  requiredConfirmations: number;
  expiresAt: string;
  metadata?: Record<string, string>;
  createdAt: string;
}

export interface SwapResult {
  swapId: string;
  status: string;
  fromAmount: string;
  fromAsset: string;
  fromNetwork: string;
  toAmount: string;
  toAsset: string;
  toNetwork: string;
  exchangeRate: string;
  fee: string;
  txHash?: string;
  createdAt: string;
}

export interface Escrow {
  escrowId: string;
  tenantId: string;
  status: string;
  amount: string;
  asset: string;
  network: string;
  contractAddress?: string;
  payerAddress: string;
  payeeAddress: string;
  arbiterAddress?: string;
  txHash?: string;
  releaseAfter?: number;
  expiresAt?: number;
  createdAt: string;
  releasedAt?: string;
}

export interface OnRampResult {
  rampId: string;
  status: string;
  fiatAmount: string;
  fiatCurrency: string;
  cryptoAmount: string;
  cryptoAsset: string;
  cryptoNetwork: string;
  exchangeRate: string;
  fee: string;
  redirectUrl?: string;
  createdAt: string;
}

export interface OffRampResult {
  rampId: string;
  status: string;
  cryptoAmount: string;
  cryptoAsset: string;
  fiatAmount: string;
  fiatCurrency: string;
  exchangeRate: string;
  fee: string;
  createdAt: string;
}

export interface ScreenResult {
  result: string;
  riskScore: string;
  flags: string[];
  provider: string;
  screenedAt: string;
}

export interface GasEstimate {
  network: string;
  gasPrice: string;
  usdCost: string;
  fastGas: string;
  slowGas: string;
}

export interface ExchangeRate {
  fromAsset: string;
  toAsset: string;
  rate: string;
  inverse: string;
  source: string;
  updatedAt: string;
}

export interface BatchPayout {
  payoutId: string;
  status: string;
  totalRecipients: number;
  completed: number;
  failed: number;
  totalAmount: string;
  totalFee: string;
  txHash?: string;
  createdAt: string;
}

// ── Resource ─────────────────────────────────────────────────────────────

export class ChainResource {
  constructor(private readonly client: MashgateClient) {}

  // Wallets
  async createWallet(data: {
    tenantId: string; userId: string; walletType: string;
    networks: string[]; label?: string;
  }): Promise<CryptoWallet> {
    return this.client.request("POST", "/v1/chain/wallets", { body: data });
  }

  async getWallet(walletId: string): Promise<CryptoWallet> {
    return this.client.request("GET", `/v1/chain/wallets/${walletId}`);
  }

  async getWalletBalance(walletId: string): Promise<{ balances: AssetBalance[] }> {
    return this.client.request("GET", `/v1/chain/wallets/${walletId}/balance`);
  }

  // Crypto Payments
  async pay(data: {
    amount: string; asset: string; network: string;
    orderId?: string; customerId?: string; destinationAddress?: string;
    metadata?: Record<string, string>;
  }): Promise<CryptoPayment> {
    return this.client.request("POST", "/v1/chain/payments", { body: data });
  }

  async getCryptoPayment(paymentId: string): Promise<CryptoPayment> {
    return this.client.request("GET", `/v1/chain/payments/${paymentId}`);
  }

  // Swaps
  async swap(data: {
    fromAmount: string; fromAsset: string; fromNetwork: string;
    toAsset: string; toNetwork: string; slippageBps?: string;
    destinationAddress?: string;
  }): Promise<SwapResult> {
    return this.client.request("POST", "/v1/chain/swaps", { body: data });
  }

  async getSwap(swapId: string): Promise<SwapResult> {
    return this.client.request("GET", `/v1/chain/swaps/${swapId}`);
  }

  // Escrow
  async createEscrow(data: {
    amount: string; asset: string; network: string;
    payerAddress: string; payeeAddress: string; arbiterAddress?: string;
    releaseAfter?: number; expiresAt?: number; useCase?: string;
    metadata?: Record<string, string>;
  }): Promise<Escrow> {
    return this.client.request("POST", "/v1/chain/escrows", { body: data });
  }

  async getEscrow(escrowId: string): Promise<Escrow> {
    return this.client.request("GET", `/v1/chain/escrows/${escrowId}`);
  }

  async releaseEscrow(escrowId: string, releasedBy: string): Promise<Escrow> {
    return this.client.request("POST", `/v1/chain/escrows/${escrowId}/release`, { body: { releasedBy } });
  }

  async disputeEscrow(escrowId: string, reason: string): Promise<Escrow> {
    return this.client.request("POST", `/v1/chain/escrows/${escrowId}/dispute`, { body: { reason } });
  }

  // On-Ramp / Off-Ramp
  async onRamp(data: {
    fiatAmount: string; fiatCurrency: string; targetAsset: string;
    targetNetwork: string; destinationAddress: string; provider?: string;
  }): Promise<OnRampResult> {
    return this.client.request("POST", "/v1/chain/on-ramp", { body: data });
  }

  async offRamp(data: {
    cryptoAmount: string; cryptoAsset: string; cryptoNetwork: string;
    targetCurrency: string; payoutMethod: string; payoutDetails: string;
    provider?: string;
  }): Promise<OffRampResult> {
    return this.client.request("POST", "/v1/chain/off-ramp", { body: data });
  }

  // Compliance
  async screenAddress(address: string, network: string): Promise<ScreenResult> {
    return this.client.request("POST", "/v1/chain/compliance/screen-address", { body: { address, network } });
  }

  async screenTransaction(txHash: string, network: string): Promise<ScreenResult> {
    return this.client.request("POST", "/v1/chain/compliance/screen-tx", { body: { txHash, network } });
  }

  // Gas & Rates
  async gasEstimate(network: string): Promise<GasEstimate> {
    return this.client.request("GET", `/v1/chain/gas/${network}`);
  }

  async exchangeRate(from: string, to: string): Promise<ExchangeRate> {
    return this.client.request("GET", `/v1/chain/rates/${from}/${to}`);
  }

  // Batch Payouts
  async batchPayout(data: {
    asset: string; network: string;
    recipients: Array<{ address: string; amount: string; reference?: string }>;
  }): Promise<BatchPayout> {
    return this.client.request("POST", "/v1/chain/payouts", { body: data });
  }
}
