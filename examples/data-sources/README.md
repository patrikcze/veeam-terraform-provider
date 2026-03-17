# Data Sources Example

This example demonstrates a health-check style configuration that queries all implemented Veeam data sources and prints summary outputs.

## Features Demonstrated

- Querying all implemented data sources from the provider
- Printing counts and basic server/license details
- Running a quick connectivity/contract validation against VBR

## Prerequisites

- Terraform >= 1.0
- Access to a Veeam Backup & Replication server
- Existing backup jobs and repositories in your Veeam environment

## Usage

1. **Set up variables**:

```hcl
veeam_host     = "your-veeam-server.com"
veeam_username = "admin"
veeam_password = "your-password"
veeam_insecure = false
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

The script in `main.tf` reads all currently implemented data sources:

- `veeam_server_info`
- `veeam_license`
- `veeam_credentials`
- `veeam_managed_servers`
- `veeam_proxies`
- `veeam_repositories`
- `veeam_repository_states`
- `veeam_backup_jobs`
- `veeam_job_states`
- `veeam_backups`
- `veeam_restore_points`
- `veeam_sessions`
- `veeam_protection_groups`
- `veeam_wan_accelerators`

## Example Outputs

The configuration provides these outputs:

- `datasource_counts`: per-data-source item counts
- `server_info`: server name/version/build
- `license_summary`: key license fields

## Use Cases

1. Connectivity check against VBR API.
2. Quick validation of provider/data-source compatibility.
3. Inventory snapshot for troubleshooting.

## Notes

- Data sources are read-only and don't modify existing resources.
- If one data source fails, it indicates contract drift for that specific endpoint/parser.
