# Email notification settings (singleton — one per provider configuration)
resource "veeam_email_settings" "main" {
  enabled            = true
  smtp_server        = "smtp.example.com"
  port               = 587
  use_ssl            = true
  use_authentication = true
  login              = "veeam-notifications@example.com"
  password           = var.smtp_password
  from               = "veeam@example.com"
  to                 = "ops-team@example.com"
  subject            = "[Veeam] Job {JobResult} — {JobName}"
  send_on_success    = false
  send_on_warning    = true
  send_on_error      = true
  send_daily_summary = true
  send_test_message  = false
}

output "email_settings_id" {
  value = veeam_email_settings.main.id
}
