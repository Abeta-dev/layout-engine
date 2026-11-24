# AGENTS.md — AI Coding Agent Guide for `abeta-dev/layout-engine`

This file provides routing, architectural invariants, and verification gates for autonomous coding agents working in **`github.com/abeta-dev/layout-engine`** (Go `1.26.0` Clean Architecture Dynamic SDUI Layout Engine & Visual Layout Studio).

---

## 1. Agent Skill & Code Owner Routing Table

Load the matching skill and Code Owner persona before planning or editing code:

| Task Domain | Skill Path | Persona / Code Owner Reference |
| :--- | :--- | :--- |
| Full Layout Engine (domain, rule/slotting engine, usecases, Gin API, Studio SPA, SQLite/Postgres persistence, CI/CD) | [skills/abeta-layout-engine/SKILL.md](skills/abeta-layout-engine/SKILL.md) | [agents/layout-engine-owner.md](agents/layout-engine-owner.md) |
| Repository architecture, persistence & runtime resolution pipeline | [docs/architecture.md](docs/architecture.md) | [agents/layout-engine-owner.md](agents/layout-engine-owner.md) |
| Visual Layout Studio, targeting rules, scheduling, slotting & rollbacks | [docs/studio-and-rules.md](docs/studio-and-rules.md) | [agents/layout-engine-owner.md](agents/layout-engine-owner.md) |
| `abeta-dev/kmp` client wire parity, drop-in BFF mode & adaptive pages | [docs/kmp-integration.md](docs/kmp-integration.md) | [agents/layout-engine-owner.md](agents/layout-engine-owner.md) |

### Multi-IDE & Multi-Model Entrypoints
- **Claude Code**: [CLAUDE.md](CLAUDE.md)
- **Gemini / Antigravity / Android Studio AI**: [GEMINI.md](GEMINI.md)
- **GitHub Copilot & Codex**: [AGENTS.md](AGENTS.md)
- **Code Ownership**: [.github/CODEOWNERS](.github/CODEOWNERS)

---

## 2. Architectural Truth Gates (Non-Negotiable)

All changes MUST satisfy these invariants:

1. **Uncle Bob Dependency Rule (Strict Go Clean Architecture)**:
   - Layering flows strictly inward:
     `cmd/layout-engine` -> `internal/handler` -> `internal/usecase` -> `internal/engine` -> `internal/domain` <- `internal/infra`.
   - `internal/domain` and `internal/engine` MUST remain pure Go with **zero** imports of `github.com/gin-gonic/gin`, `net/http`, `database/sql`, `modernc.org/sqlite`, `github.com/jackc/pgx`, `internal/infra`, or `internal/handler`.
   - This invariant is enforced in CI by `internal/domain/arch_test.go`.
2. **1:1 `abeta-dev/kmp` Wire Protocol Consumption**:
   - Runtime pages, snippets, actions, and theming MUST emit typed v1 contracts directly from `github.com/abeta-dev/kmp/bff@v0.6.0` (`pkg/sdui/{actions,atoms,snippets,molecules,theming}` — 29 Wire Atoms / 31 UI Atoms incl. `MediaPagerAtom`, 21 SnippetTypes, 36 `ComposedNodeKind`s with 9-position `box_alignment` and `ComposedNode.Tracking`/`ClickTracking`, `SnippetEnvelope.Sticky`, deterministic `snippets.EnsureComponentIDs`, 18 ActionTypes, 7 ambient `MotionLoopEffect`s, `AdaptivePageSpec`, `FloatingCartSpec` with free-delivery threshold nudges, `MotionSpec`).
   - All `/api/v1/*` JSON responses MUST wrap payloads in `github.com/abeta-dev/go-libs@v0.4.0` `ginmw.Envelope` (`ginmw.RespondSuccess`, `ginmw.RespondError`, `ginmw.RespondErrorWithDetail`) so KMP clients (`--bff-url=http://localhost:8090`) can use the Layout Engine as a drop-in BFF replacement.
3. **Safe Authoring, Snapshots, Rollback & Fallback Governance**:
   - Layouts follow `draft -> validate -> publish (immutable PublishedSnapshot) -> rollback` with automatic `ChangeLogEntry` JSON field diffs.
   - Runtime serving reads immutable `PublishedSnapshot` records—never unvalidated drafts (drafts are only evaluated via `POST /api/v1/studio/layouts/:slug/preview` with `draft: true`).
   - When snapshot resolution or composition fails at runtime, `ResolveService` MUST fall back to the pre-validated `FallbackLayout`, setting `X-Layout-Fallback: true` and populating `ResolutionTrace.UsedFallback`.
4. **Dual-Database Parity (Zero-Config SQLite + Production PostgreSQL 17)**:
   - Default zero-config startup uses `PORT=8090`, `DB_DIALECT=sqlite`, and `DB_DSN=file:layout-engine.db?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)` with auto-migration and 7 pre-seeded multi-tenant vertical workspaces (`abeta-fresh` default, `abeta-bites`, `abeta-pulse`, `abeta-mart`, `abeta-spark`, `abeta-trade`, `abeta-pay`), each seeded with all 21 canonical KMP `v1` templates, curated widgets (including Express molecule molded `composed_block` widgets across all 36 `ComposedNodeKind`s and sticky search bars), targeting & slotting rules, and published dynamic pages.
   - Any schema or SQL query change MUST maintain parity across both `migrations/sqlite/` and `migrations/postgres/` and pass parameterized queries without SQL string interpolation.
5. **Security, Transport & CI Hygiene**:
   - Preserve Brotli/gzip compression, SHA-256 `ETag`/`304`, HTTP/1.1 + `h2c` + TLS HTTP/2 support, fail-closed `AUTH_REQUIRED=true` / `JWT_SECRET` validation, explicit minimal `permissions:` + `concurrency:` in every GitHub Actions workflow, and the unified `CI Quality Gate` (`ci-quality-gate`) aggregator job in `.github/workflows/ci.yml`.
   - Secret scanning in CI MUST use the open-source MIT `gitleaks` CLI `v8.30.1` with SHA-256 verification—never paid-gated GitHub Actions wrappers.
6. **Zero Legacy Brand Leaks & Zero Machine-Specific Paths**:
   - Strictly use `abeta-dev` / `Abeta-dev` and `github.com/abeta-dev/layout-engine`.
   - Never commit proprietary legacy brand names or local home-directory paths into source, docs, scripts, or workflows.
   - Run `./scripts/check_brand_hygiene.sh` before concluding any task.

---

## 3. Verification Commands

Run the checks relevant to your change—or the full gate before opening a PR:

```bash
# 1. Brand & machine-path hygiene
./scripts/check_brand_hygiene.sh

# 2. Formatting, static analysis, Clean Architecture guard & race-enabled tests
test -z "$(gofmt -l .)"
go vet ./...
go test -count=1 -race -cover ./...

# 3. Build static binary & smoke-start on :8090
CGO_ENABLED=0 go build -trimpath -o bin/layout-engine ./cmd/layout-engine
go run ./cmd/layout-engine

# 4. Complete verification gate (also available via `make verify`)
./scripts/verify_all.sh
```

> **Honesty Rule**: Never claim a build, lint, or test suite passed unless you executed it after your final edit and directly observed its passing output.
