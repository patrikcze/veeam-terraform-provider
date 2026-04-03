resource "veeam_credential" "vcenter" {
  username    = "administrator@vsphere.local"
  password    = var.vcenter_password
  description = "vCenter service account"
  type        = "Standard"
}

resource "veeam_managed_server" "vcenter" {
  name           = "vcenter.example.com"
  description    = "Production vCenter"
  type           = "ViHost"
  credentials_id = veeam_credential.vcenter.id
  port           = 443
}

resource "veeam_repository" "primary" {
  name           = "Primary-Repo"
  description    = "Primary backup repository"
  type           = "WinLocal"
  host_id        = veeam_managed_server.vcenter.id
  path           = "D:\\VeeamBackups"
  max_task_count = 4
}

resource "veeam_backup_job" "daily" {
  name               = "Daily-Production-Backup"
  description        = "Daily backup of all production VMs"
  type               = "VSphereBackup"
  is_high_priority   = false
  repository_id      = veeam_repository.primary.id
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

output "backup_job_id" {
  value = veeam_backup_job.daily.id
}
