data "veeam_security_roles" "all" {}

output "security_role_names" {
  value = [for r in data.veeam_security_roles.all.roles : r.name]
}
