resource "veeam_encryption_password" "production" {
  password = var.encryption_password
  hint     = "Production backup encryption key — rotate annually"
}

output "encryption_password_id" {
  value = veeam_encryption_password.production.id
}
