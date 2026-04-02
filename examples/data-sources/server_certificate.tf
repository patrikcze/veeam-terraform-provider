data "veeam_server_certificate" "current" {}

output "cert_thumbprint" {
  value = data.veeam_server_certificate.current.thumbprint
}

output "cert_valid_to" {
  value = data.veeam_server_certificate.current.valid_to
}
