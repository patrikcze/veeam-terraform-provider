# Read Veeam Backup & Replication server information
data "veeam_server_info" "current" {}

output "server_name" {
  value = data.veeam_server_info.current.server_name
}

output "server_version" {
  value = data.veeam_server_info.current.version
}

output "server_build" {
  value = data.veeam_server_info.current.build_number
}

output "installation_id" {
  value = data.veeam_server_info.current.installation_id
}
