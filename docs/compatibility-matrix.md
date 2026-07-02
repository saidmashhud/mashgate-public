# Compatibility matrix

SDKs version **independently per language** (semver). Each SDK release pins to a
Mashgate contract snapshot (`contracts-sync/manifests/`); any SDK on a given
contract snapshot is compatible with any Mashgate platform version that serves
that contract major.

| SDK | Package | Min runtime | Status | Contract snapshot |
|-----|---------|-------------|--------|-------------------|
| Go | `github.com/saidmashhud/mashgate-public/sdk/go` | Go 1.22 | stable (v0.x) | v1 |
| TypeScript | `@mashgate/sdk` (npm) | Node 18 | stable (v0.x) | v1 |
| Python | `mashgate` (PyPI) | Python 3.10+ | stable (v0.x) | v1 |

## Module coverage

Go, TypeScript, and Python expose the full v1 resource namespace set. Python is
hand-maintained rather than generated, so `sdk/python/tests/test_client_parity.py`
is the parity gate: it fails if any of the 25 resource namespaces is missing,
renamed, or left unwired.

## Contract snapshots

A platform contract major (`v1`) is forward-compatible: new RPCs/fields are
additive. An SDK pinned to a `v1` snapshot keeps working as the platform adds
`v1` RPCs; it only needs an upgrade to call the newly-added ones. A platform
`v2` would ship alongside `v1` (no hard break) until `v1` is deprecated.
