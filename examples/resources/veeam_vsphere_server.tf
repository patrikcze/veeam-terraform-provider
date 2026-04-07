# Register a vCenter Server in VBR's Virtual Infrastructure (ViHost).
# After apply, VBR can discover VMs and datastores for backup jobs.

resource "veeam_credential" "vcenter" {
  username    = "VSPHERE.LOCAL\\svc-veeam"
  password    = var.vcenter_password
  description = "vCenter service account"
  type        = "Standard"
}

resource "veeam_vsphere_server" "vcenter" {
  name                   = "vcenter.example.com"
  description            = "Production vCenter Server"
  credentials_id         = veeam_credential.vcenter.id
  port                   = 443
  certificate_thumbprint = var.vcenter_thumbprint
}

output "vcenter_id" {
  description = "Managed server ID of the registered vCenter."
  value       = veeam_vsphere_server.vcenter.id
}

output "vcenter_status" {
  description = "VBR connection status for the vCenter."
  value       = veeam_vsphere_server.vcenter.status
}

# Standalone ESXi host (no thumbprint — VBR auto-trusts on first connect).
resource "veeam_vsphere_server" "esxi_standalone" {
  name           = "esxi01.example.com"
  description    = "Standalone ESXi host"
  credentials_id = veeam_credential.vcenter.id
  port           = 443
}
