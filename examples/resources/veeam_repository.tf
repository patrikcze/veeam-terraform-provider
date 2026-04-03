resource "veeam_managed_server" "repo_host" {
  name           = "repo01.example.com"
  type           = "LinuxHost"
  credentials_id = var.linux_credential_id
}

# Linux local repository with XFS fast cloning
resource "veeam_repository" "linux_primary" {
  name           = "Linux-Primary-Repo"
  description    = "Primary Linux backup repository"
  type           = "LinuxLocal"
  host_id        = veeam_managed_server.repo_host.id
  path           = "/mnt/backup"
  max_task_count = 4
  task_limit_enabled             = true
  read_write_limit_enabled       = false
  use_fast_cloning_on_xfs_volumes = true
}

# Windows local repository
resource "veeam_repository" "windows_secondary" {
  name           = "Win-Secondary-Repo"
  description    = "Windows backup repository"
  type           = "WinLocal"
  host_id        = var.windows_host_id
  path           = "D:\\VeeamBackup"
  max_task_count = 2
  task_limit_enabled = true
}

# NFS share repository
resource "veeam_repository" "nfs_share" {
  name       = "NFS-Archive-Repo"
  description = "NFS archive storage"
  type        = "Nfs"
  share_path  = "nas.example.com:/export/veeam"
}

output "linux_repo_id" {
  value = veeam_repository.linux_primary.id
}
