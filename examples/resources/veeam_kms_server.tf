resource "veeam_kms_server" "corporate" {
  name        = "Corporate KMS"
  hostname    = "kms.example.com"
  port        = 9998
  description = "Primary KMS server for backup encryption keys"
}

output "kms_server_id" {
  value = veeam_kms_server.corporate.id
}
