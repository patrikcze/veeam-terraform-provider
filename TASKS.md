# TASKS — Veeam Terraform Provider Backlog

> Generated: 2026-03-27
> Last updated: 2026-04-07
> Based on: Swagger `1.3-rev1`, current codebase audit, real VBR validation notes.
> Smoke test (datasources): All 27 datasources validated against live VBR on 2026-04-07 (see `examples/datasources-smoke-test/`).
> Smoke test (resources): Created 2026-04-07 (see `examples/resources-smoke-test/`). Tier 1 singletons safe to apply anytime; Tier 3 requires vcenter_host/linux_server/windows_server vars.

---

## Current Status Summary

### Implemented Resources (27)
| Resource | Status | VBR Validated |
|----------|--------|---------------|
| `veeam_ad_domain` | Done | — |
| `veeam_backup_job` | Done | Yes (LinuxAgentBackup, VSphereBackup models) |
| `veeam_cloud_credential` | Done | Yes (AzureStorage) |
| `veeam_configuration_backup` | Done | Yes |
| `veeam_credential` | Done | Yes |
| `veeam_email_settings` | Done | — (previously unregistered, fixed 2026-04-07) |
| `veeam_encryption_password` | Done | Yes |
| `veeam_entra_id_tenant` | Done | — |
| `veeam_event_forwarding` | Done | — |
| `veeam_general_options` | Done | — (previously unregistered, fixed 2026-04-07) |
| `veeam_global_vm_exclusion` | Done | — |
| `veeam_kms_server` | Done | — |
| `veeam_managed_server` | Done | Yes (ViHost/vSphere, LinuxHost, WindowsHost) |
| `veeam_mount_server` | Done | — |
| `veeam_notification_settings` | Done | — (previously unregistered, fixed 2026-04-07) |
| `veeam_protection_group` | Done | Yes (IndividualComputers, CloudMachines) |
| `veeam_proxy` | Done | Yes |
| `veeam_recovery_token` | Done | — |
| `veeam_repository` | Done | Yes (WinLocal) |
| `veeam_scale_out_repository` | Done | Partial |
| `veeam_security_analyzer_schedule` | Done | — |
| `veeam_security_settings` | Done | — (previously unregistered, fixed 2026-04-07) |
| `veeam_security_user` | Done | — |
| `veeam_storage_latency` | Done | — |
| `veeam_traffic_rules` | Done | — (previously unregistered, fixed 2026-04-07) |
| `veeam_unstructured_data_server` | Done | — |
| `veeam_vsphere_server` | Done | — (new 2026-04-07; dedicated ViHost resource) |

### Implemented Data Sources (27)
| Data Source | Status | VBR Validated |
|-------------|--------|---------------|
| `veeam_backups` | Done | Yes |
| `veeam_backup_jobs` | Done | Yes |
| `veeam_backup_objects` | Done | Yes (smoke test 2026-04-07) |
| `veeam_credentials` | Done | Yes |
| `veeam_job_states` | Done | Yes |
| `veeam_license` | Done | Yes |
| `veeam_malware_events` | Done | Yes (smoke test 2026-04-07) |
| `veeam_managed_servers` | Done | Yes |
| `veeam_protected_computers` | Done | Yes (smoke test 2026-04-07) |
| `veeam_protection_groups` | Done | Yes |
| `veeam_proxies` | Done | Yes |
| `veeam_proxy_states` | Done | Yes — `isOnline` bool → "Online"/"Offline" (smoke test 2026-04-07) |
| `veeam_replicas` | Done | Yes (empty — no VM replicas in test env) |
| `veeam_replica_points` | Done | Yes (empty — no VM replicas in test env) |
| `veeam_repositories` | Done | Yes |
| `veeam_repository_states` | Done | Yes — `capacityGB`/`freeGB`/`usedSpaceGB` fields (smoke test 2026-04-07) |
| `veeam_restore_points` | Done | Yes |
| `veeam_security_analyzer` | Done | Yes — 38 checks; `bestPractice`/`note` fields; `creationTime` for lastRun (smoke test 2026-04-07) |
| `veeam_security_roles` | Done | Yes (smoke test 2026-04-07) |
| `veeam_security_users` | Done | Yes — `name` field (not `login`); roles via nested array (smoke test 2026-04-07) |
| `veeam_server_certificate` | Done | Yes — `validBy` field (not `validTo`) (smoke test 2026-04-07) |
| `veeam_server_info` | Done | Yes — `buildVersion`/`vbrId` fields (smoke test 2026-04-07) |
| `veeam_server_time` | Done | Yes — `utc_offset` extracted from RFC3339 timestamp (smoke test 2026-04-07) |
| `veeam_services` | Done | Yes — API returns only `name`/`port`; status/version always empty by design |
| `veeam_sessions` | Done | Yes — `result` is nested object `{result,message,isCanceled}` (smoke test 2026-04-07) |
| `veeam_task_sessions` | Done | Yes (smoke test 2026-04-07) |
| `veeam_wan_accelerators` | Done | Yes |

