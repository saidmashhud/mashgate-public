# Exchange API example

This example uses an end-user access token. Mashgate derives the tenant,
subject, and Exchange account from that token; no account id is accepted from
the caller.

Set `MASHGATE_API_URL`, `MASHGATE_API_KEY`, and `MASHGATE_ACCESS_TOKEN`, then
run one language example. The default path only lists markets and balances.
Set `PLACE_TEST_ORDER=1` only against an isolated testnet account.

Every order, cancel, and withdrawal must use a stable `Idempotency-Key`. Reuse
the same key only when retrying the same logical command.
