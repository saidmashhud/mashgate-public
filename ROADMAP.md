# Mashgate SDK Roadmap

Status: `planned` / `in-progress` / `shipped`.

---

## v0.x — extraction + Fintech Pack (current)

- ✅ Repo bootstrapped from Mashgate monorepo SDK.
- ✅ Go + TypeScript + Python SDKs relocated, module paths updated.
- ✅ Go `sdk/go/fintech` subpackage (KYC / compliance / merchant / wallet).
- ⏳ TypeScript `@mashgate/sdk/fintech` — port types from `@kiro/mashgate-types`.
- ⏳ Python `mashgate.fintech` — port types + thin client.
- ⏳ Contract-sync pipeline (pinned-snapshot mode).
- ⏳ CI: lint + build + test per language.

## v1.0.0 — coordinated release (Mashgate Wave 4)

- [ ] `sdk/go/v1.0.0`, `sdk/typescript/v1.0.0`, `sdk/python/v1.0.0` published.
- [ ] Compatibility matrix doc published.
- [ ] Migration guides: from in-tree SDK, from Kiro hand-rolled, generic for Zist/Vint.
- [ ] Compatibility bridge in Mashgate core monorepo with 6-month deprecation window.
- [ ] ADR-0014 (SDK Product Boundary) landed in core.
- [ ] Examples for every module per language, CI-verified.

## v1.x — hardening (Mashgate Wave 5)

- [ ] Breaking-change protection CI (api-diff).
- [ ] Integration test matrix (SDK × platform version).
- [ ] Docs site (GitHub Pages).
- [ ] OTel integration helpers per language.
- [ ] Rate-limit / retry helpers harmonised across languages.

## v2.x — envelope v2 migration (Mashgate post-Wave 5)

- [ ] Envelope v2 support added.
- [ ] v1 envelope deprecated with 90-day notice.
- [ ] Major version bump in all three SDKs.

---

## Out of scope

- HookLine SDK — lives in `github.com/saidmashhud/hookline` (separate product, ADR-0013 in core).
- `mashgatectl` CLI — separate product surface, stays in Mashgate core monorepo.
- Proto contracts SoT — stays in `mashgate/contracts/`. This repo consumes pinned snapshots.
