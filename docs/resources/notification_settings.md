---
page_title: "veeam_notification_settings Resource - terraform-provider-veeam"
subcategory: ""
description: |-
  Manages Veeam Backup & Replication global notification rules.
---

# veeam_notification_settings (Resource)

Manages Veeam Backup & Replication global notification settings (`/api/v1/generalOptions/notifications`).

This is a **singleton** resource. Deleting it removes it from Terraform state only; the server configuration is not reset.

## Example Usage

```hcl
resource "veeam_notification_settings" "main" {
  notify_on_success                = false
  notify_on_warning                = true
  notify_on_error                  = true
  suppress_repeating_notifications = true
  notify_on_last_retry_only        = true

  send_snmp_on_success = false
  send_snmp_on_warning = true
  send_snmp_on_error   = true

  send_syslog_on_success = false
  send_syslog_on_warning = true
  send_syslog_on_error   = true
}
```

## Schema

### Optional

- `notify_on_success` (Boolean) Send email notification on job success.
- `notify_on_warning` (Boolean) Send email notification on job warning.
- `notify_on_error` (Boolean) Send email notification on job error.
- `suppress_repeating_notifications` (Boolean) Suppress repeated notifications for the same event.
- `notify_on_last_retry_only` (Boolean) Only notify after the last retry attempt.
- `send_snmp_on_success` (Boolean) Send SNMP trap on job success.
- `send_snmp_on_warning` (Boolean) Send SNMP trap on job warning.
- `send_snmp_on_error` (Boolean) Send SNMP trap on job error.
- `send_syslog_on_success` (Boolean) Forward syslog event on job success.
- `send_syslog_on_warning` (Boolean) Forward syslog event on job warning.
- `send_syslog_on_error` (Boolean) Forward syslog event on job error.

### Read-Only

- `id` (String) Always `"notification-settings"`. Fixed singleton identifier.

## Import

```bash
terraform import veeam_notification_settings.main "notification-settings"
```
