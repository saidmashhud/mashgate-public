# Contributing

Mashgate SDK is a polyglot repo. Each language has its own build + test flow.

## Repo structure

```
sdk/go/         Go module (go 1.22+)
sdk/typescript/ npm package (@mashgate/sdk)
sdk/python/     PyPI package (mashgate-sdk)
```

## Language-specific commands

### Go

```bash
cd sdk/go
go mod tidy
go build ./...
go vet ./...
go test ./...
```

### TypeScript

```bash
cd sdk/typescript
pnpm install
pnpm build
pnpm test
```

### Python

```bash
cd sdk/python
pip install -e ".[dev]"
pytest
```

## Contract sync

SDK consumes pinned contract snapshots from the Mashgate core monorepo. To refresh:

```bash
./contracts-sync/scripts/sync.sh
```

Updates `contracts-sync/generated/` from the snapshot tag in `contracts-sync/manifests/active.yaml`. Don't edit generated files by hand.

## Releases

Per-language tag triggers a publish workflow:

- Go: tag `sdk/go/v1.3.0`
- TS: tag `sdk/typescript/v1.3.0`
- Python: tag `sdk/python/v1.3.0`

Use `-beta.N` / `-rc.N` suffixes for pre-releases.

Each release **must** include:
- Updated language-specific `CHANGELOG.md`
- Updated compatibility matrix entry
- Migration notes (if breaking)
- Example code updated

## Breaking changes

Breaking changes require:
1. `breaking` label on the PR.
2. 90-day deprecation notice for existing consumers.
3. Major version bump.
4. Entry in migration guide.

## Do not

- Modify contracts in this repo. SoT is `mashgate/contracts/` in the core monorepo.
- Include HookLine SDK code. That's a separate product (see [`github.com/saidmashhud/hookline`](https://github.com/saidmashhud/hookline)).
- Mix languages in a single PR. Go changes → Go PR. TS changes → TS PR.
- Skip tests (`go test`, `pnpm test`, `pytest` must pass in CI).
