# Agent Instructions — Veeam Terraform Provider

## Project Overview
Terraform provider for Veeam Backup & Replication V13 REST API (OpenAPI 3.0, version 1.3-rev0).
Written in Go, using the HashiCorp Terraform Plugin Framework (`terraform-plugin-framework`).

## Security — Non-Negotiable Rules

1. **NEVER hardcode secrets** — No tokens, passwords, API keys, or connection strings in code, tests, examples, or comments.
2. **Sensitive fields** — Mark all password/token/key fields with `Sensitive: true` in Terraform schemas. Use `json:"-"` on model fields that should never be serialised to logs.
3. **Logging** — Never log `Authorization` headers, tokens, passwords, or private keys. Redact sensitive data in all log output using `tflog` with field filtering.
4. **Environment variables only** — Credentials come from env vars (`VEEAM_HOST`, `VEEAM_USERNAME`, `VEEAM_PASSWORD`, `VEEAM_INSECURE`, `VEEAM_PORT`) or Terraform config (which stores them in state encrypted). Never read from files unless explicitly designed with secure permissions.
5. **TLS by default** — `insecure` (skip TLS verify) defaults to `false`. If enabled, log a clear warning. Never disable TLS verification silently.
6. **No shell commands** — Pure Go APIs only. No `os/exec`, no `shell=true`, no subprocess calls.
7. **.gitignore** — `.env*`, `*.pem`, `*.key`, `secrets/`, `*.tfstate*`, `*.tfvars` are gitignored. Verify before every commit.

## Architecture

```
cmd/veeam/main.go              — provider entry point
internal/
  provider.go                   — provider schema + Configure
  client/
    interface.go                — APIClient interface (resources depend on this)
    client.go                   — VeeamClient: auth, HTTP, token management
    rest.go                     — generic HTTP methods + JSON helpers
    async.go                    — async task/session polling
  models/
    auth.go                     — auth request/response structs
    common.go                   — pagination, errors, enums
    credentials.go              — Credentials API models
    repositories.go             — Repository API models
    proxies.go                  — Proxy API models
    managed_servers.go          — ManagedServer API models
    jobs.go                     — Job API models (Backup, Replica, etc.)
    protection_groups.go        — ProtectionGroup API models
    sessions.go                 — Session/Task models
  utils/
    retry.go                    — retry with exponential backoff
pkg/
  resources/                    — Terraform resources (CRUD)
  datasources/                  — Terraform data sources (read-only)
tests/                          — acceptance tests (require TF_ACC=1)
```

## Veeam V13 REST API Patterns

### Authentication
- **Endpoint:** `POST /api/oauth2/token`
- **Content-Type:** `application/x-www-form-urlencoded`
- **Body:** `grant_type=password&username=DOMAIN\user&password=secret`
- **Response:** `TokenModel` with `access_token`, `refresh_token`, `expires_in`
- **Refresh:** Same endpoint with `grant_type=refresh_token&refresh_token=...`
- **Token lifetime:** 15 minutes. Refresh token: 14 days.

### API Versioning
- All requests MUST include header: `x-api-version: 1.3-rev0`
- Base URL pattern: `https://host:9419`

### Async Operations
- Many create/update/delete operations return `202 Accepted` with session ID.
- Poll `GET /api/v1/sessions/{id}` until `state` is `Stopped`.
- Check `result` field for `Success` / `Failed`.

### Polymorphic Types
- Repository, Proxy, Job, ManagedServer, ProtectionGroup use `oneOf` + `discriminator`.
- The `type` field determines the concrete subtype.
- Use Go interface + type switch for marshaling/unmarshaling.

### Pagination
- List endpoints support: `skip`, `limit`, `orderColumn`, `orderAsc`.
- Response wraps items in `{ "data": [...], "pagination": { "total": N, "count": N, "skip": N, "limit": N } }`.

## Coding Standards

### Go
- Go 1.24+, modules, no vendor (use go.sum for reproducibility).
- `gofmt` and `golangci-lint` must pass.
- Type-hint everything: exported functions, struct fields, interface methods.
- Keep functions small and pure. No global state.
- Error messages must be actionable: `"failed to create repository: API returned 409 Conflict: repository name already exists"`.
- Use `context.Context` everywhere for cancellation and logging.

### Terraform Plugin Framework
- Use `terraform-plugin-framework` (NOT the older SDK v2) for all new resources.
- Every resource implements: `Metadata`, `Schema`, `Configure`, `Create`, `Read`, `Update`, `Delete`.
- Import support via `ResourceWithImportState` for all resources with `id`.
- Use `types.String`, `types.Bool`, `types.Int64`, `types.List` — never raw Go types in models.
- Mark computed fields as `Computed: true`, sensitive as `Sensitive: true`.
- Use `planmodifiers.UseStateForUnknown()` for ID fields.

### Testing
- **Framework:** `go test` + `testify` (assert/mock/require).
- **Unit tests:** Alongside source (`*_test.go`). Mock the `APIClient` interface.
- **Acceptance tests:** In `tests/` directory. Require `TF_ACC=1` + env vars.
- **Coverage target:** 80%+ for resources, 90%+ for client.
- **httptest:** Use `httptest.NewTLSServer` for client tests, not real servers.
- **No real credentials in tests** — ever.

### Commits
- `Co-Authored-By: Oz <oz-agent@warp.dev>` on every commit.
- Conventional commit messages: `feat:`, `fix:`, `test:`, `docs:`, `refactor:`.

## Reference Files
- `Veeam13_swagger.json` — Full V13 API spec. The source of truth for endpoints, models, and enums.
- `SKILL.md` — Detailed implementation patterns and code examples.

## What NOT to Do
- Do not guess API schemas. Always verify against the swagger.
- Do not use `map[string]interface{}` for API payloads. Use typed structs.
- Do not catch `error` broadly. Handle specific error conditions.
- Do not commit `.env`, `.tfstate`, or any file containing credentials.
- Do not use deprecated Terraform SDK v2 patterns in new code.
