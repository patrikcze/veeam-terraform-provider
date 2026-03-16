# Veeam Terraform Provider

This Terraform provider allows you to manage Veeam Backup & Replication resources using Terraform. The provider enables declarative configuration management of Veeam environments, leveraging the capabilities of HashiCorp Terraform.

## Features

- **Backup Job Management**: Create, update, and delete backup jobs with full storage/schedule/retention configuration
- **Repository Management**: Polymorphic repositories — WinLocal, LinuxLocal, Nfs, Smb
- **Credential Management**: Standard (Windows/domain) and Linux SSH credentials
- **Managed Server Management**: ViHost, WindowsHost, LinuxHost managed servers
- **Proxy Management**: vSphere backup proxies with transport mode configuration
- **Encryption Passwords**: Manage encryption keys for backup encryption
- **Protection Groups**: Agent-based protection groups (IndividualComputers)
- **Data Sources**: Query credentials, backup jobs, and repositories
- **Import Support**: All resources support `terraform import`
- **Version-Resilient Architecture**: All API paths centralized in `internal/client/endpoints.go` — update one file when Veeam releases a new API version

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.24 (for building from source)
- Veeam Backup & Replication V13 server
- Network access to Veeam server REST API (port 9419)

## Installation

### From Terraform Registry

```hcl
terraform {
  required_providers {
    veeam = {
      source = "patrikcze/veeam"
      version = "~> 1.0"
    }
  }
}
```

### Building from Source

```bash
# Clone the repository
git clone https://github.com/patrikcze/terraform-provider-veeam.git
cd terraform-provider-veeam

# Build the provider
make build

# Install locally for development
make install
```

## Quick Start

### Basic Provider Configuration

```hcl
provider "veeam" {
  host     = "veeam.example.com"
  username = "admin"
  password = var.veeam_password  # Use variables for sensitive data
  insecure = false               # Set to true for self-signed certificates
}
```

### Environment Variables

For security, use environment variables instead of hardcoding credentials:

```bash
export VEEAM_HOST="veeam.example.com"
export VEEAM_USERNAME="admin"
export VEEAM_PASSWORD="your-password"
export VEEAM_INSECURE="false"
```

Then configure the provider:

```hcl
provider "veeam" {
  # Configuration will be read from environment variables
}
```

### Simple Example

```hcl
# Create a backup repository
resource "veeam_repository" "primary" {
  name        = "Primary-Backup-Repo"
  description = "Primary backup repository"
  path        = "/backup/primary"
  type        = "linux"
  capacity    = 10737418240  # 10GB in bytes
}

# Create a backup job
resource "veeam_backup_job" "daily" {
  name    = "Daily-VM-Backup"
  enabled = true
}

# Query existing backup jobs
data "veeam_backup_jobs" "all" {}

output "backup_jobs" {
  value = data.veeam_backup_jobs.all.backup_jobs
}
```

## Provider Configuration

The provider supports the following configuration options:

| Argument   | Type   | Required | Description |
|------------|--------|----------|-------------|
| `host`     | string | Yes      | Veeam Backup & Replication server hostname or IP address |
| `username` | string | Yes      | Username for authentication to the Veeam server |
| `password` | string | Yes      | Password for authentication to the Veeam server |
| `insecure` | bool   | No       | Skip TLS certificate verification (default: false) |

## Resources

- `veeam_backup_job` — Backup jobs with storage, schedule, retention, proxy settings
- `veeam_cloud_credential` — Cloud credentials for AWS/Azure/GCP integrations
- `veeam_configuration_backup` — Configuration backup settings and trigger
- `veeam_credential` — Standard (Windows/domain) and Linux SSH credentials
- `veeam_encryption_password` — Encryption passwords for backup encryption
- `veeam_managed_server` — ViHost, WindowsHost, LinuxHost managed servers
- `veeam_protection_group` — Agent protection groups (IndividualComputers)
- `veeam_proxy` — vSphere backup proxies
- `veeam_repository` — Backup repositories (WinLocal, LinuxLocal, Nfs, Smb)
- `veeam_scale_out_repository` — Scale-out backup repositories (SOBR)

## Data Sources

- `veeam_backups` — Query backups and optional backup files
- `veeam_backup_jobs` — Query backup jobs (all or by ID/name)
- `veeam_credentials` — List all credentials
- `veeam_job_states` — Aggregated job state overview
- `veeam_license` — Installed license and consumption summary
- `veeam_managed_servers` — Query managed servers
- `veeam_protection_groups` — Query protection groups
- `veeam_proxies` — Query backup proxies
- `veeam_repository_states` — Repository capacity and status
- `veeam_repositories` — Query repositories (all or by ID/name)
- `veeam_restore_points` — Query restore points
- `veeam_server_info` — Query backup server info
- `veeam_sessions` — Query session history and status
- `veeam_wan_accelerators` — Query WAN accelerators

## Documentation

Detailed documentation for each resource and data source is available in the `/docs` directory:

- [Resources Documentation](docs/resources/index.md)
- [Data Sources Documentation](docs/data-sources/index.md)
- [Examples](examples/)

## Examples

Comprehensive examples are available in the `/examples` directory:

- [Basic Setup](examples/basic/)
- [Advanced Configuration](examples/advanced/)
- [Data Source Usage](examples/data-sources/)

## Development

### Building

```bash
make build
```

### Testing

```bash
make test
```

### Linting

```bash
make lint
```

### Formatting

```bash
make fmt
```

### Vendoring Dependencies

```bash
make vendor
```

## Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

Please make sure to run tests and linting before submitting:

```bash
make check
```

## License

This project is licensed under the MPL-2.0 License - see the [LICENSE](LICENSE) file for details.


