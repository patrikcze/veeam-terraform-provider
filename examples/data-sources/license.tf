# Read the Veeam license information
data "veeam_license" "current" {}

output "license_type" {
  value = data.veeam_license.current.type
}

output "license_status" {
  value = data.veeam_license.current.status
}

output "license_expiration" {
  value = data.veeam_license.current.expiration_date
}

output "license_licensed_to" {
  value = data.veeam_license.current.licensed_to
}

output "license_socket_usage" {
  value = "${data.veeam_license.current.consumed_sockets} / ${data.veeam_license.current.licensed_sockets}"
}
