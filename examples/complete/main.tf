###############################################################################
# Veeam V13 Terraform Provider — Complete Example
#
# This example demonstrates all resource types available in the provider.
# Adjust values to match your environment.
###############################################################################

terraform {
  required_providers {
    veeam = {
      source = "patrikcze/veeam"
    }
  }
}

# Provider configuration — credentials via environment variables or tfvars
provider "veeam" {
  host     = var.veeam_host
  port     = var.veeam_port
  username = var.veeam_username
  password = var.veeam_password
  insecure = var.veeam_insecure
}

# ---------------------------------------------------------------------------
# 1. Credentials
# ---------------------------------------------------------------------------

resource "veeam_credential" "vcenter_cred" {
  username    = var.vcenter_username
  password    = var.vcenter_password
  description = "vCenter admin credential"
  type        = "Standard"
}

resource "veeam_credential" "linux_cred" {
  username           = "backup-user"
  password           = var.linux_password
  description        = "Linux backup agent credential"
  type               = "Linux"
  ssh_port           = 22
  elevate_to_root    = true
  authentication_type = "Password"
}

# ---------------------------------------------------------------------------
# 2. Encryption Password
# ---------------------------------------------------------------------------

resource "veeam_encryption_password" "backup_key" {
  password = var.encryption_password
  hint     = "Backup encryption key for production"
}

# ---------------------------------------------------------------------------
# 3. Managed Servers
# ---------------------------------------------------------------------------

resource "veeam_managed_server" "vcenter" {
  name                  = var.vcenter_host
  description           = "Production vCenter"
  type                  = "ViHost"
  credentials_id        = veeam_credential.vcenter_cred.id
  port                  = 443
  certificate_thumbprint = var.vcenter_thumbprint
}

resource "veeam_managed_server" "linux_repo_host" {
  name           = var.linux_repo_host
  description    = "Linux repository server"
  type           = "LinuxHost"
  credentials_id = veeam_credential.linux_cred.id
  ssh_fingerprint = var.linux_ssh_fingerprint
}

# ---------------------------------------------------------------------------
# 4. Repository
# ---------------------------------------------------------------------------

resource "veeam_repository" "linux_repo" {
  name           = "Linux-Backup-Repo"
  description    = "Primary Linux backup repository"
  type           = "LinuxLocal"
  host_id        = veeam_managed_server.linux_repo_host.id
  path           = "/mnt/veeam-backups"
  max_task_count = 4
}

# ---------------------------------------------------------------------------
# 5. Proxy
# ---------------------------------------------------------------------------

resource "veeam_proxy" "vsphere_proxy" {
  description           = "vSphere backup proxy"
  type                  = "ViProxy"
  host_id               = veeam_managed_server.vcenter.id
  transport_mode        = "Auto"
  failover_to_network   = true
  max_task_count        = 4
}

# ---------------------------------------------------------------------------
# 6. Backup Job
# ---------------------------------------------------------------------------

resource "veeam_backup_job" "daily_backup" {
  name               = "Daily-VM-Backup"
  description        = "Daily backup of production VMs"
  type               = "VSphereBackup"
  is_high_priority   = true
  repository_id      = veeam_repository.linux_repo.id
  proxy_auto_select  = true
  retention_type     = "RestorePoints"
  retention_quantity = 14
  schedule_enabled   = true
  schedule_time      = "22:00"
  schedule_kind      = "WeekDays"
  retry_enabled      = true
  retry_count        = 3
  retry_await_minutes = 10
}

# ---------------------------------------------------------------------------
# 7. Protection Group
# ---------------------------------------------------------------------------

resource "veeam_protection_group" "office_servers" {
  name        = "Office-Servers"
  description = "Office server protection group"
  type        = "IndividualComputers"

  computers = [
    {
      hostname        = "server1.example.com"
      connection_type = "PermanentCredentials"
      credentials_id  = veeam_credential.linux_cred.id
    },
    {
      hostname        = "server2.example.com"
      connection_type = "PermanentCredentials"
      credentials_id  = veeam_credential.linux_cred.id
    }
  ]

  options = [
    {
      install_backup_agent   = true
      distribution_server_id = local.managed_server_id
      update_automatically   = true
      reboot_if_required     = false
    }
  ]
}

# ---------------------------------------------------------------------------
# 8. Configuration Backup
# ---------------------------------------------------------------------------

resource "veeam_configuration_backup" "config" {
  enabled                = true
  repository_id          = veeam_repository.linux_repo.id
  restore_points_to_keep = 14
  encryption_enabled     = true
  encryption_password_id = veeam_encryption_password.backup_key.id
  trigger_on_apply       = false
}

# ---------------------------------------------------------------------------
# 9. Data Sources
# ---------------------------------------------------------------------------

data "veeam_credentials" "all" {}

data "veeam_backup_jobs" "all" {}

data "veeam_repositories" "all" {}

# ---------------------------------------------------------------------------
# Outputs
# ---------------------------------------------------------------------------

output "vcenter_credential_id" {
  value = veeam_credential.vcenter_cred.id
}

output "repository_id" {
  value = veeam_repository.linux_repo.id
}

output "backup_job_id" {
  value = veeam_backup_job.daily_backup.id
}

output "total_credentials" {
  value = length(data.veeam_credentials.all.credentials)
}

output "configuration_backup_id" {
  value = veeam_configuration_backup.config.id
}
