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
