data "veeam_server_time" "current" {}

output "server_time" {
  value = data.veeam_server_time.current.server_time
}

output "server_time_zone" {
  value = data.veeam_server_time.current.time_zone
}
