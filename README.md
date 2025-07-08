# Veeam Terraform Provider

This Terraform provider allows you to manage Veeam Backup & Replication resources using Terraform. The provider enables declarative configuration management of Veeam environments, leveraging the capabilities of HashiCorp Terraform.

## Features

- **Backup Job Management**: Create, update, and delete backup jobs
- **Repository Management**: Configure and manage backup repositories
- **Credential Management**: Manage authentication credentials for various systems
- **Data Sources**: Query existing backup jobs and repositories
- **Import Support**: Import existing resources into Terraform state

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21 (for building from source)
- Veeam Backup & Replication server (version 11.0 or later)
- Network access to Veeam server REST API (typically port 9419)

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

### Backup Jobs
- `veeam_backup_job` - Manages backup jobs

### Repositories
- `veeam_repository` - Manages backup repositories

### Credentials
- `veeam_credential` - Manages authentication credentials

## Data Sources

### Backup Jobs
- `veeam_backup_jobs` - Query backup jobs

### Repositories
- `veeam_repositories` - Query backup repositories

## Documentation

Detailed documentation for each resource and data source is available in the `/docs` directory:

- [Resources Documentation](docs/resources/)
- [Data Sources Documentation](docs/data-sources/)
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


