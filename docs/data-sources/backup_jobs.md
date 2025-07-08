# veeam_backup_jobs

Use this data source to query backup jobs from Veeam Backup & Replication. You can retrieve all backup jobs or filter by specific criteria.

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

## Argument Reference

The following arguments are supported:

- `job_id` - (Optional) The ID of a specific backup job to retrieve. Conflicts with `job_name`.
- `job_name` - (Optional) The name of a specific backup job to retrieve. Conflicts with `job_id`.

## Attributes Reference

The following attributes are exported:

- `id` - The identifier of the data source.
- `backup_jobs` - A list of backup jobs. Each backup job has the following attributes:
  - `id` - The unique identifier of the backup job.
  - `name` - The name of the backup job.
  - `enabled` - Whether the backup job is enabled.
  - `description` - The description of the backup job.
  - `repository` - The repository used by the backup job.
  - `schedule` - The schedule configuration of the backup job.
  - `job_type` - The type of the backup job.
  - `created_at` - The timestamp when the backup job was created.
  - `updated_at` - The timestamp when the backup job was last updated.

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

- If no arguments are provided, all backup jobs will be returned
- When using `job_id` or `job_name`, only one backup job will be returned in the list
- The `backup_jobs` attribute is always a list, even when filtering by ID or name
- If no backup jobs match the criteria, an empty list will be returned
- The data source provides read-only access to backup job information
