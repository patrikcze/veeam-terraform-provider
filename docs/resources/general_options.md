---
page_title: "veeam_general_options Resource - terraform-provider-veeam"
subcategory: ""
description: |-
  Manages Veeam Backup & Replication server-level general options.
---

# veeam_general_options (Resource)

Manages Veeam Backup & Replication server-level general options (`/api/v1/generalOptions`).

This is a **singleton** resource — only one instance may exist per provider configuration. The underlying server object always exists and cannot be deleted; removing this resource from Terraform state does not reset the server configuration.

## Example Usage

```hcl
resource "veeam_general_options" "main" {
  storage_latency_control_enabled = true
  storage_latency_limit_ms        = 20

  email_notifications_enabled = true
  email_smtp_server           = "smtp.example.com"
  email_smtp_port             = 587
  email_from                  = "veeam@example.com"
  email_to                    = "alerts@example.com"
  email_subject               = "[Veeam] %JobResult% - %JobName%"

  snmp_notifications_enabled  = false

  syslog_notifications_enabled = true
  syslog_server                = "syslog.example.com"
  syslog_port                  = 514
}
```

## Schema

### Optional

- `storage_latency_control_enabled` (Boolean) Whether storage latency control is enabled (`storageLatencyControl.isEnabled`).
- `storage_latency_limit_ms` (Number) Latency threshold in milliseconds above which backup activity is throttled (`storageLatencyControl.latencyLimitMs`).
- `email_notifications_enabled` (Boolean) Whether email notifications are enabled (`emailNotifications.isEnabled`).
- `email_smtp_server` (String) SMTP server hostname or IP address (`emailNotifications.smtpServer`).
- `email_smtp_port` (Number) SMTP server port (`emailNotifications.port`).
- `email_from` (String) Sender email address (`emailNotifications.from`).
- `email_to` (String) Recipient email address (`emailNotifications.to`).
- `email_subject` (String) Email subject template (`emailNotifications.subject`).
- `snmp_notifications_enabled` (Boolean) Whether SNMP notifications are enabled. Requires SNMP to be configured in the Veeam console before enabling.
- `syslog_notifications_enabled` (Boolean) Whether syslog event forwarding is enabled (`syslogNotifications.isEnabled`).
- `syslog_server` (String) Syslog server hostname or IP address (`syslogNotifications.dnsName`).
- `syslog_port` (Number) Syslog server UDP/TCP port (`syslogNotifications.port`).

### Read-Only

- `id` (String) Always `"general-options"`. Fixed singleton identifier.

## Import

```bash
terraform import veeam_general_options.main "general-options"
```

## Notes

- All fields are optional and computed. Unset fields retain their current server values.
- The resource uses a GET → merge → PUT pattern: it reads the current server state, applies only the fields defined in the plan, and writes back the full object.
- Delete is a no-op — the server configuration is preserved.
