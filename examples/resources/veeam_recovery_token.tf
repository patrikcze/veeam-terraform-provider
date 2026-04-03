resource "veeam_managed_server" "agent_host" {
  name           = "agent01.example.com"
  type           = "WindowsHost"
  credentials_id = var.windows_credential_id
}

# token_value is set only at creation time.
# Store it in a secret manager immediately — it cannot be retrieved later.
resource "veeam_recovery_token" "agent01" {
  name               = "agent01-recovery"
  description        = "Recovery token for Windows agent on agent01"
  managed_server_id  = veeam_managed_server.agent_host.id
}

# Write the token to an output for use by an external secret store.
# Mark sensitive to prevent it from appearing in plan output.
output "recovery_token_value" {
  value     = veeam_recovery_token.agent01.token_value
  sensitive = true
}

output "recovery_token_id" {
  value = veeam_recovery_token.agent01.id
}
