---
page_title: "veeam_configuration_backup Resource - terraform-provider-veeam"
subcategory: ""
description: |-
  Manages Veeam configuration backup settings and can trigger configuration backups.
---

# veeam_configuration_backup (Resource)

Manages backup server configuration backup settings and can trigger a backup run during apply.

## Example Usage

```hcl
resource "veeam_encryption_password" "config_key" {
  password = var.encryption_password
  hint     = "Configuration backup key"
}

resource "veeam_configuration_backup" "config" {
  enabled                = true
  repository_id          = veeam_repository.main.id
  restore_points_to_keep = 14
  encryption_enabled     = true
  encryption_password_id = veeam_encryption_password.config_key.id
  trigger_on_apply       = false
}
```

## Schema

### Required

- `enabled` (Boolean) Enables or disables configuration backup.

### Optional

- `repository_id` (String) Repository used to store configuration backups.
- `restore_points_to_keep` (Number) Number of restore points to retain.
- `encryption_enabled` (Boolean) Enables encryption for configuration backups.
- `encryption_password_id` (String) Encryption password ID (from `veeam_encryption_password.<name>.id`, not a credential password string).
- `trigger_on_apply` (Boolean) Triggers an immediate backup run on create/update.

### Read-Only

- `id` (String) Fixed configuration-backup resource ID.
- `last_session_id` (String) Most recent triggered backup session ID.
- `last_session_state` (String) Most recent session state.
- `last_session_result` (String) Most recent session result.

## Import

Import using the configured ID:

```bash
terraform import veeam_configuration_backup.example "config-backup"
```

## Notes

- Deleting this resource disables configuration backup in Veeam.
- The provider reads the current server configuration and updates only Terraform-managed fields to satisfy the V13 `ConfigBackupModel` full-object validation.
- Triggered sessions are best-effort and depend on API response timing.
- When Configuration Backup encryption uses a specific `veeam_encryption_password`, Veeam can temporarily keep that password locked as "in use by Backup Configuration Job" during destroy. This is expected Veeam behavior; retrying destroy shortly after usually succeeds.
