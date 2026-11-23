---
name: layout-engine-owner
module: "layout-engine"
layer: "Go 1.26.0 Clean Architecture Dynamic SDUI Layout Engine & Visual Layout Studio"
owned_paths:
  - "cmd/**"
  - "internal/**"
  - "pkg/**"
  - "migrations/**"
  - "docs/**"
  - "scripts/**"
  - ".github/**"
  - "Dockerfile"
  - "docker-compose.yaml"
  - "Makefile"
allowed_dependencies:
  - "cmd/layout-engine -> internal/handler -> internal/usecase -> internal/engine -> internal/domain <- internal/infra"
  - "github.com/abeta-dev/kmp/bff@v0.5.0 (pkg/sdui/{actions,atoms,snippets,theming})"
  - "github.com/abeta-dev/go-libs@v0.4.0 (ginmw, apperror, health, logger, env, shutdown)"
  - "github.com/gin-gonic/gin, github.com/andybalholm/brotli, modernc.org/sqlite, github.com/jackc/pgx/v5"
forbidden_dependencies:
  - "internal/domain or internal/engine importing gin, net/http, database/sql, modernc.org/sqlite, pgx, internal/infra, or internal/handler (enforced by internal/domain/arch_test.go)"
  - "Hardcoded secrets, proprietary legacy brand names, or machine-specific absolute paths"
required_skills:
  - "skills/abeta-layout-engine/SKILL.md"
verification_command: "./scripts/verify_all.sh"
---

# Code Owner Agent: Abeta Dynamic Layout Engine (`layout-engine-owner`)

## Ownership Scope
- **Repository**: `github.com/abeta-dev/layout-engine` (`Abeta-dev/layout-engine`, Apache-2.0)
- **Core Responsibilities**:
  - Multi-tenant Dynamic SDUI Layout Engine (7 pre-seeded vertical workspaces: `abeta-fresh`, `abeta-bites`, `abeta-pulse`, `abeta-mart`, `abeta-spark`, `abeta-trade`, `abeta-pay`, each with 21 KMP snippet templates, 28 wire atoms / 30 UI atoms, 29 `ComposedNodeKind`s, curated widgets, targeting & slotting rules, and published dynamic pages).
  - Deterministic targeting rule evaluator (platform, semver app versions, city/zone, segment, percentage rollouts, window class, dark mode, custom attributes) and UTC schedule windows.
  - Dynamic slotting engine (`insert`, `prepend`, `append`, `replace`) and v1 KMP `PageComposer`.
  - Authoring & governance lifecycle (`draft -> validate -> publish -> versions -> rollback -> fallback -> changelog`) and embedded Visual Layout Studio SPA (`GET /studio`).
  - Drop-in `abeta-dev/kmp` runtime BFF API (`GET /api/v1/screens`, `GET /api/v1/screens/:screenId`, `GET /api/v1/pages/:slug`, `GET /api/v1/config/splash`, `GET /api/v1/catalog/tokens`, `POST /api/v1/actions/execute`, `POST /api/v1/orders/express`).
  - Dual SQLite (`modernc.org/sqlite`) + PostgreSQL 17 (`pgx`) persistence, Docker/Compose packaging, OpenAPI 3.1 + Scalar docs (`/openapi.json`, `/docs`), and CI/CD security gates.

## Non-Negotiable Architecture & PR Review Gates
1. **Strict Clean Architecture**: `internal/domain` and `internal/engine` must pass `internal/domain/arch_test.go` with zero framework or database driver imports.
2. **1:1 KMP Wire Parity**: Every emitted `PageResponse`, `SnippetEnvelope`, `ActionTrigger`, `FloatingCartSpec`, `AdaptivePageSpec`, and `ginmw.Envelope` must match `github.com/abeta-dev/kmp/bff@v0.5.0` verbatim (28 wire atoms, 21 snippet types, 29 `ComposedNodeKind`s, 18 action types).
3. **Publish, Rollback & Fallback Safety**: No unvalidated draft may ever leak to runtime endpoints; publishing must store immutable snapshots with resolved widgets; runtime resolution errors must fail over cleanly to `FallbackLayout` with `X-Layout-Fallback: true`.
4. **SQLite & PostgreSQL Parity**: Migrations in `migrations/sqlite/` and `migrations/postgres/` must remain synchronized, idempotent, and backed solely by parameterized SQL queries.
5. **Security & CI Discipline**: Workflows must pin minimal `permissions:` and `concurrency:`, run SHA-256 verified MIT `gitleaks` CLI `v8.30.1`, enforce fail-closed `AUTH_REQUIRED`/`JWT_SECRET` config validation, and pass `./scripts/check_brand_hygiene.sh`.
6. **Quality Gate**: `gofmt -l .` empty, `go vet ./...` clean, `go test -count=1 -race -cover ./...` green, and `CGO_ENABLED=0 go build ./cmd/layout-engine` clean.
