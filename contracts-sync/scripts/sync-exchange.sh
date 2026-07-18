#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SOURCE="${MASHGATE_SRC:-$ROOT/../mashgate}"
DEST="$ROOT/contracts-sync/snapshots/exchange"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

[[ -d "$SOURCE/contracts" ]] || { echo "Mashgate contracts not found at $SOURCE" >&2; exit 1; }
mkdir -p "$TMP/exchange" "$DEST"

cp "$SOURCE/contracts/proto/v1/exchange.proto" "$TMP/exchange/exchange.proto"
for name in \
  _envelope.v1.json \
  exchange.order.updated.json \
  exchange.trade.executed.json \
  exchange.deposit.updated.json \
  exchange.withdrawal.updated.json \
  exchange.market.updated.json
do
  cp "$SOURCE/contracts/events/$name" "$TMP/exchange/$name"
done

if [[ "${CHECK_ONLY:-0}" == "1" ]]; then
  diff -ru "$DEST" "$TMP/exchange"
  echo "Exchange contract snapshot is in sync."
  exit 0
fi

rm -rf "$DEST"
mkdir -p "$DEST"
cp "$TMP/exchange"/* "$DEST"/
echo "Synced Exchange contracts from $SOURCE."
