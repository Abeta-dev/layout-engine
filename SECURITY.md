# Security Policy

## Supported Versions

The `Abeta-dev/layout-engine` maintainers provide security updates for the latest minor release line:

| Version | Supported |
| :--- | :--- |
| `0.1.x` | Yes |
| `< 0.1.0` | No |

---

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues, discussions, or pull requests.**

Report vulnerabilities privately through **GitHub Private Vulnerability Reporting**:

1. Navigate to the repository's **Security** tab (`https://github.com/Abeta-dev/layout-engine/security/advisories`).
2. Click **"Report a vulnerability"**.
3. Include as much actionable detail as possible:
   - Affected component (`handler`, `engine`, `usecase`, `infra`, `studio`, `migrations`, Docker/CI) and version/commit
   - Minimal reproduction steps or proof-of-concept
   - Impact assessment (e.g., tenant isolation bypass, unauthorized publish/rollback, SQL injection, XSS in `/studio`, auth bypass)
   - Suggested remediation, if known

### Response Timeline

- **Acknowledgment**: Within 48 hours of receipt
- **Triage & Assessment**: Within 5 business days
- **Remediation & Coordinated Disclosure**: Prioritized by severity (CVSS) and published with a security advisory

---

## Security Practices in `abeta-dev/layout-engine`

- **Strict Pre-Publish Validation**: Draft layouts are validated against the 14 allowed `v1` KMP `SnippetTypes`, URL-safe slugs, valid adaptive layout modes, and recursive container cycle checks before any snapshot can be published.
- **Tenant Isolation & Parameterized SQL**: Every persistence query scopes resources by `tenant_id` and binds parameters safely across both SQLite (`modernc.org/sqlite`) and PostgreSQL (`pgx`).
- **Transport & Middleware Hardening**: Uses `github.com/abeta-dev/go-libs@v0.4.0` (`ginmw`) for request IDs, panic recovery, structured logging, security headers, optional global rate limiting (`RATE_LIMIT_ENABLED=true`), and fail-closed JWT authentication (`AUTH_REQUIRED=true` requires a non-empty `JWT_SECRET`).
- **Non-Root Container Runtime**: The official `Dockerfile` compiles a static `CGO_ENABLED=0` Go binary and executes as non-root UID `10001` on `alpine:3.21`.
- **Supply-Chain & Secret Scanning**: `.gitignore` blocks `.env*`, local databases, and key material. GitHub Actions CI enforces minimal `permissions:`, brand/path hygiene checks (`./scripts/check_brand_hygiene.sh`), and SHA-256 verified open-source `gitleaks` CLI (`v8.30.1`) secret scanning on every push and pull request.
