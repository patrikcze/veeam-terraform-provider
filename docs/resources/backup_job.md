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
  name          = "Daily-VM-Backup"
  type          = "VSphereBackup"
  repository_id = veeam_repository.primary.id
}
```

## Schema

### Required

- `name` (String) Unique backup job name.
- `type` (String) Job type: `VSphereBackup`, `BackupCopy`, `VSphereReplica`, etc.
- `repository_id` (String) Target backup repository ID.

### Optional

- `description` (String) Optional description.
- `is_high_priority` (Boolean) Process this job before lower-priority jobs.
- `proxy_auto_select` (Boolean) Automatically select backup proxy.
- `retention_type` (String) Retention type: `RestorePoints` or `Days`.
- `retention_quantity` (Number) Number of restore points or days to keep.
- `schedule_enabled` (Boolean) Run the job automatically on schedule.
- `schedule_time` (String) Daily schedule time (e.g. `22:00`).
- `schedule_kind` (String) Daily schedule kind: `Everyday`, `Weekdays`, or `SelectedDays`.
- `retry_enabled` (Boolean) Retry on failure.
- `retry_count` (Number) Number of retry attempts.
- `retry_await_minutes` (Number) Minutes to wait between retries.

### Read-Only

- `id` (String) Backup job identifier assigned by Veeam.
- `is_disabled` (Boolean) Whether the job is currently disabled.

## Import

Import by job ID:

```bash
terraform import veeam_backup_job.example <job-id>
```

## Notes

- Job names must be unique within the Veeam environment.
- The `type` value must match the Veeam REST API enum. Common values: `VSphereBackup`, `BackupCopy`, `VSphereReplica`, `HyperVBackup`, `WindowsAgentBackup`, `LinuxAgentBackup`.
