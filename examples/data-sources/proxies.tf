# List all backup proxies
data "veeam_proxies" "all" {}

# Look up a specific proxy by ID
data "veeam_proxies" "vsphere" {
  proxy_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}

output "proxy_names" {
  value = [for p in data.veeam_proxies.all.proxies : p.name]
}

# Find all vSphere proxies
output "vsphere_proxy_ids" {
  value = [
    for p in data.veeam_proxies.all.proxies : p.id
    if p.type == "ViProxy"
  ]
}

output "first_proxy_id" {
  value = data.veeam_proxies.all.proxies[0].id
}
