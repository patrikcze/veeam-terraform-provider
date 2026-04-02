data "veeam_security_users" "all" {}

output "security_user_logins" {
  value = [for u in data.veeam_security_users.all.users : u.login]
}
