resource "veeam_managed_server" "win_host" {
  name           = "backup01.example.com"
  type           = "WindowsHost"
  credentials_id = var.windows_credential_id
}

# Mount server lifecycle is tied to the managed server — deleting this
# resource only removes it from Terraform state.
resource "veeam_mount_server" "backup01" {
  name               = "backup01-mount"
  description        = "Mount server for instant VM recovery on backup01"
  managed_server_id  = veeam_managed_server.win_host.id
  type               = "WinServer"
  credentials_id     = var.windows_credential_id
}

output "mount_server_id" {
  value = veeam_mount_server.backup01.id
}
