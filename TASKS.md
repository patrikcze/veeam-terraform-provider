# TASKS — Veeam Terraform Provider Backlog

> Generated: 2026-03-27
> Based on: Swagger `1.3-rev1`, current codebase audit, real VBR validation notes.

---

## Current Status Summary

### Implemented Resources (10)
| Resource | Status | VBR Validated |
|----------|--------|---------------|
| `veeam_credential` | Done | Yes |
| `veeam_managed_server` | Done | Yes (LinuxHost) |
| `veeam_repository` | Done | Yes (WinLocal) |
| `veeam_proxy` | Done | Yes |
| `veeam_scale_out_repository` | Done | Partial |
| `veeam_cloud_credential` | Done | Yes (AzureStorage) |
| `veeam_encryption_password` | Done | Yes |
| `veeam_configuration_backup` | Done | Yes |
| `veeam_backup_job` | Done | Yes (LinuxAgentBackup, VSphereBackup models) |
| `veeam_protection_group` | Done | Yes (IndividualComputers, CloudMachines) |

### Implemented Data Sources (14)
| Data Source | Status | VBR Validated |
|-------------|--------|---------------|
| `veeam_backups` | Done | Yes |
| `veeam_backup_jobs` | Done | Yes |
| `veeam_credentials` | Done | Yes |
| `veeam_job_states` | Done | Yes |
| `veeam_license` | Done | Yes |
| `veeam_managed_servers` | Done | Yes |
| `veeam_protection_groups` | Done | Yes |
| `veeam_proxies` | Done | Yes |
| `veeam_repositories` | Done | Yes |
| `veeam_repository_states` | Done | Yes |
| `veeam_restore_points` | Done | Yes |
| `veeam_server_info` | Done | Yes |
| `veeam_sessions` | Done | Yes |
| `veeam_wan_accelerators` | Done | Yes |

### Test Coverage
| Package | Coverage | Target |
|---------|----------|--------|
| `internal/client` | 70.5% | 90% |
| `internal/models` | 0.0% | 80% |
| `internal/utils` | 100% | 90% |
| `pkg/datasources` | 9.9% | 80% |
| `pkg/resources` | 29.0% | 80% |

---

## Backlog — Prioritized Tasks

### Priority 1 — Test Coverage & Quality (Must-Have for Release)

- [ ] **T1.1** Increase `internal/models` test coverage to 80%+
  - Add JSON marshal/unmarshal round-trip tests for all model files
  - Cover: `credentials.go`, `repositories.go`, `proxies.go`, `managed_servers.go`, `jobs.go`, `protection_groups.go`, `cloud_credentials.go`, `configuration_backup.go`, `scale_out_repositories.go`, `sessions.go`
  - Current: 0% (only `models_test.go` with 16 tests, likely covers `common.go`/`auth.go`)

- [ ] **T1.2** Increase `pkg/resources` test coverage to 80%+
  - Add `buildSpec()` ↔ `syncModelFromAPI()` round-trip tests for each resource
  - Add error path tests (API failure, 404 on Read, async timeout)
  - Add import state tests for all resources
  - Current: 29% across 10 resources (63 tests total)

- [ ] **T1.3** Increase `pkg/datasources` test coverage to 80%+
  - Add per-datasource response parsing + filter tests (currently only backup_jobs, repositories, server_info, helpers have tests)
  - Missing tests for: `backups`, `credentials`, `job_states`, `license`, `managed_servers`, `protection_groups`, `proxies`, `repository_states`, `restore_points`, `sessions`, `wan_accelerators`
  - Current: 9.9% across 14 data sources

- [ ] **T1.4** Increase `internal/client` test coverage to 90%+
  - Add tests for edge cases: refresh token rotation, concurrent token refresh, 429 rate-limiting
  - Add `WaitForTask` timeout/context-cancel tests
  - Current: 70.5%

- [ ] **T1.5** Fix Makefile GOROOT issue
  - `make test` fails with `go: cannot find GOROOT directory: /usr/local/go` — the `go tool compile -V` check references old GOROOT
  - Direct `go test` works fine; Makefile's GOROOT detection needs fix

---

### Priority 2 — New Resources (High-Value API Coverage)

- [ ] **T2.1** `veeam_general_options` — Singleton resource (GET/PUT)
  - API: `GET/PUT /api/v1/generalOptions`
  - Manages server-wide general options, email settings, event forwarding, notifications, storage latency
  - High value: central configuration that affects all jobs

- [ ] **T2.2** `veeam_email_settings` — Singleton resource (GET/PUT)
  - API: `GET/PUT /api/v1/generalOptions/emailSettings`
  - SMTP configuration for notification delivery
  - Action: `POST /api/v1/generalOptions/emailSettings/testMessage`

- [ ] **T2.3** `veeam_notification_settings` — Singleton resource (GET/PUT)
  - API: `GET/PUT /api/v1/generalOptions/notifications`
  - Global notification rules

- [ ] **T2.4** `veeam_traffic_rules` — Singleton resource (GET/PUT)
  - API: `GET/PUT /api/v1/trafficRules`
  - Network traffic throttling rules

