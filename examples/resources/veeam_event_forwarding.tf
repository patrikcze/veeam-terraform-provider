# Singleton resource — only one instance per provider configuration.
resource "veeam_event_forwarding" "main" {
  # SNMP trap forwarding
  snmp_enabled   = true
  snmp_host      = "snmp-receiver.example.com"
  snmp_port      = 162
  snmp_community = "veeam-public"

  # Syslog forwarding
  syslog_enabled   = true
  syslog_host      = "syslog.example.com"
  syslog_port      = 514
  syslog_protocol  = "UDP"
}

output "event_forwarding_id" {
  value = veeam_event_forwarding.main.id
}
