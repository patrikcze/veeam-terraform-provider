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
# Get all backup jobs
data "veeam_backup_jobs" "all" {}

# Get a specific backup job by ID
data "veeam_backup_jobs" "specific_by_id" {
  job_id = "backup-job-123"
}

# Get a specific backup job by name
data "veeam_backup_jobs" "specific_by_name" {
  job_name = "Daily-VM-Backup"
}

# Use the data in outputs
output "all_backup_jobs" {
  value = data.veeam_backup_jobs.all.backup_jobs
}

output "specific_job_details" {
  value = data.veeam_backup_jobs.specific_by_name.backup_jobs[0]
}

# Use backup job data in a resource
resource "veeam_repository" "job_repository" {
  name = "${data.veeam_backup_jobs.specific_by_name.backup_jobs[0].name}-repository"
  path = "/backup/${data.veeam_backup_jobs.specific_by_name.backup_jobs[0].name}"
  type = "linux"
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

### Filtering and Processing

```hcl
# Get all backup jobs and filter enabled ones
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

### Conditional Resource Creation

```hcl
# Create a repository only if a specific backup job exists
data "veeam_backup_jobs" "check_job" {
  job_name = "Critical-Backup"
}

resource "veeam_repository" "conditional" {
  count = length(data.veeam_backup_jobs.check_job.backup_jobs) > 0 ? 1 : 0
  
  name = "Critical-Backup-Repository"
  path = "/backup/critical"
  type = "linux"
}
```

## Notes

- Without filters, all backup jobs are returned.
- Results are always returned as a list, even when filtering to one item.