### Test Coverage
| Package | Coverage | Target |
|---------|----------|--------|
| `internal/client` | 70.5% | 90% |
| `internal/models` | 100% | 80% |
| `internal/utils` | 100% | 90% |
| `pkg/datasources` | 92.0% | 80% |
| `pkg/resources` | 80.2% | 80% |

---

## Backlog — Prioritized Tasks

### Priority 1 — Test Coverage & Quality (Must-Have for Release)

- [x] **T1.1** Increase `internal/models` test coverage to 80%+
  - Added tests for `auth.go` functions (`TokenInfo.IsExpired`, `WillExpireSoon`, `String`, `APIError.Error`)
  - Achieved: 100% ✅

- [x] **T1.2** Increase `pkg/resources` test coverage to 80%+
  - Added `resources_extra_test.go` with ImportState tests for all 10 resources, CRUD error paths, helper function tests
  - Achieved: 80.2% ✅

- [x] **T1.3** Increase `pkg/datasources` test coverage to 80%+
  - Added per-datasource response parsing + filter tests for all 14 data sources
  - Achieved: 92.0% ✅

- [x] **T1.4** Increase `internal/client` test coverage to 90%+
  - Added `PutJSON`, `PostJSON`, `DeleteJSON`, `GetJSON` error/happy-path tests
  - Added `WaitForTask` wrapper, `normalizeSessionResult` table tests, unknown-state branch
  - Added `NewVeeamClient` empty host, insecure flag, auth failure tests
  - Added `readAndClose` error path, broken-body transport, doRequest edge cases
  - Achieved: 92.5% ✅

- [x] **T1.5** Fix Makefile GOROOT issue
  - Verified: `make test` works fine, GOROOT issue was already resolved ✅

---

### Priority 2 — New Resources (High-Value API Coverage)

- [x] **T2.1** `veeam_general_options` — Singleton resource (GET/PUT)
  - API: `GET/PUT /api/v1/generalOptions`
  - Manages server-wide general options, email settings, event forwarding, notifications, storage latency
  - Schema: storage latency control, email notifications (SMTP), SNMP notifications, syslog/event forwarding
  - Pattern: GET → merge → PUT singleton (fixed ID `"general-options"`, no-op delete)
  - Tests: 11 unit tests (Configure, Create/Read/Update/Delete/Import) ✅

- [x] **T2.2** `veeam_email_settings` — Singleton resource (GET/PUT)
  - API: `GET/PUT /api/v1/generalOptions/emailSettings`, `POST .../testMessage`
  - Schema: enabled, SMTP server/port/SSL/auth, from/to/subject, send_on_success/warning/error/daily_summary, send_test_message
  - Tests: 12 unit tests ✅

- [x] **T2.3** `veeam_notification_settings` — Singleton resource (GET/PUT)
  - API: `GET/PUT /api/v1/generalOptions/notifications`
  - Schema: 12 bool flags for success/warning/error notifications via email, SNMP, syslog
  - Tests: 11 unit tests ✅

- [x] **T2.4** `veeam_traffic_rules` — Singleton resource (GET/PUT)
  - API: `GET/PUT /api/v1/trafficRules`
  - Schema: throttling_enabled (Bool), throttling_rules (JSON string for rules array)
  - Tests: 13 unit tests ✅

- [x] **T2.5** `veeam_security_settings` — Singleton resource (GET/PUT)
  - API: `GET/PUT /api/v1/security/settings`
  - Schema: require_ssl/mfa, block_first_login, login_attempt_limit, inactivity_timeout_min, password_expiration_days/enabled
  - Tests: 11 unit tests ✅

- [x] **T2.6** `veeam_kms_server` — Full CRUD resource
  - API: `GET/POST /api/v1/kmsServers`, `GET/PUT/DELETE /api/v1/kmsServers/{id}`, `POST .../changeCertificate`
  - Schema: name, description, hostname, port, certificate_thumbprint
  - Tests: 16 unit tests ✅

- [x] **T2.7** `veeam_security_user` — Partial CRUD resource (Create/Read/Delete)
  - API: `GET/POST /api/v1/security/users`, `GET/DELETE /api/v1/security/users/{id}`, `GET/PUT .../roles`
  - Schema: login, password (Sensitive), description, role — RequiresReplace on login+role
  - Tests: 12 unit tests ✅

