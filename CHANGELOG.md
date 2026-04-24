# Changelog

All notable changes to Mashgate SDK are documented here.
Format: [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
Versioning: per-language [SemVer](https://semver.org/spec/v2.0.0.html).

Per-language changelogs under `sdk/{go,typescript,python}/CHANGELOG.md`.
Top-level entry is the aggregate snapshot.

---

## [Unreleased]

### Added
- Initial repo extraction from `github.com/saidmashhud/mashgate` monorepo.
- Go SDK: migrated `mashgate/sdk/go/` → `mashgate-public/sdk/go/`. Module path is now `github.com/saidmashhud/mashgate-public/sdk/go`.
- Go SDK: new **Fintech Pack** subpackage under `sdk/go/fintech/` — types + clients for `kyc`, `compliance`, `merchant`, `wallet` services. Sourced from Kiro's hand-rolled client to avoid divergence.
- TypeScript SDK: migrated `mashgate/sdk/typescript/` → `mashgate-public/sdk/typescript/`. npm package `@mashgate/sdk`.
- Python SDK: migrated `mashgate/sdk/python/` → `mashgate-public/sdk/python/`. PyPI package `mashgate-sdk`.
- Repository scaffolding: README, LICENSE (Apache 2.0), CONTRIBUTING, ROADMAP, docs/, examples/, contracts-sync/, tests/, tooling/, .github/ workflows placeholders.

### Changed
- All Go imports rewritten to `github.com/saidmashhud/mashgate-public/sdk/go`.

### Planned — v1.0.0
- Coordinated release (`sdk/go/v1.0.0`, `sdk/typescript/v1.0.0`, `sdk/python/v1.0.0`).
- Compatibility matrix published.
- Migration guide for existing consumers (in-tree SDK, Kiro hand-rolled).
- Compatibility bridge in core monorepo `mashgate/sdk/` re-exporting from public packages with deprecation notices; grace window ≥6 months.
