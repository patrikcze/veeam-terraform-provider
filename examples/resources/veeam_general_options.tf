# General options — email + syslog notifications
resource "veeam_general_options" "main" {
  # Storage latency control
  storage_latency_control_enabled = true
  storage_latency_limit_ms        = 20

  # Email notifications
  email_notifications_enabled = true
  email_smtp_server           = "smtp.example.com"
  email_smtp_port             = 587
  email_from                  = "veeam@example.com"
  email_to                    = "ops-team@example.com"
  email_subject               = "[Veeam] %JobResult% - %JobName%"

  # SNMP trap forwarding (requires SNMP configured in Veeam console)
  snmp_notifications_enabled = false

  # Syslog forwarding
  syslog_notifications_enabled = true
  syslog_server                = "syslog.example.com"
  syslog_port                  = 514
}

output "general_options_id" {
  value = veeam_general_options.main.id
}
