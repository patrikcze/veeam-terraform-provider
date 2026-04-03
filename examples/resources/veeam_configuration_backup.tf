resource "veeam_encryption_password" "config_key" {
  password = var.config_encryption_password
  hint     = "Configuration backup encryption key"
}

resource "veeam_repository" "config_repo" {
  name           = "Config-Backup-Repo"
  type           = "WinLocal"
  host_id        = var.windows_host_id
  path           = "D:\\VeeamConfigBackups"
  max_task_count = 2
}

resource "veeam_configuration_backup" "main" {
  enabled                = true
  repository_id          = veeam_repository.config_repo.id
  restore_points_to_keep = 10
  encryption_enabled     = true
  encryption_password_id = veeam_encryption_password.config_key.id
  # Set to true to trigger an immediate backup on apply
  trigger_on_apply = false
}

output "configuration_backup_id" {
  value = veeam_configuration_backup.main.id
}
