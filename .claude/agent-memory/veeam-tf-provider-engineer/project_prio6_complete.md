---
name: Priority 6 resources complete
description: All 8 T6.1–T6.8 resources implemented, tested, and registered as of 2026-04-02
type: project
---

All 8 Priority 6 resources are implemented, tested, and passing `make check`.

**Why:** Provider coverage expansion milestone; resources cover mount servers, global VM exclusions, recovery tokens, Entra ID tenants, unstructured data servers, and three singleton config resources.

**How to apply:** These are complete — no further T6 work needed. Next milestone is Priority 7 (Documentation & Examples).

## Resources implemented

| Task | Resource | Pattern | Notes |
|------|----------|---------|-------|
| T6.1 | `veeam_mount_server` | Partial CRUD (no Delete) | Delete is no-op; `managed_server_id` and `type` are RequiresReplace |
| T6.2 | `veeam_global_vm_exclusion` | Partial CRUD (no Update) | Update is pass-through; all key fields RequiresReplace |
| T6.3 | `veeam_recovery_token` | Full CRUD | `token_value` Sensitive+Computed; only captured on Create, preserved in state |
| T6.4 | `veeam_entra_id_tenant` | Full CRUD | `tenant_id` RequiresReplace |
| T6.5 | `veeam_unstructured_data_server` | Full CRUD | `type` RequiresReplace |
| T6.6 | `veeam_security_analyzer_schedule` | Singleton GET→merge→PUT | Fixed ID `security-analyzer-schedule`; typed struct `models.SecurityAnalyzerScheduleModel` |
| T6.7 | `veeam_event_forwarding` | Singleton GET→merge→PUT | Fixed ID `event-forwarding`; map-based merge; nested `snmp` / `syslog` keys |
| T6.8 | `veeam_storage_latency` | Singleton GET→merge→PUT | Fixed ID `storage-latency`; nested `throttlingIo` / `stopJobs` keys |

## New model files
- `internal/models/mount_servers.go`
- `internal/models/global_exclusions.go`
- `internal/models/recovery_tokens.go`
- `internal/models/entra_id.go`
- `internal/models/unstructured_data.go`
- `internal/models/security_analyzer.go`

## New endpoint constants (internal/client/endpoints.go)
`PathMountServers`, `PathMountServerByID`, `PathGlobalVMExclusions`, `PathGlobalVMExclusionByID`, `PathRecoveryTokens`, `PathRecoveryTokenByID`, `PathEntraIDTenants`, `PathEntraIDTenantByID`, `PathUnstructuredDataServers`, `PathUnstructuredDataServerByID`, `PathSecurityAnalyzerSchedule`, `PathEventForwarding`, `PathStorageLatency`
