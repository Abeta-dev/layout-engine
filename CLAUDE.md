# CLAUDE.md — Claude Code Instructions for `abeta-dev/layout-engine`

This repository is **Abeta Dynamic Layout Engine (`github.com/abeta-dev/layout-engine`)**, a standalone Go `1.26.0` Clean Architecture Server-Driven UI (SDUI) Layout Engine & Visual Layout Studio consuming `github.com/abeta-dev/kmp/bff@v0.1.0` and `github.com/abeta-dev/go-libs@v0.4.0`.

Read [AGENTS.md](AGENTS.md) for the complete architectural contract. Below is the operational quick reference for every session.

---

## Skill & Owner Routing

Before editing code, load:
- **Engineering Skill**: [skills/abeta-layout-engine/SKILL.md](skills/abeta-layout-engine/SKILL.md)
- **Code Owner Persona**: [agents/layout-engine-owner.md](agents/layout-engine-owner.md)
- **Deep Guides**: [docs/architecture.md](docs/architecture.md), [docs/studio-and-rules.md](docs/studio-and-rules.md), [docs/kmp-integration.md](docs/kmp-integration.md)

---

## Architecture & Invariants

- **Clean Architecture Boundaries**:
  - `cmd/layout-engine`: Entrypoint, env config (`go-libs/env`), auto-migration, `abeta-fresh` starter seed, HTTP/1.1 + `h2c` + TLS HTTP/2 server, graceful shutdown (`go-libs/shutdown`).
  - `internal/domain`: Pure entities (`Tenant`, `WidgetTemplate`, `Widget`, `TargetingRule`, `ScheduleWindow`, `Layout`, `LayoutSlot`, `SlottingRule`, `PublishedSnapshot`, `FallbackLayout`, `SplashConfig`, `TokenOverride`, `ChangeLogEntry`, `RequestContext`, `ResolutionTrace`) and ports (`Store`, `RuleEvaluator`, `SlottingEngine`, `PageComposer`, `LayoutValidator`). Zero HTTP/Gin/SQL imports (`arch_test.go`).
  - `internal/engine`: Pure Go rule evaluation (semver, `%` rollouts, `all`/`any`), schedule windows, dynamic slotting (`insert`/`prepend`/`append`/`replace`), v1 KMP page composition, and pre-publish validation.
  - `internal/usecase`: `ResolveService`, `StudioService`, `ActionService`.
  - `internal/handler`: Gin router, embedded Visual Studio SPA (`GET /studio`), OpenAPI 3.1 (`GET /openapi.json`) & Scalar UI (`GET /docs`), runtime KMP-compatible SDUI API, Studio authoring API, compression (`br`/`gzip`), SHA-256 `ETag`.
  - `internal/infra`: Dual SQLite (`modernc.org/sqlite`) + PostgreSQL (`pgx`) `Store`, migrations (`migrations/{sqlite,postgres}`), and starter seeder (`abeta-fresh`).
- **Default Runtime**: `PORT=8090`, `DB_DIALECT=sqlite`, `DB_DSN=file:layout-engine.db?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)`.
- **Drop-In KMP BFF Compatibility**: `GET /api/v1/screens`, `GET /api/v1/screens/:screenId`, `GET /api/v1/pages/:slug` (`?include_trace=true`), `GET /api/v1/config/splash`, `GET /api/v1/catalog/tokens`, `POST /api/v1/actions/execute`, `POST /api/v1/orders/express`.
- **Brand & Path Hygiene**: Strictly use `abeta-dev` / `Abeta-dev` and `github.com/abeta-dev/layout-engine`. Never commit legacy brand names or local filesystem paths.

---

## Essential Commands

```bash
# Run locally with zero-config SQLite + starter seed on :8090
go run ./cmd/layout-engine

# Brand & path hygiene check
./scripts/check_brand_hygiene.sh

# Format, vet, and race-enabled unit/integration tests
gofmt -w .
go vet ./...
go test -count=1 -race -cover ./...

# Full verification suite
./scripts/verify_all.sh
```
