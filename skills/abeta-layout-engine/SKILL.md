---
name: abeta-layout-engine
description: Develop, test, and operate the Abeta Dynamic SDUI Layout Engine (github.com/abeta-dev/layout-engine v0.3.0, Go 1.26.0 Clean Architecture) consuming github.com/abeta-dev/kmp/bff@v0.5.0 and github.com/abeta-dev/go-libs@v0.4.0. Use when modifying domain entities, rule/slotting engines, Studio authoring APIs, embedded Visual Studio SPA, SQLite/Postgres persistence, or CI/CD workflows.
---

# Abeta Dynamic Layout Engine (`abeta-dev/layout-engine`)

Use this skill whenever designing, implementing, debugging, or reviewing features in `github.com/abeta-dev/layout-engine`.

---

## 1. Clean Architecture Package Layout

```text
layout-engine/
├── cmd/layout-engine/       # Entrypoint, go-libs/env config, auto-migrate, starter seed, HTTP/2+h2c server, graceful shutdown
├── internal/
│   ├── domain/              # Pure entities, value objects, sentinel errors, persistence & engine ports + arch_test.go
│   ├── engine/              # Pure Go RuleEvaluator, ScheduleWindow matcher, SlottingEngine, PageComposer, LayoutValidator
│   ├── usecase/             # ResolveService, StudioService (draft/validate/publish/rollback/fallback/changelog), ActionService
│   ├── handler/             # Gin router, runtime_handler, studio_handler, embedded studio_assets SPA (/studio), OpenAPI 3.1 (/openapi.json, /docs), compression & ETag middleware
│   └── infra/               # Dual SQLite (modernc.org/sqlite) + PostgreSQL (pgx) Store, migrations runner, abeta-fresh starter seeder
├── migrations/
│   ├── sqlite/              # SQLite DDL migrations (0001_initial_schema.sql, embedded)
│   └── postgres/            # PostgreSQL DDL migrations (0001_initial_schema.sql, embedded)
├── scripts/                 # check_brand_hygiene.sh, verify_all.sh
├── docs/                    # architecture.md, studio-and-rules.md, kmp-integration.md
└── .github/workflows/       # ci.yml (hygiene, gitleaks CLI, SQLite + Postgres tests, Docker build), release.yml (cross-platform binaries + multi-arch GHCR)
```

### Dependency Rule Enforcement
- `internal/domain` and `internal/engine` MUST NEVER import `github.com/gin-gonic/gin`, `net/http`, `database/sql`, `modernc.org/sqlite`, `github.com/jackc/pgx`, `internal/infra`, or `internal/handler`.
- `internal/domain/arch_test.go` automatically fails `go test ./...` if any forbidden import is introduced.

---

## 2. Wire Contract & Companion Kit Rules

1. **Reuse `abeta-dev/kmp/bff` Types Directly**:
   - Import `github.com/abeta-dev/kmp/bff/pkg/sdui/{actions,atoms,snippets,theming}` (`v0.5.0`) rather than redefining wire structs.
   - Maintain full compatibility with all **21 KMP SnippetTypes** (`header_v1`, `banner_v1`, `metric_card`, `card_v1`, `action_row`, `form_v1`, `empty_state`, `divider`, `spacer`, `carousel`, `grid`, `tabs`, `accordion`, `product_card`, `search_bar`, `chip_group`, `swipe_deck`, `nudge_bar`, `marquee_strip`, `order_tracker`, `composed_block`), **28 Wire Atoms / 30 UI Atoms**, **29 `ComposedNodeKind`s** (`maxDepth <= 6`), and **18 ActionTypes**.
2. **Envelope & Drop-In KMP BFF Compatibility**:
   - Wrap all `/api/v1/*` JSON responses in `github.com/abeta-dev/go-libs/ginmw` envelopes (`ginmw.RespondSuccess` / `ginmw.RespondError`).
   - Preserve drop-in compatibility for `abeta-dev/kmp` clients pointed at `--bff-url=http://localhost:8090` (`GET /api/v1/screens`, `GET /api/v1/screens/:screenId`, `GET /api/v1/pages/:slug`, `GET /api/v1/config/splash`, `GET /api/v1/catalog/tokens`, `POST /api/v1/actions/execute`, `POST /api/v1/orders/express`).
   - Emit `X-Layout-Version` on resolved pages and `X-Layout-Fallback: true` whenever a fallback layout was served.

---

## 3. Authoring, Rules, Slotting & Governance Rules

- **Multi-Tenant Isolation**: Always scope templates, widgets, layouts, rules, slotting rules, snapshots, fallbacks, splash configs, token overrides, and changelogs by `tenant_id` (default tenant slug: `abeta-fresh`, selectable via `?tenant=` or `X-Tenant-ID`).
- **Draft vs Published Separation**:
  - Runtime endpoints (`GET /api/v1/screens/:screenId`, `GET /api/v1/pages/:slug`) resolve only immutable `PublishedSnapshot` records (with automatic failover to `FallbackLayout`).
  - Draft preview is restricted to `POST /api/v1/studio/layouts/:slug/preview` with `{"draft": true, "context": {...}}`.
- **Auditability**: Every Studio mutation records a `ChangeLogEntry` with actor (`X-Studio-Actor`, default `studio-admin`) and field-level `DiffItem` entries.
- **Deterministic Evaluation**: Rule evaluation (`RequestContext`) MUST be pure and deterministic given `reqCtx.Now` and `reqCtx.UserID` so Studio preview (`?at=<RFC3339>`) reproduces production output byte-for-byte.

---

## 4. Dual-Database & Starter Seed Guidelines

- **SQLite Default (Zero-Config)**: `DB_DIALECT=sqlite`, `DB_DSN=file:layout-engine.db?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)`.
- **PostgreSQL 17 Production**: `DB_DIALECT=postgres`, `DB_DSN=postgres://layout:layout@postgres:5432/layout_engine?sslmode=disable`.
- Keep both `migrations/sqlite/` and `migrations/postgres/` in sync and ensure 7-vertical starter seeding (`abeta-fresh`, `abeta-bites`, `abeta-pulse`, `abeta-mart`, `abeta-spark`, `abeta-trade`, `abeta-pay` — 21 templates per tenant) is idempotent.

---

## 5. Verification Commands

```bash
# Brand & machine-path hygiene
./scripts/check_brand_hygiene.sh

# Format, static analysis, race tests & binary compilation
test -z "$(gofmt -l .)"
go vet ./...
go test -count=1 -race -cover ./...
CGO_ENABLED=0 go build -trimpath -o /dev/null ./cmd/layout-engine

# Full verification pipeline
./scripts/verify_all.sh
```
