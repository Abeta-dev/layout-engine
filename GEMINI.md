# GEMINI.md — Gemini, Antigravity & JetBrains AI Instructions for `abeta-dev/layout-engine`

This repository is **Abeta Dynamic Layout Engine (`github.com/abeta-dev/layout-engine`, `v0.3.0`)**, a standalone open-source Go `1.26.0` Clean Architecture Server-Driven UI (SDUI) Layout Engine & Visual Layout Studio built for `abeta-dev/kmp` (`github.com/abeta-dev/kmp/bff@v0.5.0` — 28 Wire Atoms / 30 UI Atoms, 21 SnippetTypes, 29 `ComposedNodeKind`s, 18 ActionTypes, 7 pre-seeded vertical workspaces including `abeta-pay`) and `github.com/abeta-dev/go-libs@v0.4.0`.

Read [AGENTS.md](AGENTS.md) for the authoritative architectural contract. This file adapts routing for **Gemini CLI**, **Gemini Code Assist**, **Antigravity**, and **JetBrains AI**.

---

## Skill & Code Owner Routing

| Area | Guide / Skill | Code Owner Persona |
| :--- | :--- | :--- |
| Full Layout Engine Development | [skills/abeta-layout-engine/SKILL.md](skills/abeta-layout-engine/SKILL.md) | [agents/layout-engine-owner.md](agents/layout-engine-owner.md) |
| Clean Architecture & Persistence | [docs/architecture.md](docs/architecture.md) | [agents/layout-engine-owner.md](agents/layout-engine-owner.md) |
| Visual Studio, Rules & Slotting | [docs/studio-and-rules.md](docs/studio-and-rules.md) | [agents/layout-engine-owner.md](agents/layout-engine-owner.md) |
| KMP Wire Contract Integration | [docs/kmp-integration.md](docs/kmp-integration.md) | [agents/layout-engine-owner.md](agents/layout-engine-owner.md) |

---

## Non-Negotiable Invariants

1. **Strict Clean Architecture**: `cmd/layout-engine` -> `internal/handler` -> `internal/usecase` -> `internal/engine` -> `internal/domain` <- `internal/infra`. Zero HTTP/Gin/SQL imports inside `internal/domain` and `internal/engine` (`internal/domain/arch_test.go`).
2. **1:1 `abeta-dev/kmp` Wire Parity**: Emit typed v1 contracts from `github.com/abeta-dev/kmp/bff/pkg/sdui/{actions,atoms,snippets,theming}` wrapped in `github.com/abeta-dev/go-libs/ginmw` envelopes (`ginmw.RespondSuccess` / `ginmw.RespondError`).
3. **Immutable Publishing & Fallback Chain**: Never serve unvalidated drafts on runtime endpoints (`/api/v1/screens/:screenId`, `/api/v1/pages/:slug`). Validate -> Publish `PublishedSnapshot` -> log `ChangeLogEntry` diffs -> preserve automatic `FallbackLayout` failover (`X-Layout-Fallback: true`).
4. **Dual SQLite + PostgreSQL Support**: Keep `migrations/sqlite/` and `migrations/postgres/` synchronized and use parameterized SQL queries only.
5. **Brand, Secret & Path Hygiene**: Zero proprietary legacy brand references and zero local machine paths. CI scans commits with SHA-256 verified open-source `gitleaks` CLI `v8.30.1`.

---

## Verification Gates

```bash
./scripts/check_brand_hygiene.sh
./scripts/verify_all.sh
```
