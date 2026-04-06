# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Build
```sh
make build          # Build for current platform
make build-all      # Build for all platforms (linux/amd64, linux/arm64, darwin, windows)
make install        # Build and install to ~/.terraform.d/plugins/
```

### Test
```sh
make test           # Unit tests only (./internal/... and ./pkg/...)
make check          # Run fmt-check, vet, lint, and unit tests together (mirrors PR CI)

# Acceptance tests (require VEEAM_HOST, VEEAM_USERNAME, VEEAM_PASSWORD, VEEAM_INSECURE)
make testacc                     # All acceptance tests
make testacc-credential          # Single resource acceptance test
make testacc-repository
make testacc-backup-job
make testacc-workflow
make testacc-proxy
make testacc-scale-out-repository
```

To run a single unit test:
```sh
go test ./pkg/resources/ -run TestCredentialResource -v
```

### Lint & Format
```sh
make lint       # golangci-lint
make fmt        # Format code
make vet        # go vet
```

### Docs
```sh
make docs       # Regenerate docs/ via tfplugindocs
```

## Architecture

This is a Terraform Plugin Framework provider for Veeam Backup & Replication V13.

**Layers:**
- `cmd/veeam/main.go` — entry point, serves the provider
- `internal/provider.go` — provider config (host/port/username/password/insecure), registers all resources and data sources
- `internal/client/` — API client: `client.go` (OAuth2 auth), `rest.go` (HTTP ops), `async.go` (job polling), `endpoints.go` (all API paths)
- `internal/models/` — Go structs for API request/response bodies
- `pkg/resources/` — 10 CRUD resource implementations
- `pkg/datasources/` — 14 read-only data source implementations
- `tests/` — acceptance tests

**Key design decisions:**

1. **Centralized API paths**: All API endpoints live in `internal/client/endpoints.go`. When upgrading the Veeam API version, update only this file.

2. **Resource pattern**: Every resource follows the same structure:
   - Compile-time interface assertions (`var _ resource.Resource = &Foo{}`)
   - A model struct with `tfsdk` tags
   - `buildSpec()` to convert Terraform state → API request
   - `syncModelFromAPI()` to merge API response → Terraform state
   - Sensitive fields (passwords) are never overwritten from API responses

3. **Async operations**: Long-running API calls return a job ID; `internal/client/async.go` polls until completion.

4. **Retry logic**: `internal/utils/` handles exponential backoff. Credential deletes retry up to 12 times (resources may hold references).

5. **Import support**: All resources implement `ImportState()`, so `terraform import veeam_resource.name <id>` works for every resource.

6. **Environment variables**: Provider attributes map to `VEEAM_HOST`, `VEEAM_PORT`, `VEEAM_USERNAME`, `VEEAM_PASSWORD`, `VEEAM_INSECURE`. Default port is 9419.
