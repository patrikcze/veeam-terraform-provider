---
page_title: "Upgrade Guide - terraform-provider-veeam"
subcategory: "Guides"
description: |-
  How to upgrade the provider and migrate between Veeam REST API versions.
---

# Upgrade Guide

This guide explains the provider versioning policy, how to upgrade the API version the provider targets, and lists the fields that trigger resource replacement or are treated as sensitive.

---

## Provider Versioning

This provider follows [Semantic Versioning](https://semver.org/):

- **Patch** (x.y.Z) — bug fixes, documentation updates. No schema changes.
- **Minor** (x.Y.0) — new resources, data sources, or optional attributes. Backwards compatible.
- **Major** (X.0.0) — breaking schema changes, field removals, or provider configuration changes. A migration guide will accompany these releases.

Pin to a minor version range to receive patches automatically while avoiding major breaking changes:

```hcl
terraform {
  required_providers {
    veeam = {
      source  = "patrikcze/veeam"
      version = "~> 1.0"
    }
  }
}
```

---

## Upgrading the Target API Version (V13 → V14)

When Veeam releases a new REST API version, the provider can be updated to target it by changing exactly two files:

### 1. Update the API version constant

File: `internal/client/client.go`

Find the `APIVersion` constant and update it to the new version string:

```go
// Before
const APIVersion = "1.3-rev1"

// After
const APIVersion = "1.4"
```

### 2. Update endpoint path constants

File: `internal/client/endpoints.go`

Review any paths that include the API version in the URL segment (e.g. `/api/v1/...` → `/api/v2/...`) and update the constants accordingly.

All resource and data source implementations reference these path constants — no other files need to change for a pure API version bump.

---

## V13 → V14 Migration Checklist

When migrating to a provider version targeting the Veeam V14 API:

1. **Read the V14 Swagger spec** — compare it against the V13 spec for field renames, added required fields, removed fields, or changed enum values.
2. **Update `endpoints.go`** — update path constants for any endpoints whose URL changed.
3. **Update `client.go`** — update the `APIVersion` constant.
4. **Check model structs** in `internal/models/` — update JSON tags for any renamed fields.
5. **Run `make build`** — ensure the provider compiles cleanly.
6. **Run `make test`** — verify unit tests pass.
7. **Run acceptance tests** against a V14 server — `make testacc`.
8. **Inspect `terraform plan`** on an existing state — verify no unexpected drift.

---

## Fields That Trigger Destroy and Recreate (RequiresReplace)

Changing any of the following fields will cause Terraform to destroy and recreate the resource.

| Resource | Field(s) |
|---|---|
| `veeam_ad_domain` | `name`, `username` |
| `veeam_backup_job` | `type` |
| `veeam_entra_id_tenant` | `tenant_id` |
| `veeam_global_vm_exclusion` | `name`, `type`, `host_name`, `object_id` |
| `veeam_mount_server` | `managed_server_id`, `type` |
| `veeam_recovery_token` | `managed_server_id` |
| `veeam_security_user` | `login`, `role` |
| `veeam_unstructured_data_server` | `type` |

For `veeam_protection_group`, changing the group `type` (e.g. `IndividualComputers` → `ADObjects`) also triggers a destroy and recreate because the API does not support in-place type conversion.

---

## Sensitive Fields

The following fields are marked `Sensitive: true` in the provider schema. Their values are redacted from Terraform plan and apply output and are stored in the state file as-is (not encrypted unless you use a state backend with encryption).

| Resource | Field(s) |
|---|---|
| `veeam_ad_domain` | `password` |
| `veeam_cloud_credential` | `access_key`, `shared_key`, `secret_key`, `tenant_id`, `application_id`, `application_key`, `project_id`, `service_account` |
| `veeam_credential` | `password`, `private_key`, `passphrase` |
| `veeam_email_settings` | `smtp_password` |
| `veeam_encryption_password` | `password` |
| `veeam_recovery_token` | `token_value` |
| `veeam_security_user` | `password` |

All sensitive fields are **write-only** — the Veeam REST API never returns these values in read responses. The provider preserves the value from the plan in state and does not overwrite it from the API response.

---

## State Format Compatibility

Terraform state format is stable within a major provider version. Minor and patch upgrades do not change how resources are represented in state.

If a major upgrade includes state-breaking changes, a `terraform state mv` or `terraform import` workflow will be documented in the release notes accompanying that version.

---

## Singleton Resources and State

The following resources are singletons — they represent server-level configuration objects that always exist and cannot be deleted via the API. Deleting these resources in Terraform only removes the state entry; the server configuration is unchanged.

| Resource | Fixed Import ID |
|---|---|
| `veeam_general_options` | `general-options` |
| `veeam_email_settings` | `email-settings` |
| `veeam_notification_settings` | `notification-settings` |
| `veeam_security_settings` | `security-settings` |
| `veeam_traffic_rules` | `traffic-rules` |
| `veeam_configuration_backup` | `configuration-backup` |
| `veeam_event_forwarding` | `event-forwarding` |
| `veeam_security_analyzer_schedule` | `security-analyzer-schedule` |
| `veeam_storage_latency` | `storage-latency` |

If a singleton resource is removed from Terraform state (e.g., through `terraform state rm`), it can be re-imported using its fixed ID without any server-side impact.
