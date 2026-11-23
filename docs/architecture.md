# Architecture — Abeta Dynamic Layout Engine

**Abeta Dynamic Layout Engine (`github.com/abeta-dev/layout-engine`, `v0.3.0`)** is a Go `1.26.0` Server-Driven UI (SDUI) control plane and runtime page resolver designed around **Clean Architecture**, the **Uncle Bob Dependency Rule**, and **1:1 wire compatibility** with [`abeta-dev/kmp`](https://github.com/Abeta-dev/kmp) (`github.com/abeta-dev/kmp/bff@v0.5.0`) and [`abeta-dev/go-libs`](https://github.com/Abeta-dev/go-libs) (`v0.4.0`).

---

## 1. Clean Architecture Layering

Dependencies point strictly **inward** toward pure domain entities and interfaces. Infrastructure drivers, HTTP frameworks, and transport middleware are outermost adapters.

```mermaid
flowchart LR
    CMD["cmd/layout-engine\n(Bootstrap, Config, HTTP/2 & h2c, Shutdown)"]
    HANDLER["internal/handler\n(Gin Router, Studio SPA, OpenAPI/Scalar,\nCompression, ETag, go-libs/ginmw)"]
    USECASE["internal/usecase\n(ResolveService, StudioService, ActionService)"]
    ENGINE["internal/engine\n(RuleEvaluator, SlottingEngine,\nPageComposer, LayoutValidator, DiffEngine)"]
    DOMAIN["internal/domain\n(Pure Entities, Ports, Errors,\narch_test.go)"]
    INFRA["internal/infra\n(SQLite + Postgres SQLStore,\nMigrations, 7-Vertical Seeder)"]

    CMD --> HANDLER
    CMD --> USECASE
    CMD --> INFRA
    HANDLER --> USECASE
    USECASE --> ENGINE
    USECASE --> DOMAIN
    ENGINE --> DOMAIN
    INFRA -. "implements domain.Store" .-> DOMAIN
```

### Automated Boundary Guard (`internal/domain/arch_test.go`)

`internal/domain/arch_test.go` parses AST imports across `internal/domain` and `internal/engine` during every `go test ./...` run and fails the build if any non-test file imports:
- `github.com/gin-gonic/gin`
- `net/http`
- `database/sql`
- `modernc.org/sqlite`
- `github.com/jackc/pgx`
- `internal/infra`
- `internal/handler`

---

## 2. Package Responsibilities

| Package | Responsibility |
| :--- | :--- |
| `cmd/layout-engine` | Application entrypoint loading env config via `go-libs/env`, opening SQLite or PostgreSQL, running auto-migrations and idempotent 7-vertical starter seeding (`abeta-fresh`, `abeta-bites`, `abeta-pulse`, `abeta-mart`, `abeta-spark`, `abeta-trade`, `abeta-pay`), serving HTTP/1.1 + `h2c` + TLS HTTP/2, and coordinating graceful shutdown via `go-libs/shutdown`. |
| `internal/domain` | Pure business models (`Tenant`, `WidgetTemplate`, `Widget`, `TargetingRule`, `ScheduleWindow`, `Layout`, `LayoutSlot`, `LayoutPageConfig`, `SlottingRule`, `PublishedSnapshot`, `FallbackLayout`, `SplashConfig`, `TokenOverride`, `ChangeLogEntry`, `DiffItem`, `RequestContext`, `ResolutionTrace`, `ResolvedPageResult`, `ValidationIssue`), sentinel errors, and all storage/engine interfaces (`Store`, `RuleEvaluator`, `SlottingEngine`, `PageComposer`, `LayoutValidator`). |
| `internal/engine` | Pure deterministic domain engines: `DefaultRuleEvaluator` (13 operators + UTC `ScheduleWindow`), `DefaultSlottingEngine` (`insert`/`prepend`/`append`/`replace` with priority & pinned-slot protection), `DefaultLayoutValidator` (all 21 KMP snippet types, 29 `ComposedNodeKind`s with `maxDepth <= 6`, and recursive container cycle detection), and JSON `DiffEngine`. |
| `internal/usecase` | Application orchestration: `ResolveService` (published snapshot resolution, Studio draft preview, splash & token resolution, fallback failover), `StudioService` (CRUD, validation, immutable snapshot publishing, historical rollback, fallback promotion, audit diff recording), and `ActionService` (all 18 KMP `v1` action types + express order follow-ups). |
| `internal/handler` | Gin HTTP delivery layer: runtime KMP-compatible endpoints, Studio authoring/governance API, embedded zero-build Visual Layout Studio SPA (`internal/handler/studio_assets/` with 21 snippet renderers & 29-kind Atomic Snippet Molder), OpenAPI 3.1 + Scalar docs (`/openapi.json`, `/docs`), `/llms.txt` & `/llms-full.txt`, Brotli/gzip compression, SHA-256 `ETag`/`304`, and `go-libs/ginmw` middlewares. |
| `internal/infra/persistence` | Unified `SQLStore` implementing `domain.Store` across pure-Go SQLite (`modernc.org/sqlite`) and PostgreSQL (`github.com/jackc/pgx/v5/stdlib`), plus embedded migration runner (`migrations/{sqlite,postgres}`) and 7-vertical seeder. |

---

## 3. Core Domain Entities & Persistence Ports

`domain.Store` composes 10 cohesive persistence ports behind a single closeable adapter:

| Port Interface | Core Entity | Purpose |
| :--- | :--- | :--- |
| `TenantStore` | `Tenant` | Isolates workspaces per brand/app (`slug`, `active`). Default workspace: `abeta-fresh`. |
| `TemplateStore` | `WidgetTemplate` | Reusable blueprints mapping 1:1 to all 21 KMP `v1` `SnippetTypes` (28 wire atoms, 29 `ComposedNodeKind`s). |
| `WidgetStore` | `Widget` | Configured reusable UI/commerce/container components with `props`, `tracking`, and `motion`. |
| `LayoutStore` | `Layout` | Editable draft page definitions holding `LayoutPageConfig` (toolbar, bottom bar, floating cart, adaptive split-pane, pagination, pull-to-refresh) and ordered `[]LayoutSlot`. |
| `RuleStore` | `TargetingRule` | Reusable targeting predicates (`match_mode`, `conditions`, optional `ScheduleWindow`, `priority`). |
| `SlottingStore` | `SlottingRule` | Dynamic slot-injection overrides scoped to a `layout_slug`. |
| `SnapshotStore` | `PublishedSnapshot` | Immutable versioned snapshots storing frozen `PageConfig`, `Slots`, and `ResolvedWidgets`. |
| `FallbackStore` | `FallbackLayout` | Pre-validated `snippets.PageResponse` snapshots served automatically on runtime error. |
| `ConfigStore` | `SplashConfig`, `TokenOverride` | Rule- and schedule-targeted launch splash screens and design token overrides. |
| `ChangeLogStore` | `ChangeLogEntry` | Immutable audit log capturing actor, summary, and field-level `[]DiffItem` JSON diffs. |

---

## 4. Runtime Page Resolution Pipeline

Every runtime request to `GET /api/v1/screens/:screenId` or `GET /api/v1/pages/:slug` executes a deterministic pipeline:

```mermaid
flowchart TD
    REQ["HTTP Request\n(Headers & Query Params -> RequestContext)"]
    TENANT["1. Resolve Tenant\n(?tenant= or X-Tenant-ID; default abeta-fresh)"]
    SNAP["2. Load Latest Immutable PublishedSnapshot\n(or Draft when called from Studio Preview)"]
    SNAP_OK{"Snapshot\nloaded?"}
    FALLBACK["Automatic Fallback Chain\nServe pre-validated FallbackLayout\nSet X-Layout-Fallback: true"]
    SLOT_FILTER["3. Evaluate Base Slot Rules & UTC ScheduleWindows\nRecord matched_rule_ids & skipped_slot_ids"]
    SLOTTING["4. Apply Dynamic SlottingEngine\nPriority DESC -> insert/prepend/append/replace\nRespect Pinned slots & deduplicate WidgetIDs"]
    COMPOSE["5. Compose v1 KMP PageResponse\nResolve widgets, container children & overrides"]
    COMP_OK{"Composition\nsucceeded?"}
    RESP["6. Emit Response\nginmw.Envelope + X-Layout-Version\nOptional ResolutionTrace (?include_trace=true)\nSHA-256 ETag + Brotli/Gzip"]

    REQ --> TENANT --> SNAP --> SNAP_OK
    SNAP_OK -- "Error / Missing" --> FALLBACK --> RESP
    SNAP_OK -- "Found" --> SLOT_FILTER --> SLOTTING --> COMPOSE --> COMP_OK
    COMP_OK -- "Error" --> FALLBACK
    COMP_OK -- "OK" --> RESP
```

### Runtime Response Headers
- `X-Layout-Version`: Integer version of the `PublishedSnapshot` (or fallback snapshot) rendered for this request.
- `X-Layout-Fallback: true`: Present whenever the automatic fallback chain served a `FallbackLayout`.
- `ETag` & `Cache-Control`: SHA-256 response validators supporting `304 Not Modified` when `ETAG_ENABLED=true`.
- `Content-Encoding` (`br` / `gzip`) & `Vary: Accept-Encoding`: Negotiated compression when `COMPRESSION_ENABLED=true`.

---

## 5. Dual-Database Persistence Architecture

`internal/infra/persistence.SQLStore` runs identically against both supported databases with zero CGO requirement:

| Mode | Dialect | Driver | Typical Configuration |
| :--- | :--- | :--- | :--- |
| **Local / Edge / CI Default** | `sqlite` | `modernc.org/sqlite` (pure Go) | `DB_DIALECT=sqlite`, `DB_DSN=file:layout-engine.db?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)` |
| **Production / Docker Compose** | `postgres` | `github.com/jackc/pgx/v5/stdlib` | `DB_DIALECT=postgres`, `DB_DSN=postgres://layout:layout@postgres:5432/layout_engine?sslmode=disable` |

- **Schema Parity**: Embedded SQL migrations live under `migrations/sqlite/0001_initial_schema.sql` and `migrations/postgres/0001_initial_schema.sql`.
- **Chronological Precision**: SQLite stores fixed-width UTC RFC3339Nano timestamps (`2006-01-02T15:04:05.000000000Z`) so lexical ordering matches exact nanosecond chronology.
- **Starter Seeding**: When `DB_AUTO_MIGRATE=true` and `SEED_DEFAULT_DATA=true`, first startup idempotently provisions all **7 multi-tenant vertical workspaces** (`abeta-fresh`, `abeta-bites`, `abeta-pulse`, `abeta-mart`, `abeta-spark`, `abeta-trade`, `abeta-pay`), each with all 21 canonical KMP `v1` templates, curated widgets, rules, slotting overrides, immutable `v1` snapshots, and hydrated fallbacks.

---

## 6. Transport, Security & Observability

- **Protocols**: Configured via `http.Server.Protocols` to accept `HTTP/1.1`, cleartext HTTP/2 (`h2c`), and TLS HTTP/2 when `TLS_CERT_FILE` and `TLS_KEY_FILE` are provided.
- **Probes & Telemetry**: `GET /healthz` (liveness), `GET /readyz` (database ping readiness), and `GET /metrics` (Gin metrics tracker via `go-libs/ginmw`).
- **API Documentation**: Live OpenAPI 3.1 JSON at `GET /openapi.json` and interactive Scalar reference UI at `GET /docs`.
- **Optional Production Guards**: `RATE_LIMIT_ENABLED=true` enables `ginmw.GlobalRateLimit()`, and `AUTH_REQUIRED=true` enables fail-closed `ginmw.AuthMiddleware()` (startup fails immediately if `JWT_SECRET` is empty). See [../.env.example](../.env.example) and [../docker-compose.yaml](../docker-compose.yaml).

---

## Related Guides

- [Visual Layout Studio, Rules & Slotting](studio-and-rules.md)
- [`abeta-dev/kmp` Client Integration & Wire Protocol](kmp-integration.md)
- [Repository Overview](../README.md) · [AI Agent Contract](../AGENTS.md)
