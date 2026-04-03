resource "veeam_notification_settings" "main" {
  notify_on_success = false
  notify_on_warning = true
  notify_on_error   = true

  # Suppress duplicate alerts when the same job fails repeatedly
  suppress_repeating_notifications = true
  notify_on_last_retry_only        = true

  # SNMP traps (requires snmp_notifications_enabled in veeam_general_options)
  send_snmp_on_success = false
  send_snmp_on_warning = true
  send_snmp_on_error   = true

  # Syslog (requires syslog_notifications_enabled in veeam_general_options)
  send_syslog_on_success = false
  send_syslog_on_warning = true
  send_syslog_on_error   = true
}

output "notification_settings_id" {
  value = veeam_notification_settings.main.id
}
