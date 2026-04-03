# Standard Windows credential
resource "veeam_credential" "windows_admin" {
  username    = "CORP\\backup-admin"
  password    = var.windows_password
  description = "Windows backup administrator"
  type        = "Standard"
}

# Linux SSH credential with password authentication
resource "veeam_credential" "linux_backup" {
  username            = "backup-user"
  password            = var.linux_password
  description         = "Linux repository server credential"
  type                = "Linux"
  ssh_port            = 22
  elevate_to_root     = true
  add_to_sudoers      = true
  authentication_type = "Password"
}

# Linux SSH credential with key-based authentication
resource "veeam_credential" "linux_key" {
  username            = "veeam-agent"
  description         = "Linux agent credential (key auth)"
  type                = "Linux"
  ssh_port            = 22
  elevate_to_root     = true
  authentication_type = "PrivateKey"
  private_key         = var.ssh_private_key
  passphrase          = var.ssh_passphrase
}

output "windows_credential_id" {
  value = veeam_credential.windows_admin.id
}

output "linux_credential_id" {
  value = veeam_credential.linux_backup.id
}
