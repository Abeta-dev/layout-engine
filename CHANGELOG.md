# Changelog

All notable changes to **`abeta-dev/layout-engine`** are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [0.1.0] - 2026-11-21

### Added
- **Clean Architecture Core (`internal/domain`, `internal/engine`, `internal/usecase`)**:
  - Multi-tenant entities and ports (`Tenant`, `WidgetTemplate`, `Widget`, `TargetingRule`, `ScheduleWindow`, `Layout`, `LayoutSlot`, `SlottingRule`, `PublishedSnapshot`, `FallbackLayout`, `SplashConfig`, `TokenOverride`, `ChangeLogEntry`, `RequestContext`, `ResolutionTrace`).
  - Automated Dependency Rule enforcement in `internal/domain/arch_test.go`.
  - Pure deterministic `DefaultRuleEvaluator` supporting `all`/`any` match modes, 13 operators (`eq`, `neq`, `in`, `not_in`, `gt`, `gte`, `lt`, `lte`, `semver_gte`, `semver_lt`, `contains`, `regex`, FNV-1a `bucket_pct`), and UTC `ScheduleWindow` evaluation including overnight hour wrap-around.
  - Dynamic `DefaultSlottingEngine` supporting `insert`, `prepend`, `append`, and `replace` rules with priority ordering, `Pinned` slot protection, and widget deduplication.
  - Pre-publish `DefaultLayoutValidator` checking all 14 KMP `v1` `SnippetTypes`, adaptive layout modes, inline envelopes, and cyclic `child_widget_ids` detection.
  - Immutable `PublishedSnapshot` versioning, instant historical rollback, field-level JSON diff auditing (`ChangeLogEntry`), and automatic `FallbackLayout` failover (`X-Layout-Fallback: true`).
- **Drop-In `abeta-dev/kmp` Runtime SDUI API & Authoring API (`internal/handler`)**:
  - Runtime endpoints emitting `github.com/abeta-dev/kmp/bff@v0.1.0` `v1` contracts inside `github.com/abeta-dev/go-libs@v0.4.0` `ginmw.Envelope`: `GET /api/v1/screens`, `GET /api/v1/screens/:screenId`, `GET /api/v1/pages/:slug` (`?include_trace=true`), `GET /api/v1/config/splash`, `GET /api/v1/catalog/tokens`, `POST /api/v1/actions/execute`, `POST /api/v1/orders/express`.
  - Complete Studio CMS & governance API at `/api/v1/studio/{tenants,templates,widgets,layouts,rules,slotting,splash,tokens,changelog}` including `validate`, `publish`, `versions`, `rollback`, `fallback`, and `preview`.
  - Embedded zero-build Visual Layout Studio SPA at `GET /studio`.
  - Live OpenAPI 3.1 spec at `GET /openapi.json` and interactive Scalar API Reference UI at `GET /docs`.
  - Negotiated Brotli (`br`) & gzip compression, SHA-256 `ETag`/`304`, HTTP/1.1 + `h2c` + TLS HTTP/2, `/healthz`, `/readyz`, and `/metrics`.
- **Dual Persistence & Starter Seed (`internal/infra`, `migrations/`)**:
  - Zero-config pure-Go SQLite WAL (`modernc.org/sqlite`) default on `:8090` plus production PostgreSQL 17 (`pgx`) support with embedded schema migrations.
  - Idempotent `abeta-fresh` starter seed containing 14 KMP snippet templates, 16+ widgets, 5 targeting rules, 2 dynamic slotting rules, and 5 published pages (`store-home`, `flash-sale`, `home`, `catalog`, `watch-status`).
- **DevOps, Containers, CI/CD & Documentation**:
  - Multi-stage non-root `Dockerfile`, healthchecked PostgreSQL 17 `docker-compose.yaml`, `Makefile`, `.env.example`, and `.gitleaks.toml`.
  - GitHub Actions CI (`ci.yml`: brand/path hygiene, SHA-256 verified `gitleaks` CLI `v8.30.1`, SQLite + PostgreSQL 17 race tests, OCI build) and SemVer Release (`release.yml`: cross-platform static binaries + multi-arch GHCR container publishing).
  - Full documentation suite (`docs/architecture.md`, `docs/studio-and-rules.md`, `docs/kmp-integration.md`), AI coding agent guides (`AGENTS.md`, `CLAUDE.md`, `GEMINI.md`), skill (`skills/abeta-layout-engine/SKILL.md`), and Code Owner persona (`agents/layout-engine-owner.md`).
