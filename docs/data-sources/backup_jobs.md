---
page_title: "veeam_backup_jobs Data Source - terraform-provider-veeam"
subcategory: ""
description: |-
  Queries backup jobs from Veeam Backup & Replication.
---

# veeam_backup_jobs (Data Source)

Queries backup jobs from Veeam Backup & Replication.

## Example Usage

```hcl
# List all backup jobs
data "veeam_backup_jobs" "all" {}

output "all_backup_job_names" {
  value = [for j in data.veeam_backup_jobs.all.backup_jobs : j.name]
}

# Look up a specific job by name
data "veeam_backup_jobs" "daily" {
  job_name = "Daily-VM-Backup"
}

output "daily_job_id" {
  value = data.veeam_backup_jobs.daily.backup_jobs[0].id
}
```

## Schema

### Optional

- `job_id` (String) Filters to a single job by ID (conflicts with `job_name`).
- `job_name` (String) Filters to a single job by name (conflicts with `job_id`).

### Read-Only

- `id` (String) Data source state identifier.
- `backup_jobs` (List of Object) Backup jobs with key metadata.

## Usage Examples

### Filter enabled jobs

```hcl
data "veeam_backup_jobs" "all" {}

locals {
  enabled_jobs = [
    for job in data.veeam_backup_jobs.all.backup_jobs : job
    if job.enabled
  ]
}

output "enabled_backup_jobs" {
  value = local.enabled_jobs
}
```

## Notes

- Without filters, all backup jobs are returned.
- Results are always returned as a list, even when filtering to a single item.
- Each item exposes: `id`, `name`, `enabled`, `description`, `repository`, `schedule`, `job_type`, `created_at`, `updated_at`.
