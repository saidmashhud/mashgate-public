#!/usr/bin/env bash
# Pinned contract snapshot sync.
#
# Reads contracts-sync/manifests/active.yaml, downloads the pinned snapshot
# from the core Mashgate monorepo, and regenerates contracts-sync/generated/.
#
# Usage:
#   ./contracts-sync/scripts/sync.sh

set -euo pipefail

MANIFEST="contracts-sync/manifests/active.yaml"
if [[ ! -f "$MANIFEST" ]]; then
  echo "Missing $MANIFEST" >&2
  exit 1
fi

# TODO(v0.2): parse YAML (yq), fetch tag archive, run protoc + openapi-generator.
# For now, this is a stub — the SDKs in sdk/go/, sdk/typescript/, sdk/python/
# were hand-migrated from mashgate/sdk/ at repo bootstrap time.

echo "Contract sync not yet automated — see ROADMAP.md v0.2."
echo "Until then, update SDK types manually when mashgate/contracts/proto/*.proto changes."
exit 0
