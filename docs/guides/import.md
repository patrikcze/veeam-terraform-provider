---
page_title: "Import Guide - terraform-provider-veeam"
subcategory: "Guides"
description: |-
  How to import existing Veeam Backup & Replication resources into Terraform state.
---

# Import Guide

All resources in this provider support `terraform import`. Use this guide to bring existing Veeam Backup & Replication configuration under Terraform management without destroying and recreating it.

## General Usage

```bash
terraform import <resource_type>.<resource_name> <import_id>
```

After importing, run `terraform plan` to see whether the current server configuration matches your Terraform configuration. Add any missing attributes until the plan shows no changes.

---

## Standard Resources (UUID Import ID)

These resources use the Veeam-assigned UUID as the import ID. The UUID can be found in the Veeam console or via the REST API endpoint listed.

### veeam_credential

```bash
terraform import veeam_credential.main <uuid>
```

Find the UUID via `GET /api/v1/credentials` or in the Veeam console under **Manage Credentials**.

Note: `password`, `private_key`, and `passphrase` are write-only fields — they will not be populated after import and must be set manually in your configuration.

### veeam_managed_server

```bash
terraform import veeam_managed_server.main <uuid>
```

Find the UUID via `GET /api/v1/backupInfrastructure/managedServers` or in the Veeam console under **Backup Infrastructure > Managed Servers**.

### veeam_repository

```bash
terraform import veeam_repository.main <uuid>
```

Find the UUID via `GET /api/v1/backupInfrastructure/repositories` or in the Veeam console under **Backup Infrastructure > Backup Repositories**.

### veeam_scale_out_repository

```bash
terraform import veeam_scale_out_repository.main <uuid>
```

Find the UUID via `GET /api/v1/backupInfrastructure/scaleOutRepositories`.

### veeam_proxy

```bash
terraform import veeam_proxy.main <uuid>
```

Find the UUID via `GET /api/v1/backupInfrastructure/proxies`.

### veeam_cloud_credential

```bash
terraform import veeam_cloud_credential.main <uuid>
```

Find the UUID via `GET /api/v1/cloudCredentials`.

### veeam_encryption_password

```bash
terraform import veeam_encryption_password.main <uuid>
```

Find the UUID via `GET /api/v1/encryptionPasswords`.

Note: `password` is write-only and will not be populated after import.

### veeam_backup_job

```bash
terraform import veeam_backup_job.main <uuid>
```

Find the UUID via `GET /api/v1/jobs` or in the Veeam console under **Home > Jobs**.

### veeam_protection_group

```bash
terraform import veeam_protection_group.main <uuid>
```

Find the UUID via `GET /api/v1/protectionGroups`.

### veeam_ad_domain

```bash
terraform import veeam_ad_domain.main <uuid>
```

Find the UUID via `GET /api/v1/activeDirectory`.

### veeam_kms_server

```bash
terraform import veeam_kms_server.main <uuid>
```

Find the UUID via `GET /api/v1/kmsServers`.

### veeam_security_user

```bash
terraform import veeam_security_user.main <uuid>
```

Find the UUID via `GET /api/v1/security/users`.

Note: `password` is write-only and will not be populated after import. `login` and `role` carry `RequiresReplace` — ensure they match exactly to avoid unintended destroy and recreate.

### veeam_entra_id_tenant

```bash
terraform import veeam_entra_id_tenant.main <uuid>
```

Find the UUID via `GET /api/v1/inventory/entraId/tenants`.

### veeam_global_vm_exclusion

```bash
terraform import veeam_global_vm_exclusion.main <uuid>
```

Find the UUID via `GET /api/v1/globalExclusions/vm`.

### veeam_mount_server

```bash
terraform import veeam_mount_server.main <uuid>
```

Find the UUID via `GET /api/v1/backupInfrastructure/mountServers`.

### veeam_recovery_token

```bash
terraform import veeam_recovery_token.main <uuid>
```

Find the UUID via `GET /api/v1/agents/recoveryTokens`.

Note: `token_value` is only available at creation time. After import, `token_value` will be empty in Terraform state. The token itself remains valid on the server.

### veeam_unstructured_data_server

```bash
terraform import veeam_unstructured_data_server.main <uuid>
```

Find the UUID via `GET /api/v1/inventory/unstructuredDataServers`.

---

## Singleton Resources (Fixed Import ID)

These resources represent server-level configuration objects that always exist. They use a fixed string as the import ID instead of a UUID.

### veeam_general_options

```bash
terraform import veeam_general_options.main general-options
```

### veeam_email_settings

```bash
terraform import veeam_email_settings.main email-settings
```

### veeam_notification_settings

```bash
terraform import veeam_notification_settings.main notification-settings
```

### veeam_security_settings

```bash
terraform import veeam_security_settings.main security-settings
```

### veeam_traffic_rules

```bash
terraform import veeam_traffic_rules.main traffic-rules
```

### veeam_configuration_backup

```bash
terraform import veeam_configuration_backup.main configuration-backup
```

### veeam_event_forwarding

```bash
terraform import veeam_event_forwarding.main event-forwarding
```

### veeam_security_analyzer_schedule

```bash
terraform import veeam_security_analyzer_schedule.main security-analyzer-schedule
```

### veeam_storage_latency

```bash
terraform import veeam_storage_latency.main storage-latency
```

---

## After Importing

1. Run `terraform plan` to compare the imported state against your configuration.
2. Add or adjust attributes in your configuration until the plan shows no changes.
3. For write-only fields (passwords, token values), set them in your configuration — Terraform will detect a diff but the API will not return an error if the value matches the current server state.
4. Commit the updated state and configuration together.
