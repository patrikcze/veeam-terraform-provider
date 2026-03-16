---
page_title: "veeam_backup_job Resource - terraform-provider-veeam"
subcategory: ""
description: |-
  Manages a Veeam backup job.
---

# veeam_backup_job (Resource)

Creates, updates, reads, and deletes backup jobs in Veeam Backup & Replication.

## Example Usage

```hcl
resource "veeam_backup_job" "daily" {
  name    = "Daily-VM-Backup"
  enabled = true
}
```

## Schema

### Required

- `name` (String) Unique backup job name.

### Optional

- `enabled` (Boolean) Whether the job is enabled. Defaults to `true`.

### Read-Only

- `id` (String) Backup job identifier assigned by Veeam.

## Import

Import by job name:

```bash
terraform import veeam_backup_job.example "Daily-VM-Backup"
```

## Notes

- Job names must be unique in the Veeam environment.
- Disabling a job keeps it configured but prevents scheduled runs.
