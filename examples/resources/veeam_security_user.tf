# Veeam portal administrator
resource "veeam_security_user" "admin" {
  login       = "CORP\\veeam-admin"
  password    = var.admin_password
  description = "Veeam portal administrator"
  role        = "PortalAdministrator"
}

# Read-only auditor account
resource "veeam_security_user" "auditor" {
  login       = "CORP\\veeam-audit"
  password    = var.auditor_password
  description = "Compliance auditor — read-only access"
  role        = "PortalReadOnlyUser"
}

output "admin_user_id" {
  value = veeam_security_user.admin.id
}
