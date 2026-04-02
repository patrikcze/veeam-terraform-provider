---
page_title: "veeam_security_settings Resource - terraform-provider-veeam"
subcategory: ""
description: |-
  Manages Veeam Backup & Replication server security hardening settings.
---

# veeam_security_settings (Resource)

Manages Veeam Backup & Replication server security settings (`/api/v1/security/settings`).

This is a **singleton** resource. Deleting it removes it from Terraform state only; the server configuration is not reset.

## Example Usage

```hcl
resource "veeam_security_settings" "main" {
  require_ssl                  = true
  require_mfa                  = false
  block_first_login            = true
  login_attempt_limit          = 5
  inactivity_timeout_min       = 30
  password_expiration_enabled  = true
  password_expiration_days     = 90
}
```

## Schema

### Optional

- `require_ssl` (Boolean) Require SSL/TLS for all API connections.
- `require_mfa` (Boolean) Require multi-factor authentication for console logins.
- `block_first_login` (Boolean) Block first-time login until password is changed.
- `login_attempt_limit` (Number) Number of failed login attempts before account lockout.
- `inactivity_timeout_min` (Number) Session inactivity timeout in minutes.
- `password_expiration_enabled` (Boolean) Whether password expiration policy is enforced.
- `password_expiration_days` (Number) Number of days before passwords expire.

### Read-Only

- `id` (String) Always `"security-settings"`. Fixed singleton identifier.

## Import

```bash
terraform import veeam_security_settings.main "security-settings"
```
