# veeam_backup_job

Manages a Veeam Backup Job. This resource allows you to create, update, and delete backup jobs in Veeam Backup & Replication.

## Example Usage

```hcl
# Basic backup job
resource "veeam_backup_job" "example" {
  name    = "Daily-VM-Backup"
  enabled = true
}

# Backup job with custom settings
resource "veeam_backup_job" "advanced" {
  name    = "Weekly-Full-Backup"
  enabled = false
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name of the backup job. Must be unique within the Veeam environment.
- `enabled` - (Optional) Whether the backup job is enabled. Defaults to `true`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The unique identifier of the backup job.

## Import

Backup jobs can be imported using their name:

```bash
terraform import veeam_backup_job.example "Daily-VM-Backup"
```

## Notes

- Backup job names must be unique within the Veeam environment
- Disabling a backup job will prevent it from running on schedule
- Deleting a backup job will remove it from Veeam but will not affect existing backup files
