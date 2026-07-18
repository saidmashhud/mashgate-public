#!/usr/bin/env bash
# Mashgate external-alpha gate.
#
# Runs the checks an invited developer should be able to reproduce before using
# Mashgate SDK + HookLine as an integration surface.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
HOOKLINE_REPO="${HOOKLINE_REPO:-$ROOT/../hookline}"
REQUIRE_HOOKLINE="${REQUIRE_HOOKLINE:-1}"
RUN_CONTRACT_SYNC="${RUN_CONTRACT_SYNC:-0}"
SKIP_GO="${SKIP_GO:-0}"
SKIP_TS="${SKIP_TS:-0}"
SKIP_PYTHON="${SKIP_PYTHON:-0}"

PASS=0
FAIL=0
TMP_ROOT=""

green() { printf '\033[32m%s\033[0m\n' "$*"; }
red() { printf '\033[31m%s\033[0m\n' "$*"; }
yellow() { printf '\033[33m%s\033[0m\n' "$*"; }
info() { printf '\033[36m%s\033[0m\n' "$*"; }

cleanup() {
  [[ -n "$TMP_ROOT" && -d "$TMP_ROOT" ]] && rm -rf "$TMP_ROOT"
}
trap cleanup EXIT

need() {
  command -v "$1" >/dev/null 2>&1 || {
    red "missing required command: $1"
    return 1
  }
}

can_run_container() {
  command -v nerdctl >/dev/null 2>&1 && [[ -S /run/k3s/containerd/containerd.sock ]]
}

run_container() {
  local image="$1"
  local workdir="$2"
  local script="$3"
  local path_prefix='export PATH=/usr/local/go/bin:/usr/local/bin:/usr/local/sbin:/usr/bin:/usr/sbin:/bin:/sbin:$PATH; '
  nerdctl --address /run/k3s/containerd/containerd.sock --namespace k8s.io \
    run --net=host --rm -v "$ROOT:$ROOT" -w "$workdir" "$image" sh -lc "${path_prefix}${script}"
}

run() {
  local name="$1"
  shift
  info "-- $name"
  if "$@"; then
    PASS=$((PASS + 1))
    green "PASS: $name"
  else
    FAIL=$((FAIL + 1))
    red "FAIL: $name"
  fi
}

go_sdk() {
  if command -v go >/dev/null 2>&1; then
    (cd "$ROOT/sdk/go" && go test ./...)
    return
  fi
  can_run_container || { red "go not found and container fallback unavailable"; return 1; }
  run_container golang:1.24-alpine "$ROOT/sdk/go" 'go test ./...'
}

go_example() {
  if command -v go >/dev/null 2>&1; then
    (cd "$ROOT/examples/go" && go build -o /tmp/mashgate-go-quickstart .)
    return
  fi
  can_run_container || { red "go not found and container fallback unavailable"; return 1; }
  run_container golang:1.24-alpine "$ROOT/examples/go" 'go build -o /tmp/mashgate-go-quickstart .'
}

go_exchange_example() {
  if command -v go >/dev/null 2>&1; then
    (cd "$ROOT/examples/exchange/go" && go build -o /tmp/mashgate-go-exchange-example .)
    return
  fi
  can_run_container || { red "go not found and container fallback unavailable"; return 1; }
  run_container golang:1.24-alpine "$ROOT/examples/exchange/go" 'go build -o /tmp/mashgate-go-exchange-example .'
}

ts_sdk() {
  need npm || return 1
  (cd "$ROOT/sdk/typescript" && npm ci && npm run build && npm test)
}

ts_example() {
  need npm || return 1
  (cd "$ROOT/examples/typescript" && npm ci && npm run build)
}

ts_exchange_example() {
  need npm || return 1
  (cd "$ROOT/examples/exchange/typescript" && npm install --package-lock=false && npm run build)
}

