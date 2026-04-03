resource "veeam_security_settings" "main" {
  require_ssl       = true
  require_mfa       = false
  block_first_login = false

  # Lock account after 5 consecutive failed login attempts
  login_attempt_limit = 5

  # Auto-logout idle sessions after 30 minutes
  inactivity_timeout_min = 30

  # Enforce password rotation every 90 days
  password_expiration_enabled = true
  password_expiration_days    = 90
}

output "security_settings_id" {
  value = veeam_security_settings.main.id
}
