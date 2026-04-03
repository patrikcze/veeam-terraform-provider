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

# ---------------------------------------------------------------------------
# Priority 6 — Advanced Resources
# ---------------------------------------------------------------------------

# 10. Entra ID Tenant
resource "veeam_entra_id_tenant" "m365" {
  name           = "Contoso M365 Tenant"
  description    = "Microsoft 365 backup tenant"
  tenant_id      = var.entra_tenant_id
  credentials_id = var.entra_credentials_id
}

# 11. Event Forwarding (singleton)
resource "veeam_event_forwarding" "main" {
  snmp_enabled   = true
  snmp_host      = var.snmp_host
  snmp_port      = 162
  snmp_community = "veeam-public"

  syslog_enabled  = true
  syslog_host     = var.syslog_host
  syslog_port     = 514
  syslog_protocol = "UDP"
}

# 12. Global VM Exclusion
resource "veeam_global_vm_exclusion" "ci_vms" {
  name        = "CI-Folder"
  type        = "Folder"
  host_name   = var.vcenter_host
  object_id   = var.ci_folder_moref
  description = "CI/CD VMs — excluded from backup jobs"
}

# 13. Mount Server (lifecycle tied to the managed server — delete is a no-op)
resource "veeam_mount_server" "primary" {
  name              = "backup01-mount"
  description       = "Instant VM recovery mount server"
  managed_server_id = veeam_managed_server.vcenter.id
  type              = "WinServer"
  credentials_id    = veeam_credential.vcenter_cred.id
}

# 14. Recovery Token
resource "veeam_recovery_token" "agent01" {
  name              = "agent01-recovery"
  description       = "Recovery token for agent01 Windows agent"
  managed_server_id = veeam_managed_server.linux_repo_host.id
}

# 15. Security Analyzer Schedule (singleton)
resource "veeam_security_analyzer_schedule" "main" {
  run_automatically    = true
  daily_enabled        = true
  daily_local_time     = "03:00"
  monthly_enabled      = false
  monthly_day_of_month = 1
}

# 16. Storage Latency (singleton)
resource "veeam_storage_latency" "main" {
  enabled               = true
  latency_limit_ms      = 20
  throttling_io_enabled = true
  throttling_io_limit   = 512
  stop_jobs_enabled     = true
  stop_jobs_limit_ms    = 40
}

# 17. Unstructured Data Server
resource "veeam_unstructured_data_server" "cifs" {
  name           = "Corp File Server"
  description    = "Corporate file share for unstructured data backup"
  type           = "CifsShare"
  host_name      = var.fileserver_hostname
  credentials_id = veeam_credential.vcenter_cred.id
}

# ---------------------------------------------------------------------------
# Priority 6 Outputs
# ---------------------------------------------------------------------------

output "entra_tenant_id" {
  value = veeam_entra_id_tenant.m365.id
}

output "recovery_token_value" {
  value     = veeam_recovery_token.agent01.token_value
  sensitive = true
}

output "unstructured_server_id" {
  value = veeam_unstructured_data_server.cifs.id
}
