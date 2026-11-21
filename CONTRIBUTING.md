# Contributing to `abeta-dev/layout-engine`

Thank you for contributing to **Abeta Dynamic Layout Engine**! We welcome bug reports, documentation enhancements, targeting/slotting capabilities, Visual Layout Studio improvements, and persistence optimizations that respect our Clean Architecture boundaries.

---

## Development Environment

- **Go**: `1.26.0` (pure Go; `CGO_ENABLED=0` works out of the box with `modernc.org/sqlite`)
- **Docker** (optional): For testing against `postgres:17-alpine` via `docker compose up --build`
- **OS**: macOS, Linux, or Windows (WSL2 recommended for shell verification scripts)

### Initial Setup

```bash
git clone https://github.com/Abeta-dev/layout-engine.git
cd layout-engine
./scripts/verify_all.sh
```

---

## Architectural Rules for Pull Requests

Every pull request MUST preserve these invariants:

1. **Strict Go Clean Architecture (`internal/domain/arch_test.go`)**:
   - Dependencies flow strictly inward: `cmd/layout-engine` -> `internal/handler` -> `internal/usecase` -> `internal/engine` -> `internal/domain` <- `internal/infra`.
   - `internal/domain` and `internal/engine` MUST NEVER import `gin`, `net/http`, `database/sql`, `modernc.org/sqlite`, `pgx`, `internal/infra`, or `internal/handler`. See [docs/architecture.md](docs/architecture.md).
2. **1:1 `abeta-dev/kmp` Wire Protocol Compatibility**:
   - Consume v1 SDUI types directly from `github.com/abeta-dev/kmp/bff@v0.1.0` (`pkg/sdui/{actions,atoms,snippets,theming}`).
   - Wrap `/api/v1/*` HTTP responses in `github.com/abeta-dev/go-libs@v0.4.0` `ginmw.Envelope` responses. See [docs/kmp-integration.md](docs/kmp-integration.md).
3. **Immutable Snapshots, Validation & Fallback Safety**:
   - Runtime serving MUST read validated `PublishedSnapshot` records with automatic failover to `FallbackLayout`.
   - Studio mutations MUST emit audit `ChangeLogEntry` records with field-level JSON diffs. See [docs/studio-and-rules.md](docs/studio-and-rules.md).
4. **Dual SQLite + PostgreSQL Parity**:
   - Schema changes MUST update both `migrations/sqlite/` and `migrations/postgres/` identically and use parameterized SQL queries only.
5. **Brand, Secret & Path Hygiene**:
   - Use `abeta-dev` / `Abeta-dev` and `github.com/abeta-dev/layout-engine`.
   - Never commit proprietary legacy brand references, secrets, or machine-specific absolute paths.

---

## Pull Request Workflow

1. Create a feature branch from `main`:
   ```bash
   git checkout -b feat/my-feature
   ```
2. Follow [Conventional Commits](https://www.conventionalcommits.org/) (`feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`).
3. Format Go code with `gofmt`:
   ```bash
   gofmt -w .
   ```
4. Run the full verification pipeline before opening your PR:
   ```bash
   ./scripts/verify_all.sh
   ```
5. Open a Pull Request against `main` summarizing:
   - What changed and why
   - Which architectural layers (`domain`, `engine`, `usecase`, `handler`, `infra`, `migrations`) were modified
   - Verification commands run and observed results
