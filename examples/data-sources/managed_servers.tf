# List all managed servers
data "veeam_managed_servers" "all" {}

# Look up a specific server by ID
data "veeam_managed_servers" "vcenter" {
  server_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}

output "managed_server_names" {
  value = [for s in data.veeam_managed_servers.all.servers : s.name]
}

# Find all Linux hosts
output "linux_host_ids" {
  value = [
    for s in data.veeam_managed_servers.all.servers : s.id
    if s.type == "LinuxHost"
  ]
}

# Use a discovered server ID as a repository host
output "first_linux_host_id" {
  value = data.veeam_managed_servers.all.servers[0].id
}
