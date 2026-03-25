# Veeam Terraform Provider

Terraform provider for [Veeam Backup & Replication V13](https://www.veeam.com/), built on the
HashiCorp Terraform Plugin Framework and the Veeam REST API (v1.3-rev1).

See [CHANGELOG.md](CHANGELOG.md) for release notes.

## Requirements

- Terraform >= 1.6
- Veeam Backup & Replication V13 (REST API port 9419)
- Go >= 1.26 (only for building from source)

## Installation

```hcl
terraform {
  required_providers {
    veeam = {
      source  = "patrikcze/veeam"
      version = "~> 1.0"
    }
  }
}
```

## Provider Configuration

```hcl
provider "veeam" {
  host     = "veeam.example.com"
  username = "administrator"
  password = var.veeam_password
  # port     = 9419   # optional, default 9419
  # insecure = false  # set true only for self-signed certs in lab environments
}
```

All arguments can also be supplied via environment variables:

| Argument   | Environment variable | Default  | Description |
|------------|----------------------|----------|-------------|
| `host`     | `VEEAM_HOST`         | —        | VBR server hostname or IP |
| `port`     | `VEEAM_PORT`         | `9419`   | REST API port |
| `username` | `VEEAM_USERNAME`     | —        | Login username |
| `password` | `VEEAM_PASSWORD`     | —        | Login password |
| `insecure` | `VEEAM_INSECURE`     | `false`  | Skip TLS verification |

```bash
export VEEAM_HOST="veeam.example.com"
export VEEAM_USERNAME="administrator"
export VEEAM_PASSWORD="secret"
```

```hcl
provider "veeam" {} # reads everything from env vars
```

## Real-World Examples

### Linux backup server + encrypted backup job

This is the most common flow: register a Linux host, create a local repository on it,
add an encryption password, and configure a daily backup job.

```hcl
# 1. Credential for the Linux host
resource "veeam_credential" "linux_backup" {
  type            = "Linux"
  username        = "backupadmin"
  password        = var.linux_password
  description     = "Linux backup server admin"
  elevate_to_root = true
  add_to_sudoers  = true
}

# 2. Register the Linux server in Veeam infrastructure
resource "veeam_managed_server" "linux_backup" {
  name           = "backup01.example.com"
  type           = "LinuxHost"
  credentials_id = veeam_credential.linux_backup.id
  description    = "Primary Linux backup server"
}

# 3. Create a backup repository on that server
resource "veeam_repository" "primary" {
  name           = "Linux-Primary-Repo"
  type           = "LinuxLocal"
  host_id        = veeam_managed_server.linux_backup.id
  path           = "/mnt/backup"
  max_task_count = 4
  description    = "Primary backup repository"
}

# 4. Encryption password for backup data-at-rest encryption
resource "veeam_encryption_password" "main" {
  hint     = "Primary backup encryption key"
  password = var.encryption_password
}

# 5. Daily backup job targeting the repository
resource "veeam_backup_job" "vms_daily" {
  name        = "VMs-Daily-Backup"
  type        = "VSphereBackup"
  description = "Daily backup of production VMs"

  virtual_machines {
    includes {
      platform  = "VSphere"
      type      = "VirtualMachine"
      host_name = "vcenter.example.com"
      name      = "web-prod-01"
      object_id = "vm-101"
    }
    exclude_templates = false
  }

  storage {
    repository_id      = veeam_repository.primary.id
    proxy_auto_select  = true
    retention_type     = "RestorePoints"
    retention_quantity = 14
  }

  schedule {
    run_automatically   = true
    daily_enabled       = true
    daily_local_time    = "22:00"
    daily_kind          = "WeekDays"
    retry_enabled       = true
    retry_count         = 3
    retry_await_minutes = 10
  }
}
```

### Windows repository (WinLocal)

```hcl
resource "veeam_credential" "windows_admin" {
  type        = "Standard"
  username    = "EXAMPLE\\backupadmin"
  password    = var.windows_password
  description = "Windows domain backup account"
}

# Windows managed server must already exist (or be registered separately)
resource "veeam_repository" "win_primary" {
  name               = "Windows-Backup-Repo"
  type               = "WinLocal"
  host_id            = var.windows_host_id   # ID of the registered Windows managed server
  path               = "D:\\VeeamBackup"
  max_task_count     = 4
  task_limit_enabled = true
}
```

### SMB/NFS network share repositories

```hcl
# SMB share — requires credentials for authentication
resource "veeam_repository" "smb_nas" {
  name           = "NAS-SMB-Repo"
  type           = "Smb"
  share_path     = "\\\\nas.example.com\\veeambackup"
  credentials_id = veeam_credential.windows_admin.id
  max_task_count = 2
}

# NFS share — no credentials required
resource "veeam_repository" "nfs_nas" {
  name       = "NAS-NFS-Repo"
  type       = "Nfs"
  share_path = "nas.example.com:/export/veeambackup"
}
```

### Scale-out backup repository (SOBR)

```hcl
# Aggregate two repositories into a SOBR
resource "veeam_scale_out_repository" "sobr" {
  name        = "SOBR-Production"
  description = "Scale-out repository spanning two nodes"

  performance_extent_ids = [
    veeam_repository.linux.id,
    veeam_repository.win_primary.id,
  ]
}
```

### Backup proxy

```hcl
# vSphere proxy
resource "veeam_proxy" "vsphere" {
  type           = "ViProxy"
  description    = "Primary vSphere proxy"
  host_id        = veeam_managed_server.linux_backup.id
  transport_mode = "Auto"
  max_task_count = 4
}

# Hyper-V proxy
resource "veeam_proxy" "hyperv" {
  type          = "HvProxy"
  host_id       = var.hv_host_id
  max_task_count = 2
}
```

### Cloud credentials

```hcl
# AWS S3 integration
resource "veeam_cloud_credential" "aws" {
  name       = "AWS-Backup-Account"
  type       = "Amazon"
  access_key = var.aws_access_key    # sensitive
  secret_key = var.aws_secret_key    # sensitive
}

# Azure Blob Storage
resource "veeam_cloud_credential" "azure_blob" {
  name        = "Azure-Blob-Storage"
  type        = "AzureStorage"
  account     = var.azure_storage_account
  shared_key  = var.azure_shared_key    # sensitive
}

# Azure Compute (for VM backup/restore)
resource "veeam_cloud_credential" "azure_compute" {
  name            = "Azure-Compute"
  type            = "AzureCompute"
  connection_name = "MyAzureSubscription"
  creation_mode   = "ExistingAccount"
  deployment_type = "MicrosoftAzure"
  tenant_id       = var.azure_tenant_id       # sensitive
  application_id  = var.azure_application_id  # sensitive
  application_key = var.azure_application_key # sensitive
}
```

### Query infrastructure state with data sources

```hcl
data "veeam_server_info"  "this" {}
data "veeam_license"      "this" {}
data "veeam_repositories" "all"  {}
data "veeam_backup_jobs"  "all"  {}
data "veeam_job_states"   "all"  {}

output "vbr_version" {
  value = data.veeam_server_info.this.server_version
}

output "running_jobs" {
  value = [
    for j in data.veeam_job_states.all.job_states : j.name
    if j.status == "Running"
  ]
}

output "repository_names" {
  value = [for r in data.veeam_repositories.all.repositories : r.name]
}
```

### Import existing resources

All resources support `terraform import`:

```bash
terraform import veeam_credential.linux_backup <credential-id>
terraform import veeam_managed_server.linux_backup <server-id>
terraform import veeam_repository.primary <repository-id>
terraform import veeam_backup_job.vms_daily <job-id>
```

## Resources

| Resource | Description |
|----------|-------------|
| `veeam_backup_job` | Backup jobs for VMware (`VSphereBackup`), Hyper-V (`HyperVBackup`), and Veeam Agent (`WindowsAgentBackup`, `LinuxAgentBackup`) with storage, schedule, and guest processing |
| `veeam_cloud_credential` | Cloud credentials for AWS, Azure Blob, Azure Compute, Google Cloud |
| `veeam_configuration_backup` | VBR configuration backup settings |
| `veeam_credential` | Standard (Windows/domain) and Linux SSH credentials |
| `veeam_encryption_password` | Encryption passwords for backup data-at-rest encryption |
| `veeam_managed_server` | Managed servers: ViHost, WindowsHost, LinuxHost |
| `veeam_protection_group` | Agent-based protection groups (IndividualComputers, CloudMachines) |
| `veeam_proxy` | Backup proxies: ViProxy (vSphere), HvProxy (Hyper-V), GeneralPurposeProxy |
| `veeam_repository` | Backup repositories: WinLocal, LinuxLocal, Nfs, Smb with task/rate limits |
| `veeam_scale_out_repository` | Scale-out backup repositories (SOBR) with performance extents |

## Data Sources

| Data Source | Description |
|-------------|-------------|
| `veeam_backups` | Backups and optional backup files |
| `veeam_backup_jobs` | Backup jobs (all or filtered by ID/name) |
| `veeam_credentials` | All saved credentials |
| `veeam_job_states` | Aggregated job state overview |
| `veeam_license` | Installed license and consumption |
| `veeam_managed_servers` | Managed servers |
| `veeam_protection_groups` | Protection groups |
| `veeam_proxies` | Backup proxies |
| `veeam_repositories` | Repositories (all or filtered by ID/name) |
| `veeam_repository_states` | Repository capacity and free space |
| `veeam_restore_points` | Restore points |
| `veeam_server_info` | VBR server version and configuration |
| `veeam_sessions` | Session history and status |
| `veeam_wan_accelerators` | WAN accelerators |

## Development

```bash
make build    # build for current platform
make install  # install to ~/.terraform.d/plugins/
make test     # unit tests
make check    # fmt + vet + lint + unit tests
make lint     # golangci-lint
make docs     # regenerate docs/
```

Run a single test:

```bash
go test ./pkg/resources/ -run TestCredential -v
```

Acceptance tests require a live VBR server:

```bash
export VEEAM_HOST="veeam.lab.example.com"
export VEEAM_USERNAME="administrator"
export VEEAM_PASSWORD="secret"
export VEEAM_INSECURE="true"

# Required for repository and proxy acceptance tests:
export TF_VAR_test_host_id="<uuid-of-registered-linux-host>"
export TF_VAR_test_repo_path="/tmp/tf-acc-repo"   # optional, defaults to /tmp/tf-acc-repo

# Required for backup job acceptance tests:
export TF_VAR_test_repo_id="<uuid-of-existing-repository>"
export TF_VAR_test_vcenter_host="vcenter.lab.example.com"
export TF_VAR_test_vm_name="vm-display-name"
export TF_VAR_test_vm_object_id="vm-101"

TF_ACC=1 make testacc
```

Run a single resource's acceptance tests:

```bash
TF_ACC=1 make testacc-repository
TF_ACC=1 make testacc-proxy
TF_ACC=1 make testacc-scale-out-repository
TF_ACC=1 make testacc-backup-job
```

## License

[MPL-2.0](LICENSE)
