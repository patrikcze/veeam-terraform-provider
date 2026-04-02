---
page_title: "veeam_email_settings Resource - terraform-provider-veeam"
subcategory: ""
description: |-
  Manages Veeam Backup & Replication email notification (SMTP) settings.
---

# veeam_email_settings (Resource)

Manages Veeam Backup & Replication email notification settings (`/api/v1/generalOptions/emailSettings`).

This is a **singleton** resource. Deleting it removes it from Terraform state only; the server SMTP configuration is not reset.

## Example Usage

```hcl
resource "veeam_email_settings" "main" {
  enabled            = true
  smtp_server        = "smtp.example.com"
  port               = 587
  use_ssl            = true
  use_authentication = true
  login              = "veeam-smtp@example.com"
  password           = var.smtp_password
  from               = "veeam@example.com"
  to                 = "alerts@example.com"
  subject            = "[Veeam] %JobResult% - %JobName%"
  send_on_success    = false
  send_on_warning    = true
  send_on_error      = true
  send_daily_summary = true
}
```

## Schema

### Optional

- `enabled` (Boolean) Whether email notifications are enabled.
- `smtp_server` (String) SMTP server hostname or IP address.
- `port` (Number) SMTP server port.
- `use_ssl` (Boolean) Whether to use SSL/TLS for the SMTP connection.
- `use_authentication` (Boolean) Whether SMTP authentication is required.
- `login` (String) SMTP authentication username.
- `password` (String, Sensitive) SMTP authentication password. Write-only — never read back from the API.
- `from` (String) Sender email address.
- `to` (String) Recipient email address.
- `subject` (String) Email subject template. Supports Veeam variables such as `%JobResult%`, `%JobName%`.
- `send_on_success` (Boolean) Send notification on job success.
- `send_on_warning` (Boolean) Send notification on job warning.
- `send_on_error` (Boolean) Send notification on job error.
- `send_daily_summary` (Boolean) Send a daily summary email.
- `send_test_message` (Boolean) When `true`, triggers a test email (`POST .../testMessage`) after each apply. Not stored in state.

### Read-Only

- `id` (String) Always `"email-settings"`. Fixed singleton identifier.

## Import

```bash
terraform import veeam_email_settings.main "email-settings"
```

## Notes

- `password` is write-only; the API never returns it. Changing the password in config will trigger an update but the prior value cannot be detected via plan diff.
- `send_test_message` is an action flag, not a configuration value. It fires on every apply when set to `true`.