- [x] **T2.8** `veeam_ad_domain` — Partial CRUD resource (Create/Read/Delete)
  - API: `GET/POST /api/v1/adDomains`, `GET/DELETE /api/v1/adDomains/{id}`
  - Schema: name, username, password (Sensitive), description — RequiresReplace on name+username
  - Tests: 11 unit tests ✅

---

### Priority 3 — New Data Sources (Read-Only Visibility)

- [x] **T3.1** `veeam_security_roles` — Read-only
  - API: `GET /api/v1/security/roles`
  - List available RBAC roles and permissions

- [x] **T3.2** `veeam_security_users` — Read-only
  - API: `GET /api/v1/security/users`
  - List configured RBAC users

- [x] **T3.3** `veeam_backup_objects` — Read-only
  - API: `GET /api/v1/backupObjects`
  - List objects within backups (VMs, machines)

- [x] **T3.4** `veeam_replicas` — Read-only
  - API: `GET /api/v1/replicas`
  - List VM replicas

- [x] **T3.5** `veeam_replica_points` — Read-only
  - API: `GET /api/v1/replicaPoints`
  - List replica restore points

- [x] **T3.6** `veeam_proxy_states` — Read-only
  - API: `GET /api/v1/backupInfrastructure/proxies/states`
  - Proxy health/state information

- [x] **T3.7** `veeam_protected_computers` — Read-only
  - API: `GET /api/v1/agents/protectedComputers`
  - List agent-protected computers

- [x] **T3.8** `veeam_services` — Read-only
  - API: `GET /api/v1/services`
  - List VBR services and their status

- [x] **T3.9** `veeam_server_time` — Read-only
  - API: `GET /api/v1/serverTime`
  - Server time (useful for schedule validation)

- [x] **T3.10** `veeam_server_certificate` — Read-only
  - API: `GET /api/v1/serverCertificate`
  - Server TLS certificate details

- [x] **T3.11** `veeam_task_sessions` — Read-only
  - API: `GET /api/v1/taskSessions`
  - Granular task-level session details

- [x] **T3.12** `veeam_security_analyzer` — Read-only
  - API: `GET /api/v1/securityAnalyzer/bestPractices`, `GET /api/v1/securityAnalyzer/lastRun`
  - Security compliance best practices check results

- [x] **T3.13** `veeam_malware_events` — Read-only
  - API: `GET /api/v1/malwareDetection/events`
  - Malware detection events

---

### Priority 4 — Job Type Expansion

- [ ] **T4.1** Backup job: `HyperVBackup` support *(not planned — vSphere-only policy)*
  - Skipped: provider targets VMware vSphere environments only

- [ ] **T4.2** Backup job: `BackupCopy` support *(not planned — vSphere-only policy)*
  - Skipped: distinct schema, out of current scope

- [ ] **T4.3** Backup job: `VSphereReplica` support *(not planned — vSphere-only policy)*
  - Skipped: distinct schema, out of current scope

- [x] **T4.4** Backup job: `WindowsAgentBackup` (and `LinuxAgentBackup`) support
  - Fully implemented in `veeam_backup_job` resource
  - Schema: `agent_computers`, `agent_backup_mode`, `agent_type`, `include_usb_drives`, `volumes_scope`, `files_scope`
  - Tests: covered in `backup_job_test.go` ✅

- [x] **T4.5** Protection group: `ADObjects` type support
  - API models added: `ADObjectsProtectionGroupSpec/Model`, `ADObjectsAccount`, `ADObject`
  - Resource schema: `ad_account` block + `ad_objects` list
  - Full CRUD in `veeam_protection_group` resource ✅

- [x] **T4.6** Protection group: `CSVFile` type support
  - API models added: `CSVFileProtectionGroupSpec/Model`, `ECSVDelimiterType`
  - Resource schema: `csv_file_path` + `csv_delimiter_type`
  - Full CRUD in `veeam_protection_group` resource ✅

---

### Priority 5 — Infrastructure & CI/CD

- [x] **T5.1** CI pipeline setup (GitHub Actions)
  - Added `.github/workflows/ci.yml`: fmt-check, vet, golangci-lint, unit tests on every PR/push to master
  - Build matrix job: linux/amd64, linux/arm64, darwin/arm64, windows/amd64 — runs after lint+test gate
  - Acceptance test job intentionally omitted (requires live VBR — see T5.4)
  - Uses `go-version-file: go.mod` so CI tracks the authoritative Go version automatically ✅

