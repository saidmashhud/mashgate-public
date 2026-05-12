# Mashgate Service Catalog

> Источник правды о сервисах Mashgate платформы. Обновляется при добавлении/удалении сервисов или версионных бампах protos.

Auto-generated baseline from k8s state + `contracts/proto/v1/*.proto` snapshot.
Last regenerated: **2026-05-12** | Cluster: `srv2 (95.142.87.230)` | k3s v1.34.5

---

## TL;DR

| # | Category | Count | Purpose |
|---|---|---|---|
| 1 | [Core Platform](#1-core-platform) | 5 | Identity, money, accounts |
| 2 | [Risk & Compliance](#2-risk--compliance) | 3 | Fraud, guard, KYC/compliance |
| 3 | [Communication](#3-communication) | 3 | Notify, mail, chat |
| 4 | [Crypto Layer (mgChain)](#4-crypto-layer-mgchain) | 3 | RPC, indexer, orchestrator |
| 5 | [Analytics & Events](#5-analytics--events) | 5 | Analytics, mg-events, event-stream, ch-writer, logs |
| 6 | [Platform Plumbing](#6-platform-plumbing) | 8 | Control-plane, flags, metering, orchestrator, outbox, webhook, ext-authz, hookline |
| 7 | [Infrastructure](#7-infrastructure) | 9 | envoy, postgres, kafka, dragonfly, clickhouse, minio, frontend, grafana, prometheus, waypoint |

**Total Mashgate deployments**: 41 (32 business + 9 infra)
**Total gRPC services**: 33 (in `contracts/proto/v1/*.proto`)
**Total RPCs**: ~350+

---

## 1. Core Platform

### auth-service
- **Deployment**: `mashgate/auth-service` × 2 replicas
- **Image**: `mg/auth-service:0.1.13`
- **Container port**: `50052` (gRPC)
- **Service**: `auth-service.mashgate.svc:50052` (grpc)
- **Proto**: `contracts/proto/v1/auth.proto` (26 RPCs)
- **gRPC service**: `AuthService`
- **Key RPCs**: Register, Login, RefreshToken, Logout, ValidateToken, CreateSession, GetSession, DeleteSession, GetUserProfile, UpdateUserProfile, …
- **Companion (proto/v1)**: `iam.proto` (`IamService`, 48 RPCs) — RBAC/ABAC permissions
- **Companion (proto/v1)**: `oidc.proto` (`OidcService`, 11 RPCs) — OIDC handshake
- **Consumers**: qrapp, vint, zist, grid, kiro, mail (JWT verify), любой backend через JWKS

### billing-service
- **Deployment**: `mashgate/billing-service` × 2
- **Image**: `mg/billing-service:latest`
- **Port**: `50070` (gRPC)
- **Proto**: `contracts/proto/v1/billing.proto` (29 RPCs)
- **gRPC service**: `BillingService`
- **Key RPCs**: ListPlatformPlans, GetTenantSubscription, ChangePlan, CancelPlan, ListPaymentMethods, AddPaymentMethod, …
- **Companion**: `subscriptions.proto` (`SubscriptionService`, 11 RPCs)
- **Consumers**: tenants (subscription mgmt), qrapp (renew flow)

### payments-orchestrator
- **Deployment**: `mashgate/payments-orchestrator` × 2 + legacy alias `orchestrator` × 1
- **Image**: `mg/payments-orchestrator:latest`
- **Port**: `9090` (http internal)
- **Proto**: `contracts/proto/v1/service.proto` (`PaymentsService`, 23 RPCs), `local_payments.proto` (`LocalPaymentsService`, 9 RPCs)
- **Key RPCs**: Authorize, Capture, Void, Refund, CreatePaymentIntent, …
- **Consumers**: qrapp/checkout, vint/orders, zist/payments
- **Status**: ⚠️ External provider **not provisioned** (живёт в mock); см. ADR-0009

### checkout-service
- **Deployment**: `mashgate/checkout-service` × 2
- **Image**: `mg/checkout-service:0.4.1`
- **Port**: `50056` (gRPC)
- **Proto**: `checkout.proto` (`CheckoutService`, 4 RPCs)
- **Key RPCs**: CreateCheckoutSession, GetCheckoutSession, ExpireCheckoutSession, CompleteCheckoutSession
- **Consumers**: qrapp (activation flow), vint/zist (order checkout)

### card-processor
- **Deployment**: `mashgate/card-processor` × 2
- **Image**: `mg/card-processor:latest`
- **Port**: `50055` (gRPC, container) / `50062` (svc)
- **Proto**: `card_processor.proto` (`CardProcessorService`, 4 RPCs)
- **ADR**: 0003-internal-card-processor
- **Status**: Internal-only, не consumed напрямую вертикалями (используется payments-orchestrator)

---

## 2. Risk & Compliance

### fraud-service
- **Deployment**: `mashgate/fraud-service` × 2
- **Image**: `mg/fraud-service:latest`
- **Ports**: `50058` (gRPC), `8080` (http internal)
- **Proto**: `risk.proto` (`RiskService`, 9 RPCs)
- **Key RPCs**: AssessTransaction, ListAssessments, GetAssessment, …
- **Consumers**: payments-orchestrator (pre-auth risk score)

### guard-service
- **Deployment**: `mashgate/guard-service` × 1
- **Image**: `mg/guard-service:latest`
- **Port**: `50054` (gRPC) — конфликт нумерации с `event-stream`, см. **Gaps**
- **Proto**: `guard.proto` (`GuardService`, 7 RPCs)
- **Purpose**: Trust signals / 3DS-like flow

### compliance (proto only — service?)
- **Proto**: `compliance.proto` (`ComplianceService`, 8 RPCs), `kyc.proto` (`KycService`, 4 RPCs)
- **Status**: Proto есть; конкретный deployment не выделен — функция, возможно, внутри auth-service или платформы

---

## 3. Communication

### notify-service
- **Deployment**: `mashgate/notify-service` × 1
- **Image**: `mg/notify-service:latest`
- **Port**: `8080` (http)
- **Proto**: `notify.proto` (`NotifyService`, 7 RPCs)
- **Channels**: email, in-app
- **Missing**: **Telegram, SMS** (см. Gaps → план E)

### mail-service
- **Deployment**: `mashgate/mail-service` × 1
- **Image**: `mg/mail-service:0.4.3`
- **Port**: `50080` (gRPC)
- **Proto**: `mail.proto` (`MailService`, 12 RPCs)
- **Domain**: транзакционные письма (SendGrid/SES, см. ADR/runbook)

### chat-service
- **Deployment**: `mashgate/chat-service` × 1
- **Image**: `mg/chat-service:latest`
- **Port**: `50051` (gRPC)
- **Proto**: `chat.proto` (`ChatService`, 6 RPCs)
- **Consumers**: qrapp messages, vint support
- **Companion**: storage-service (attachments)

---

## 4. Crypto Layer (mgChain)

### chain-rpc
- **Deployment**: `mashgate/chain-rpc` × 1
- **Image**: `mg/chain-rpc:latest`
- **Port**: `50074` (gRPC)
- **Proto**: `chain.proto` (`ChainService`, 21 RPCs), `chain_internal.proto` (`ChainRpcService`, 11 RPCs)
- **Language**: Rust (per memory project_mgchain.md)
- **Purpose**: Read/write node-facing операции

### chain-indexer
- **Deployment**: `mashgate/chain-indexer` × 1
- **Image**: `mg/chain-indexer:latest`
- **Language**: Rust
- **Purpose**: индексирует transactions/blocks → ClickHouse

### mgchain-orchestrator
- **Deployment**: `mashgate/mgchain-orchestrator` × 1
- **Image**: `mg/mgchain-orchestrator:latest`
- **Port**: `50070` (gRPC) — pop конфликт с `billing-service` namespaced-port, OK на gRPC уровне (разные svc)
- **Language**: Scala/ZIO (per memory)
- **Purpose**: Координация chain-rpc + chain-indexer; deposit/withdraw flow

### wallet (proto only)
- **Proto**: `wallet.proto` (`WalletService`, 12 RPCs)
- **Implementation**: предположительно внутри mgchain-orchestrator

---

## 5. Analytics & Events

### analytics-service
- **Deployment**: `mashgate/analytics-service` × 2
- **Image**: `mg/analytics-service:latest`
- **Port**: `50072` (gRPC)
- **Proto**: `analytics.proto` (`AnalyticsService`, 12 RPCs), `reports.proto` (`ReportsService`, 9 RPCs)
- **Storage**: ClickHouse

### mg-events
- **Deployment**: `mashgate/mg-events` × 1
- **Image**: `mg/mg-events:latest`
- **Port**: `50059` (gRPC)
- **Proto**: `mg_events.proto` (`MgEventsService`, 16 RPCs), `events.proto`
- **Language**: Elixir (BEAM) — per `beam.smp` в top processes
- **Purpose**: Event bus / outbox publisher

### event-stream
- **Deployment**: `mashgate/event-stream` × 2
- **Image**: `mg/event-stream:latest`
- **Port**: `50054` (gRPC) — конфликт с `guard-service` container port (разные svc DNS, OK)
- **Proto**: `stream.proto` (`EventStreamService`, 3 RPCs)

### ch-writer
- **Deployment**: `mashgate/ch-writer` × 1
- **Image**: `mg/ch-writer:latest`
- **Purpose**: Batch writer Kafka → ClickHouse

### logs-service
- **Deployment**: `mashgate/logs-service` × 1
- **Image**: `mg/logs-service:latest`
- **Port**: `50055` (gRPC) — конфликт с `card-processor` container port (svc DNS разные, OK)
- **Proto**: `logs.proto` (`LogService`, 6 RPCs)

---

## 6. Platform Plumbing

### control-plane
- **Deployment**: `mashgate/control-plane` × 1
- **Image**: `mg/control-plane:v1`
- **Port**: `80` (http)
- **Purpose**: Tenant management (creation, updates, capabilities)

### platform-service
- **Deployment**: `mashgate/platform-service` × 1
- **Image**: `mg/platform-service:v3`
- **Port**: `8082`
- **Proto**: `platform.proto` (`PlatformService`, 15 RPCs)

### flags-service
- **Deployment**: `mashgate/flags-service` × 1
- **Image**: `mg/flags-service:latest`
- **Port**: `50050` (gRPC)
- **Proto**: `feature_flags.proto` (`FeatureFlagService`, 6 RPCs)

### metering-service
- **Deployment**: `mashgate/metering-service` × 2
- **Image**: `mg/metering-service:latest`
- **Port**: `50071` (gRPC)
- **Proto**: `metering.proto` (`MeteringService`, 3 RPCs)
- **Purpose**: Usage tracking (billing input)

### outbox-relay
- **Deployment**: `mashgate/outbox-relay` × 1
- **Image**: `mg/outbox-relay:0.4.0`
- **Purpose**: Publishes domain events from postgres outbox → Kafka

### webhook-delivery
- **Deployment**: `mashgate/webhook-delivery` × 2
- **Image**: `mg/webhook-delivery:0.4.1`
- **Proto**: `webhooks.proto` (`WebhookService`, 9 RPCs)
- **Consumers**: vint, zist (receive webhook events)

### ext-authz
- **Deployment**: `mashgate/ext-authz` × 2
- **Image**: `mg/ext-authz:0.1.1`
- **Port**: `50060` (gRPC)
- **Purpose**: Envoy `ext_authz` filter — централизованная авторизация перед бизнес-сервисами
- **ADR**: 0004-iam-rbac-abac

### hookline
- **Deployment**: `mashgate/hookline` × 1
- **Image**: `hookline:0.2.0`
- **Port**: `8080`
- **ADR**: 0011-hookline-embed, 0013-hookline-boundary
- **Purpose**: Embed widget (checkout iframe / signup form), интегрируется на любой сайт

### invoice-service
- **Deployment**: `mashgate/invoice-service` × 1
- **Image**: `mg/invoice-service:latest`
- **Port**: `50076` (gRPC)
- **Proto**: `invoices.proto` (`InvoiceService`, 8 RPCs)

### storage-service
- **Deployment**: `mashgate/storage-service` × 1
- **Image**: `mg/storage-service:latest`
- **Port**: `50053` (gRPC)
- **Proto**: `storage.proto` (`StorageService`, 8 RPCs)
- **Key RPCs**: GetUploadUrl, GetDownloadUrl, GetObject, DeleteObject, ListBuckets
- **Backend**: MinIO (`mashgate/minio`)
- **Consumers**: qrapp (photos, avatar, chat attachments), любой вертикал через SDK

### subscription-service
- **Deployment**: `mashgate/subscription-service` × 1
- **Image**: `mg/subscription-service:latest`
- **Port**: `50077` (gRPC)
- **Proto**: `subscriptions.proto`

### ledger-core
- **Deployment**: `mashgate/ledger-core` × 2
- **Image**: `mg/ledger-core:0.4.0`
- **Ports**: `50051` (gRPC) `9663` (metrics?)
- **Purpose**: Double-entry ledger, исходный SSOT для money movement

---

## 7. Infrastructure

| Сервис | Image | Назначение |
|---|---|---|
| `envoy` × 2 | envoyproxy/envoy-contrib:v1.29-latest | Service mesh gateway (REST/gRPC/WebSocket fronting) |
| `postgres` × 1 | postgres:15-alpine | Primary OLTP DB |
| `kafka` × 1 | apache/kafka:3.7.0 | Event bus |
| `dragonfly` × 1 | dragonflydb/dragonfly:latest | Redis-compat cache |
| `clickhouse` × 1 | clickhouse:23.8-alpine | Analytics OLAP |
| `minio` × 1 | minio/minio:latest | S3-compat object storage (Storage backend) |
| `frontend` × 2 | mg/frontend:redesign-v3 | Admin/console SPA (`test.entry-i.com`) |
| `grafana` × 1 | grafana/grafana:11.3.0 | Observability UI (deployed 2026-05-12) |
| `prometheus` × 1 | prom/prometheus:v3.0.0 | Metrics (deployed 2026-05-12) |
| `waypoint` × 1 | istio/proxyv2:1.29.1-distroless | Istio Ambient L7 waypoint |

---

## Verticals — How They Use Mashgate

| Вертикал | SDK | Capabilities | Status |
|---|---|---|---|
| **qrapp** | Go SDK (full) | Storage, Payments, Auth (HS256+JWKS), Hookline | ✅ Production |
| **vint** | Raw HTTP (no Go SDK) | mgID (Auth), mgPay (Payments), Webhooks | ⚠️ Awaiting Go SDK |
| **zist** | Raw HTTP (no Go SDK) | mgID, mgPay, Webhooks | ⚠️ Awaiting Go SDK |
| **grid** | Direct (env: URL+API_KEY+TENANT) | Flags, Notify, Storage, ext-authz | ✅ |
| **kiro** | JWT verify only | (consumer of tokens) | ✅ |
| **mail** | OIDC widget | Auth (OIDC redirect) | 🟡 Pre-mailcore |
| casino | — | (intentionally isolated, regulatory) | N/A |
| lana | — | (standalone crawler) | N/A |

---

## Gaps & Conventions

### Известные несоответствия

1. **Port naming**: Некоторые services используют разные gRPC порты для container vs service mapping. Это OK для DNS, но запутывает чтение `kubectl get all`. Стандартизировать (`grpc=50050+offset`).
2. **`mg/*:latest` tags**: Большинство деплойментов на `latest` — нет фиксации версий, нельзя откатить точно. Перейти на semver tags (как auth-service 0.1.13, ledger-core 0.4.0, checkout-service 0.4.1).
3. **Replicas=1 для критичных**: postgres, kafka, dragonfly, clickhouse, minio, control-plane, hookline, mgchain-orchestrator. Перезапуск = downtime. Добавить PodDisruptionBudget хотя бы.

### Чего не хватает (из roadmap)

| Pri | Gap | Влияние |
|---|---|---|
| P1 | **Go SDK codegen** (TS уже есть) | vint/zist строят raw HTTP клиенты, contract drift |
| P1 | **Payments external provider** | Все checkout flows в mock |
| P1 | **SMS provider wire** | Реальный OTP login не работает |
| P2 | **Telegram channel в notify-service** | qrapp+casino дублируют Telegram-клиент |
| P2 | **Webhook self-service subscriptions** | Подписки настраиваются вручную через config |
| P2 | **Service-level Prometheus metrics** | Сейчас видим только envoy aggregate, не per-service internals |
| P2 | **mTLS / AuthorizationPolicy конвенция** | Istio Ambient + ZTUN: bezopaen, но onboarding нового сервиса — не задокументирован |
| P3 | **Multi-tenant federation** | `mashgate-1` ns как клон, а не tenant-as-data |

---

## ADR Index

`/opt/mashgate/docs/adr/`:

- [0001 Monorepo](../../mashgate/docs/adr/0001-monorepo.md)
- [0002 Architecture guardrails](../../mashgate/docs/adr/0002-architecture-guardrails.md)
- [0003 Internal card processor](../../mashgate/docs/adr/0003-internal-card-processor.md)
- [0004 IAM RBAC/ABAC](../../mashgate/docs/adr/0004-iam-rbac-abac.md)
- [0005 Frontend MFE contract](../../mashgate/docs/adr/0005-frontend-mfe-contract.md)
- [0006 API keys & rate limits](../../mashgate/docs/adr/0006-api-keys-and-rate-limits.md)
- [0007 Consumer side-effects contract](../../mashgate/docs/adr/0007-consumer-side-effects-contract.md)
- [0008 Backend structure](../../mashgate/docs/adr/0008-backend-structure.md)
- [0009 Local payment methods](../../mashgate/docs/adr/0009-local-payment-methods.md)
- [0010 Permissions model](../../mashgate/docs/adr/0010-permissions-model.md)
- [0011 Hookline embed](../../mashgate/docs/adr/0011-hookline-embed.md)
- [0012 Platform constitution](../../mashgate/docs/adr/0012-platform-constitution.md)
- [0013 Hookline boundary](../../mashgate/docs/adr/0013-hookline-boundary.md)
- [0014 SDK product boundary](../../mashgate/docs/adr/0014-sdk-product-boundary.md)
- [0015 Pack is config, not code](../../mashgate/docs/adr/0015-pack-is-config-not-code.md)
- [0016 mgcrypto BaaS surface](../../mashgate/docs/adr/0016-mgcrypto-baas-surface.md)
- [0017 SLIP-10 derivation](../../mashgate/docs/adr/0017-slip10-derivation.md)
- *(ADR-0020 Tenant SoT track — нет в /docs/adr/, в memory упоминается phase A+B done)*

---

## Adding a New Service — Skeleton Checklist

1. **Proto contract**: новый `contracts/proto/v1/<name>.proto` с `service XService { rpc ... }`
2. **Codegen**: `make sdk-gen-ts` (Go SDK pending — план C)
3. **Implementation**: язык по выбору; обычно Scala/ZIO или Go
4. **k8s manifest**: Deployment + Service в `mashgate` ns
   - Container port в range `50050-50099` (gRPC)
   - Standard labels: `app: <name>`, `app.kubernetes.io/component: ...`
   - Resource requests/limits (минимум 50m CPU / 128Mi RAM)
5. **Istio Ambient policy**: AuthorizationPolicy если требуются специфические caller permissions (default — внутренний trafic из mashgate/istio-system/ingress-nginx/kube-system/grid через `allow-internal`)
6. **Image registry**: `mg/<name>:<semver>` (НЕ `:latest`!), pushed в `registry.mashgate.svc:5000`
7. **Update этого catalog** + добавить ADR если архитектурное решение

---

*Каталог — живой документ. PR-ы welcome. Auto-validate: `make catalog-check` (планируется).*
