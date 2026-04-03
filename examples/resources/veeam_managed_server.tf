# vCenter / ESXi host (ViHost)
resource "veeam_credential" "vcenter" {
  username    = "CORP\\svc-veeam"
  password    = var.vcenter_password
  description = "vCenter service account"
  type        = "Standard"
}

resource "veeam_managed_server" "vcenter" {
  name                   = "vcenter.example.com"
  description            = "Production vCenter"
  type                   = "ViHost"
  credentials_id         = veeam_credential.vcenter.id
  port                   = 443
  certificate_thumbprint = var.vcenter_thumbprint
}

# Linux host (LinuxHost) — SSH fingerprint auto-resolved by the provider
resource "veeam_credential" "linux" {
  username        = "backup-user"
  password        = var.linux_password
  description     = "Linux repository credential"
  type            = "Linux"
  ssh_port        = 22
  elevate_to_root = true
}

resource "veeam_managed_server" "linux_repo" {
  name           = "repo01.example.com"
  description    = "Linux backup repository host"
  type           = "LinuxHost"
  credentials_id = veeam_credential.linux.id
}

# Windows host (WindowsHost)
resource "veeam_managed_server" "win_proxy" {
  name           = "proxy01.example.com"
  description    = "Windows proxy server"
  type           = "WindowsHost"
  credentials_id = veeam_credential.vcenter.id
}

output "vcenter_id" {
  value = veeam_managed_server.vcenter.id
}

output "linux_repo_id" {
  value = veeam_managed_server.linux_repo.id
}