python_sdk() {
  if python_sdk_host; then
    return
  fi
  can_run_container || { red "python venv/pip unavailable and container fallback unavailable"; return 1; }
  run_container python:3.12-alpine "$ROOT/sdk/python" \
    'python -m pip install --upgrade pip >/dev/null &&
     python -m pip install -e .[dev] >/dev/null &&
     rm -rf ./*.egg-info &&
     python -m pytest &&
     PYTHONPATH=. python -m py_compile ../../examples/python/quickstart.py &&
     PYTHONPATH=. python -m py_compile ../../examples/exchange/python/main.py &&
     PYTHONPATH=. python - <<'"'"'PY'"'"'
from mashgate import MashgateClient, verify_webhook_signature
assert MashgateClient is not None
assert callable(verify_webhook_signature)
PY'
}

python_sdk_host() {
  need python3 || return 1
  TMP_ROOT="${TMP_ROOT:-$(mktemp -d)}"
  local venv="$TMP_ROOT/mashgate-python"
  rm -rf "$venv"
  python3 -m venv "$venv" >/dev/null 2>&1 || return 1
  "$venv/bin/python" -m pip install --upgrade pip >/dev/null || return 1
  "$venv/bin/python" -m pip install -e "$ROOT/sdk/python[dev]" >/dev/null || return 1
  (cd "$ROOT/sdk/python" && "$venv/bin/python" -m pytest)
  PYTHONPATH="$ROOT/sdk/python" "$venv/bin/python" -m py_compile \
    "$ROOT/examples/python/quickstart.py" \
    "$ROOT/examples/exchange/python/main.py"
  PYTHONPATH="$ROOT/sdk/python" "$venv/bin/python" - <<'PY'
from mashgate import MashgateClient, verify_webhook_signature
assert MashgateClient is not None
assert callable(verify_webhook_signature)
PY
}

contract_snapshot() {
  local manifest="$ROOT/contracts-sync/manifests/active.yaml"
  test -s "$manifest"
  grep -q "head_ref:" "$manifest"
  grep -q "sdk_versions:" "$manifest"
  test -s "$ROOT/sdk/go/_generated/openapi.yaml"
  test -s "$ROOT/sdk/go/_generated/types.gen.go"
  test -s "$ROOT/sdk/typescript/src/_generated/openapi.yaml"
  test -s "$ROOT/sdk/typescript/src/_generated/types.ts"
  test -s "$ROOT/contracts-sync/manifests/exchange-alpha.yaml"
  test -s "$ROOT/contracts-sync/snapshots/exchange/exchange.proto"
  for event in order trade deposit withdrawal market; do
    test -s "$ROOT/contracts-sync/snapshots/exchange/exchange.${event}.updated.json" || \
      test -s "$ROOT/contracts-sync/snapshots/exchange/exchange.${event}.executed.json"
  done
}

contract_sync_check() {
  CHECK_ONLY=1 "$ROOT/contracts-sync/scripts/sync.sh"
}

hookline_gate() {
  if [[ ! -d "$HOOKLINE_REPO" ]]; then
    if [[ "$REQUIRE_HOOKLINE" == "1" ]]; then
      red "HookLine repo not found: $HOOKLINE_REPO"
      return 1
    fi
    yellow "SKIP: HookLine repo not found: $HOOKLINE_REPO"
    return 0
  fi
  bash "$HOOKLINE_REPO/test/sdk-alpha-smoke.sh"
}

docs_contract() {
  test -s "$ROOT/docs/external-alpha.md"
  test -s "$ROOT/docs/guides/building-a-vertical.md"
  test -s "$ROOT/docs/modules/events-webhooks.md"
  grep -q "external-alpha-smoke.sh" "$ROOT/docs/external-alpha.md"
  grep -q "HookLine" "$ROOT/docs/external-alpha.md"
}

info "== Mashgate external-alpha gate =="
info "repo: $ROOT"
info "hookline: $HOOKLINE_REPO"

[[ "$SKIP_GO" == "1" ]] || run "Go SDK tests" go_sdk
[[ "$SKIP_GO" == "1" ]] || run "Go quickstart builds" go_example
[[ "$SKIP_GO" == "1" ]] || run "Go Exchange example builds" go_exchange_example
[[ "$SKIP_TS" == "1" ]] || run "TypeScript SDK build + tests" ts_sdk
[[ "$SKIP_TS" == "1" ]] || run "TypeScript quickstart builds" ts_example
[[ "$SKIP_TS" == "1" ]] || run "TypeScript Exchange example builds" ts_exchange_example
[[ "$SKIP_PYTHON" == "1" ]] || run "Python SDK install + tests" python_sdk
run "contract snapshot presence" contract_snapshot
[[ "$RUN_CONTRACT_SYNC" == "1" ]] && run "contract sync check" contract_sync_check
run "HookLine SDK alpha gate" hookline_gate
run "external alpha docs contract" docs_contract

info "== Results: $PASS passed, $FAIL failed =="
[[ "$FAIL" -eq 0 ]]
