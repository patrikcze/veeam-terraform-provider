# Singleton resource — only one instance per provider configuration.
resource "veeam_security_analyzer_schedule" "main" {
  run_automatically = true

  # Daily scan at 03:00 local server time
  daily_enabled    = true
  daily_local_time = "03:00"

  # Monthly scan is disabled; daily is sufficient
  monthly_enabled    = false
  monthly_day_of_month = 1
}

output "security_analyzer_schedule_id" {
  value = veeam_security_analyzer_schedule.main.id
}
