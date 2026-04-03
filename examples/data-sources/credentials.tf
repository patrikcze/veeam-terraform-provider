# List all saved credentials
data "veeam_credentials" "all" {}

output "credential_count" {
  value = length(data.veeam_credentials.all.credentials)
}

output "credential_usernames" {
  value = [for c in data.veeam_credentials.all.credentials : c.username]
}

# Reference the first Linux credential ID in another resource
output "first_linux_credential_id" {
  value = [
    for c in data.veeam_credentials.all.credentials : c.id
    if c.type == "Linux"
  ][0]
}