- [ ] **T2.5** `veeam_security_settings` — Singleton resource (GET/PUT)
  - API: `GET/PUT /api/v1/security/settings`
  - Security hardening configuration

- [ ] **T2.6** `veeam_kms_server` — Full CRUD resource
  - API: `GET/POST /api/v1/kmsServers`, `GET/PUT/DELETE /api/v1/kmsServers/{id}`
  - Key Management Server integration for encryption
  - Action: `POST .../changeCertificate`

- [ ] **T2.7** `veeam_security_user` — Partial CRUD resource (Create/Read/Delete)
  - API: `GET/POST /api/v1/security/users`, `GET/DELETE /api/v1/security/users/{id}`
  - RBAC user management with role assignment via `GET/PUT .../roles`
  - No update — only create/delete + role assignment

- [ ] **T2.8** `veeam_ad_domain` — Partial CRUD resource (Create/Read/Delete)
  - API: `GET/POST /api/v1/adDomains`, `GET/DELETE /api/v1/adDomains/{id}`
  - Active Directory domain registration

---

### Priority 3 — New Data Sources (Read-Only Visibility)

- [ ] **T3.1** `veeam_security_roles` — Read-only
  - API: `GET /api/v1/security/roles`
  - List available RBAC roles and permissions

- [ ] **T3.2** `veeam_security_users` — Read-only
  - API: `GET /api/v1/security/users`
  - List configured RBAC users

- [ ] **T3.3** `veeam_backup_objects` — Read-only
  - API: `GET /api/v1/backupObjects`
  - List objects within backups (VMs, machines)

- [ ] **T3.4** `veeam_replicas` — Read-only
  - API: `GET /api/v1/replicas`
  - List VM replicas

- [ ] **T3.5** `veeam_replica_points` — Read-only
  - API: `GET /api/v1/replicaPoints`
  - List replica restore points

- [ ] **T3.6** `veeam_proxy_states` — Read-only
  - API: `GET /api/v1/backupInfrastructure/proxies/states`
  - Proxy health/state information

- [ ] **T3.7** `veeam_protected_computers` — Read-only
  - API: `GET /api/v1/agents/protectedComputers`
  - List agent-protected computers

- [ ] **T3.8** `veeam_services` — Read-only
  - API: `GET /api/v1/services`
  - List VBR services and their status

- [ ] **T3.9** `veeam_server_time` — Read-only
  - API: `GET /api/v1/serverTime`
  - Server time (useful for schedule validation)

- [ ] **T3.10** `veeam_server_certificate` — Read-only
  - API: `GET /api/v1/serverCertificate`
  - Server TLS certificate details

- [ ] **T3.11** `veeam_task_sessions` — Read-only
  - API: `GET /api/v1/taskSessions`
  - Granular task-level session details

- [ ] **T3.12** `veeam_security_analyzer` — Read-only
  - API: `GET /api/v1/securityAnalyzer/bestPractices`, `GET /api/v1/securityAnalyzer/lastRun`
  - Security compliance best practices check results

- [ ] **T3.13** `veeam_malware_events` — Read-only
  - API: `GET /api/v1/malwareDetection/events`
  - Malware detection events

---

### Priority 4 — Job Type Expansion

- [ ] **T4.1** Backup job: `HyperVBackup` support
  - Models exist in `jobs.go` (HyperVBackupJobSpec/Model stubs)
  - Need resource schema variant, buildSpec, syncModelFromAPI, tests
  - Verify against Swagger `HyperVBackupJobModel`

- [ ] **T4.2** Backup job: `BackupCopy` support
  - Copy job for offsite/secondary copies
  - Different schema: source backup reference, target repository, copy schedule
  - Verify against Swagger `BackupCopyJobModel`

- [ ] **T4.3** Backup job: `VSphereReplica` support
  - VM replication jobs
  - Different target: ESXi host/datastore instead of repository
  - Verify against Swagger `VSphereReplicaJobModel`

- [ ] **T4.4** Backup job: `WindowsAgentBackup` support
  - Models already exist in `jobs.go`
  - Extend resource to create Windows agent backup jobs
  - Test with real VBR Windows agent

- [ ] **T4.5** Protection group: `ADObjects` type support
  - Active Directory-based computer discovery
  - Requires AD domain + container/OU configuration

- [ ] **T4.6** Protection group: `CSVFile` type support
  - CSV file-based computer list import

---

### Priority 5 — Infrastructure & CI/CD

- [ ] **T5.1** CI pipeline setup (GitHub Actions)
  - Lint (`golangci-lint`), vet, fmt-check, unit tests on every PR
  - Build matrix: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64
  - Acceptance test job (manual trigger with secrets)

- [ ] **T5.2** GoReleaser configuration
  - Automated binary builds + GitHub Releases
  - Terraform Registry signing (GPG key)
  - Changelog generation from conventional commits

