resource "veeam_managed_server" "proxy_host" {
  name           = "proxy01.example.com"
  type           = "WindowsHost"
  credentials_id = var.windows_credential_id
}

# VMware vSphere proxy (ViProxy)
resource "veeam_proxy" "vsphere" {
  description           = "Primary vSphere backup proxy"
  type                  = "ViProxy"
  host_id               = veeam_managed_server.proxy_host.id
  transport_mode        = "Auto"
  failover_to_network   = true
  host_to_proxy_encryption = false
  max_task_count        = 4
}

# Hyper-V proxy (HvProxy)
resource "veeam_proxy" "hyperv" {
  description    = "Hyper-V backup proxy"
  type           = "HvProxy"
  host_id        = veeam_managed_server.proxy_host.id
  max_task_count = 2
}

output "vsphere_proxy_id" {
  value = veeam_proxy.vsphere.id
}

output "vsphere_proxy_name" {
  value = veeam_proxy.vsphere.name
}
