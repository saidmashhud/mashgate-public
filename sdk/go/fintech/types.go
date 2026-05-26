package fintech

// Typed string aliases for currency, network, and SPL token mint values.
// Wire format is plain JSON string (these types serialize transparently),
// but the typed surface gives callers IDE autocomplete on known values
// and a compile-time hint when assigning from untyped variables. Untyped
// string literals like `fintech.Currency("XYZ")` still work, so this is
// not a hard whitelist — it's a guide-rail. Server-side validation is
// authoritative.

// ── Currency ─────────────────────────────────────────────────────────────────

// Currency identifies the unit of account a wallet holds. Fiat values
// follow ISO 4217; crypto values use the on-chain ticker / symbol used
// across the Mashgate stack.
type Currency string

// Fiat. ISO 4217.
const (
	CurrencyUZS Currency = "UZS" // Uzbekistani Som
	CurrencyKZT Currency = "KZT" // Kazakhstani Tenge
	CurrencyKGS Currency = "KGS" // Kyrgyzstani Som
	CurrencyTJS Currency = "TJS" // Tajikistani Somoni
	CurrencyRUB Currency = "RUB" // Russian Ruble
	CurrencyUSD Currency = "USD"
	CurrencyEUR Currency = "EUR"
)

// Crypto / stablecoin tickers. The same ticker on different networks is
// considered the same Currency here — `Network` plus `Mint` (for SPL)
// disambiguates which on-chain asset is meant.
const (
	CurrencyUSDT Currency = "USDT"
	CurrencyUSDC Currency = "USDC"
	CurrencySOL  Currency = "SOL"
	CurrencyETH  Currency = "ETH"
	CurrencyTRX  Currency = "TRX"
	CurrencyBNB  Currency = "BNB"
	CurrencyTON  Currency = "TON"
)

// ── Network ──────────────────────────────────────────────────────────────────

// Network identifies the blockchain a crypto wallet operates on. Mirrors
// the `network` field accepted by chain-rpc and ledger-core (uppercase).
type Network string

const (
	NetworkSolana   Network = "SOLANA"
	NetworkEthereum Network = "ETHEREUM"
	NetworkBase     Network = "BASE"
	NetworkPolygon  Network = "POLYGON"
	NetworkBSC      Network = "BSC"
	NetworkTron     Network = "TRON"
	NetworkTON      Network = "TON"
)

// ── SPL Mint ─────────────────────────────────────────────────────────────────

// Mint is an SPL token mint address on Solana. Used as the `mint` field in
// `DepositAddress` / `WithdrawRequest` to select an SPL token versus the
// native SOL transfer path. Empty Mint = native asset.
type Mint string

// Mainnet mints for stablecoins commonly handled by Mashgate. These values
// are well-known and immutable for the life of the token program.
const (
	MintUSDCSolanaMainnet Mint = "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"
	MintUSDTSolanaMainnet Mint = "Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB"
	// TRC-20 USDT on TRON mainnet. For TRON withdrawals the `Mint` field
	// carries the contract address (base58check "T..."), not an SPL mint —
	// the type alias is shared because both networks use the same SDK
	// shape: `Withdraw(req)` selects between native and token transfer
	// by presence/absence of this field, and server-side ledger-core
	// branches on `Network` to know how to interpret the contents.
	MintUSDTTronMainnet Mint = "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t"
)

// String implements fmt.Stringer for callers that pass these into formatters
// or log fields — keeps logs readable without a manual cast.
func (c Currency) String() string { return string(c) }
func (n Network) String() string  { return string(n) }
func (m Mint) String() string     { return string(m) }
