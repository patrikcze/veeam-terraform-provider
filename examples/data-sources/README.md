# Data Sources Example

This example demonstrates how to use Veeam data sources to query existing resources and use them in your Terraform configuration.

## Features Demonstrated

- Querying all backup jobs and repositories
- Filtering specific resources by name or ID
- Using data sources to create conditional resources
- Complex data processing with locals
- Comprehensive output generation

## Prerequisites

- Terraform >= 1.0
- Access to a Veeam Backup & Replication server
- Existing backup jobs and repositories in your Veeam environment

## Usage

1. **Set up variables**:

```hcl
veeam_host        = "your-veeam-server.com"
veeam_username    = "admin"
veeam_password    = "your-password"
veeam_insecure    = false
backup_job_id     = "optional-job-id"
repository_id     = "optional-repo-id"
```

2. **Initialize and apply**:

```bash
terraform init
terraform plan
terraform apply
```

3. **View comprehensive outputs**:

```bash
terraform output
```

## Data Sources Used

### Backup Jobs
- `veeam_backup_jobs` - Query all backup jobs
- `veeam_backup_jobs` with `job_name` - Query specific job by name
- `veeam_backup_jobs` with `job_id` - Query specific job by ID

### Repositories
- `veeam_repositories` - Query all repositories
- `veeam_repositories` with `repository_name` - Query specific repository by name
- `veeam_repositories` with `repository_id` - Query specific repository by ID

## Example Outputs

The configuration provides several useful outputs:

- **backup_jobs_summary**: Statistics about all backup jobs
- **repositories_summary**: Statistics about all repositories including capacity utilization
- **specific_resources**: Information about specific resources queried by name
- **repository_details**: Detailed information about each repository
- **backup_job_details**: Detailed information about each backup job

## Use Cases

1. **Monitoring**: Check repository utilization and identify repositories with low free space
2. **Conditional Resources**: Create resources only if certain conditions are met
3. **Resource Discovery**: Find existing resources and use their properties
4. **Reporting**: Generate comprehensive reports about your Veeam environment

## Notes

- Data sources are read-only and don't modify existing resources
- Use filters to query specific resources when you know their names or IDs
- Combine multiple data sources for comprehensive environment analysis
- Use locals for complex data processing and calculations