- [x] **T5.2** GoReleaser configuration
  - `.goreleaser.yml` fully configured: multi-platform builds, GPG-conditional signing, changelog groups ✅
  - `release.yml` handles both signed (GPG_PRIVATE_KEY secret present) and unsigned release paths ✅
  - Changelog groups follow conventional commits (feat/fix/docs/others) ✅
  - NOTE: `release.yml` pins `go-version: '1.24.4'`; `go.mod` declares `go 1.26.1` — Go's GOTOOLCHAIN=auto
    resolves this at runtime but aligning these explicitly is advisable before the first public release.

- [x] **T5.3** Terraform Registry publishing preparation (review only — not publishing yet)
  - `terraform-registry-manifest.json` exists with correct `protocol_versions: ["6.0"]` ✅
  - Fixed: `terraform-registry-manifest.json` now included in each platform zip (required by Registry) ✅
  - Binary naming follows registry convention: `terraform-provider-veeam_vVERSION` ✅
  - SHA256SUMS + SHA256SUMS.sig produced by GoReleaser ✅
  - Remaining before publishing: (1) claim `patrikcze` namespace on registry.terraform.io,
    (2) register GPG public key in the Registry UI, (3) push a signed tag

- [ ] **T5.4** Acceptance test environment automation — **POSTPONED**
  - VBR requires Windows Server + SQL Server; not containerizable in CI
  - `scripts/setup-ubuntu-test-env.sh` exists for developer machine setup
  - Live validation workflow documented in `TESTING.md` (real VBR scenarios covered manually)
  - Revisit if a shared VBR test instance becomes available

---

### Priority 6 — Advanced Features & Polish

- [x] **T6.1** `veeam_mount_server` — Partial CRUD resource (Create/Read/Update, no-op Delete)
  - API: `GET/POST /api/v1/backupInfrastructure/mountServers`, `GET/PUT /api/v1/backupInfrastructure/mountServers/{id}`
  - No delete endpoint — mount server lifecycle tied to managed server; Delete is a no-op

- [x] **T6.2** `veeam_global_vm_exclusion` — Partial CRUD resource (Create/Read/Delete, Update is pass-through)
  - API: `GET/POST /api/v1/globalExclusions/vm`, `GET/DELETE /api/v1/globalExclusions/vm/{id}`
  - All key fields have RequiresReplace; Update method is a no-op pass-through

- [x] **T6.3** `veeam_recovery_token` — Full CRUD resource
  - API: `GET/POST /api/v1/agents/recoveryTokens`, `GET/PUT/DELETE /api/v1/agents/recoveryTokens/{id}`
  - `token_value` is Sensitive and only captured on Create; preserved in state on subsequent reads

- [x] **T6.4** `veeam_entra_id_tenant` — Full CRUD resource
  - API: `GET/POST /api/v1/inventory/entraId/tenants`, `GET/PUT/DELETE /api/v1/inventory/entraId/tenants/{id}`
  - `tenant_id` carries RequiresReplace

- [x] **T6.5** `veeam_unstructured_data_server` — Full CRUD resource
  - API: `GET/POST /api/v1/inventory/unstructuredDataServers`, `GET/PUT/DELETE /api/v1/inventory/unstructuredDataServers/{id}`
  - `type` carries RequiresReplace

- [x] **T6.6** `veeam_security_analyzer_schedule` — Singleton resource (GET → merge → PUT)
  - API: `GET/PUT /api/v1/securityAnalyzer/schedule`
  - Fixed ID `security-analyzer-schedule`; no-op Delete; typed struct via `internal/models/security_analyzer.go`

- [x] **T6.7** `veeam_event_forwarding` — Singleton resource (GET → merge → PUT)
  - API: `GET/PUT /api/v1/generalOptions/eventForwarding`
  - Fixed ID `event-forwarding`; map-based merge pattern; no-op Delete

- [x] **T6.8** `veeam_storage_latency` — Singleton resource (GET → merge → PUT)
  - API: `GET/PUT /api/v1/generalOptions/storageLatency`
  - Fixed ID `storage-latency`; map-based merge pattern; no-op Delete

---

### Priority 7 — Documentation & Examples

- [x] **T7.1** Complete example configurations for all resource types
  - `examples/complete/main.tf` exists — verified all resources shown; P6 section appended
  - Added per-resource isolated examples in `examples/resources/` for all 19 missing resources

- [x] **T7.2** Data source usage examples
  - Added 14 example files in `examples/data-sources/` with list, by-ID, and cross-reference patterns

- [x] **T7.3** Import guide
  - `docs/guides/import.md` — covers all resources with exact import commands and UUID discovery tips

- [x] **T7.4** Upgrade guide & migration notes
  - `docs/guides/upgrade.md` — versioning policy, V13→V14 two-file migration, RequiresReplace table, sensitive fields table, singleton list

- [ ] **T7.5** Regenerate docs via `tfplugindocs`
  - (skipped — manual doc management preferred, tfplugindocs not used)

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