- [ ] **T5.3** Terraform Registry publishing preparation
  - `terraform-registry-manifest.json` already exists
  - Verify registry metadata, provider docs format (`tfplugindocs`)
  - Test `terraform init` from local registry mirror

- [ ] **T5.4** Acceptance test environment automation
  - Docker/Vagrant-based VBR test environment (or documented manual setup)
  - `scripts/setup-ubuntu-test-env.sh` exists — verify and extend
  - Environment variable template for CI secrets

---

### Priority 6 — Advanced Features & Polish

- [ ] **T6.1** `veeam_mount_server` — Partial CRUD resource (Create/Read/Update)
  - API: `GET/POST /api/v1/backupInfrastructure/mountServers`, `GET/PUT /api/v1/backupInfrastructure/mountServers/{id}`
  - No delete endpoint — mount server lifecycle tied to managed server

- [ ] **T6.2** `veeam_global_vm_exclusion` — Partial CRUD resource (Create/Read/Delete)
  - API: `GET/POST /api/v1/globalExclusions/vm`, `GET/DELETE /api/v1/globalExclusions/vm/{id}`
  - Global VM exclusion list management

- [ ] **T6.3** `veeam_recovery_token` — Full CRUD resource
  - API: `GET/POST /api/v1/agents/recoveryTokens`, `GET/PUT/DELETE /api/v1/agents/recoveryTokens/{id}`
  - Agent recovery token management

- [ ] **T6.4** `veeam_entra_id_tenant` — Full CRUD resource
  - API: `GET/POST /api/v1/inventory/entraId/tenants`, `GET/PUT/DELETE /api/v1/inventory/entraId/tenants/{id}`
  - Microsoft Entra ID (Azure AD) tenant inventory management

- [ ] **T6.5** `veeam_unstructured_data_server` — Full CRUD resource
  - API: `GET/POST /api/v1/inventory/unstructuredDataServers`, `GET/PUT/DELETE /api/v1/inventory/unstructuredDataServers/{id}`
  - NAS/file share backup source management

- [ ] **T6.6** `veeam_security_analyzer_schedule` — Singleton resource (GET/PUT)
  - API: `GET/PUT /api/v1/securityAnalyzer/schedule`
  - Manage security compliance scan schedule

- [ ] **T6.7** Event forwarding resource
  - API: `GET/PUT /api/v1/generalOptions/eventForwarding`
  - SNMP/syslog event forwarding configuration

- [ ] **T6.8** Storage latency rules resource
  - API: `GET/PUT /api/v1/generalOptions/storageLatency`
  - Datastore latency throttle control + per-datastore overrides

---

### Priority 7 — Documentation & Examples

- [ ] **T7.1** Complete example configurations for all resource types
  - `examples/complete/main.tf` exists — verify all 10 resources are shown
  - Add per-resource isolated examples in `examples/resources/`

- [ ] **T7.2** Data source usage examples
  - `examples/data-sources/` exists — expand with realistic filtering patterns
  - Show data source → resource reference patterns

- [ ] **T7.3** Import guide
  - Document `terraform import` commands for every resource
  - Add import examples to resource docs

- [ ] **T7.4** Upgrade guide & migration notes
  - Prepare for future API version bumps (V14)
  - Document which file to change (`endpoints.go` + API version constant)

- [ ] **T7.5** Regenerate docs via `tfplugindocs`
  - Ensure all schema descriptions are complete and accurate
  - Validate generated markdown against Terraform Registry format

---

## API Endpoints NOT Planned for Provider (Operational/One-Off Actions)

These endpoints are intentionally excluded — they represent one-off operations, restore workflows, or browser sessions that don't map well to Terraform's declarative model:

- `/api/v1/backupBrowser/*` — Interactive file-level restore browsing
- `/api/v1/restore/*` — VM/file/Entra restore operations
- `/api/v1/failover/*`, `/api/v1/failback/*` — DR failover workflows
- `/api/v1/automation/*` — Bulk import/export (not declarative)
- `/api/v1/deployment/*` — Agent deployment kit generation
- `/api/v1/dataIntegration/*` — Data integration mounts (session-based)
- `/api/v1/cloudBrowser/*` — Cloud vault browsing
- `/api/v1/exportlogs/*` — Log export (operational)
- `/api/v1/inventory` — Inventory scanning (discovery action, not state)
- `/api/v1/registerVbr` — One-time VBR registration
- `/api/v1/malwareDetection/scanBackup` — On-demand scan action
- `/api/v1/license/install`, `/api/v1/license/update` — License management actions

---

## Quick Reference — Next Sprint Suggestions

**Sprint 1 (Test Quality):** T1.1, T1.2, T1.3, T1.5
**Sprint 2 (Singleton Resources):** T2.1, T2.2, T2.3, T2.4
**Sprint 3 (Security + CRUD):** T2.5, T2.6, T2.7, T2.8
**Sprint 4 (Data Sources):** T3.1–T3.6
**Sprint 5 (Job Types):** T4.1, T4.2, T4.4
**Sprint 6 (CI/CD + Release):** T5.1, T5.2, T5.3, T7.5
