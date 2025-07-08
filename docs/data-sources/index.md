# Data Sources

This section contains documentation for all Veeam Terraform Provider data sources.

## Available Data Sources

### [veeam_backup_jobs](backup_jobs.md)
Query backup jobs from your Veeam environment. Use this data source to retrieve information about existing backup jobs.

**Key features:**
- Query all backup jobs
- Filter by job name or ID
- Access job details like status, schedule, and repository
- Use in conditional resource creation

### [veeam_repositories](repositories.md)
Query repositories from your Veeam environment. Use this data source to retrieve information about existing backup repositories.

**Key features:**
- Query all repositories
- Filter by repository name or ID
- Access repository details like capacity, usage, and status
- Monitor repository utilization

## Common Patterns

### Query All Resources
Get comprehensive information about all resources of a specific type:

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
Query specific resources by name or ID:

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
Create resources only if certain conditions are met:

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
Use locals for complex data manipulation:

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
Generate comprehensive reports about your Veeam environment:

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

1. **Use specific filters**: When you know the resource name or ID, use filters for better performance
2. **Handle empty results**: Always check if the data source returned any results before using them
3. **Use locals for processing**: Complex data manipulation is easier with locals
4. **Combine data sources**: Use multiple data sources together for comprehensive analysis
5. **Cache results**: Data sources are read during plan, so results are cached for the apply phase

## Use Cases

### Environment Discovery
- Inventory existing backup jobs and repositories
- Understand current backup strategy and coverage
- Identify unused or misconfigured resources

### Monitoring and Alerting
- Track repository capacity utilization
- Monitor backup job status and schedules
- Identify performance issues or failures

### Conditional Infrastructure
- Create resources based on existing infrastructure
- Implement backup policies based on current state
- Scale resources based on usage patterns

### Reporting and Compliance
- Generate compliance reports
- Track backup coverage and retention
- Document backup infrastructure

## Error Handling

Common issues and solutions:

- **Resource not found**: Check if the resource exists in Veeam
- **Authentication failed**: Verify provider configuration and credentials
- **API timeout**: Large environments may require increased timeout values
- **Permission denied**: Ensure proper read permissions for your Veeam user

## See Also

- [Resources Documentation](../resources/)
- [Examples](../../examples/)
- [Data Sources Examples](../../examples/data-sources/)
