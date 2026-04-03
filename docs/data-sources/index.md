---
page_title: "Data Sources - terraform-provider-veeam"
subcategory: ""
description: |-
  Index of Terraform data sources supported by the Veeam provider.
---

# Data Sources

Reference for all Veeam Terraform Provider data sources.

## Available Data Sources

### [veeam_backup_jobs](backup_jobs.md)
Query backup jobs (all or filtered by ID/name).

### [veeam_backup_objects](backup_objects.md)
Query backup objects contained within backups.

### [veeam_backups](backups.md)
Query backups and optional backup files.

### [veeam_credentials](credentials.md)
List all saved credentials.

### [veeam_job_states](job_states.md)
Get aggregated job state overview.

### [veeam_license](license.md)
Get installed license and usage summary.

### [veeam_malware_events](malware_events.md)
Query malware detection events.

### [veeam_managed_servers](managed_servers.md)
Query managed servers registered in the Veeam infrastructure.

### [veeam_protected_computers](protected_computers.md)
Query protected computers managed by Veeam agents.

### [veeam_protection_groups](protection_groups.md)
Query protection groups.

### [veeam_proxies](proxies.md)
Query backup proxies.

### [veeam_proxy_states](proxy_states.md)
Get backup proxy operational state.

### [veeam_replica_points](replica_points.md)
Query VM replica restore points.

### [veeam_replicas](replicas.md)
Query VM replicas.

### [veeam_repositories](repositories.md)
Query repositories (all or filtered by ID/name).

### [veeam_repository_states](repository_states.md)
Get repository capacity and health state.

### [veeam_restore_points](restore_points.md)
List restore points globally or per backup object.

### [veeam_security_analyzer](security_analyzer.md)
Query security analyzer scan results.

### [veeam_security_roles](security_roles.md)
List available RBAC security roles.

### [veeam_security_users](security_users.md)
List security user accounts.

### [veeam_server_certificate](server_certificate.md)
Get the TLS certificate installed on the VBR server.

### [veeam_server_info](server_info.md)
Get backup server version and installation details.

### [veeam_server_time](server_time.md)
Get current server date and time.

### [veeam_services](services.md)
List Veeam services running on the server.

### [veeam_sessions](sessions.md)
Get session history and status data.

### [veeam_task_sessions](task_sessions.md)
Query task-level session details.

### [veeam_wan_accelerators](wan_accelerators.md)
Query WAN accelerators.

## Common Patterns

### Query All Resources
Read inventory data for backup jobs and repositories:

```hcl
data "veeam_backup_jobs" "all" {}

data "veeam_repositories" "all" {}

output "job_count" {
  value = length(data.veeam_backup_jobs.all.backup_jobs)
}

output "total_capacity" {
  value = sum([
    for repo in data.veeam_repositories.all.repositories : repo.capacity
  ])
}
```

### Filter Specific Resources
Filter by ID or name when you need one object:

```hcl
data "veeam_backup_jobs" "critical" {
  job_name = "Critical-Production-Backup"
}

data "veeam_repositories" "main" {
  repository_name = "Main-Repository"
}

# Use the data in other resources
resource "veeam_backup_job" "secondary" {
  name    = "Secondary-${data.veeam_backup_jobs.critical.backup_jobs[0].name}"
  enabled = true
}
```

### Conditional Resource Creation
Use data source outputs in resource `count` expressions:

```hcl
data "veeam_repositories" "check_repo" {
  repository_name = "Required-Repository"
}

resource "veeam_backup_job" "conditional" {
  count = length(data.veeam_repositories.check_repo.repositories) > 0 ? 1 : 0
  
  name    = "Conditional-Backup-Job"
  enabled = true
}
```

### Data Processing with Locals
Use locals to shape data for outputs or policy checks:

```hcl
data "veeam_backup_jobs" "all" {}
data "veeam_repositories" "all" {}

locals {
  # Filter enabled backup jobs
  enabled_jobs = [
    for job in data.veeam_backup_jobs.all.backup_jobs : job
    if job.enabled
  ]
  
  # Calculate repository utilization
  repo_utilization = {
    for repo in data.veeam_repositories.all.repositories : repo.name => {
      utilization_percent = repo.capacity > 0 ? (repo.used_space / repo.capacity) * 100 : 0
      free_space_gb      = repo.free_space / 1073741824
      used_space_gb      = repo.used_space / 1073741824
    }
  }
  
  # Find repositories with low free space
  low_space_repos = [
    for repo in data.veeam_repositories.all.repositories : repo
    if repo.capacity > 0 && (repo.free_space / repo.capacity) < 0.1
  ]
}

output "environment_summary" {
  value = {
    total_jobs      = length(data.veeam_backup_jobs.all.backup_jobs)
    enabled_jobs    = length(local.enabled_jobs)
    total_repos     = length(data.veeam_repositories.all.repositories)
    low_space_repos = [for repo in local.low_space_repos : repo.name]
  }
}
```

### Monitoring and Reporting
Create summary outputs for operational visibility:

```hcl
data "veeam_backup_jobs" "all" {}
data "veeam_repositories" "all" {}

output "backup_job_report" {
  value = {
    for job in data.veeam_backup_jobs.all.backup_jobs : job.name => {
      id          = job.id
      enabled     = job.enabled
      description = job.description
      repository  = job.repository
      job_type    = job.job_type
      created_at  = job.created_at
      updated_at  = job.updated_at
    }
  }
}

output "repository_report" {
  value = {
    for repo in data.veeam_repositories.all.repositories : repo.name => {
      id              = repo.id
      path            = repo.path
      type            = repo.type
      capacity_gb     = repo.capacity / 1073741824
      free_space_gb   = repo.free_space / 1073741824
      used_space_gb   = repo.used_space / 1073741824
      utilization_pct = repo.capacity > 0 ? (repo.used_space / repo.capacity) * 100 : 0
      status          = repo.status
    }
  }
}
```

## Best Practices

- Filter by ID/name when possible for predictable plans.
- Handle empty lists before indexing into results.
- Use locals to keep Terraform expressions readable.

## See Also

- [Resources Documentation](../resources/)
- [Examples](../../examples/)
- [Data Sources Examples](../../examples/data-sources/)
