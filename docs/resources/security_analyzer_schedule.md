---
page_title: "veeam_security_analyzer_schedule Resource - terraform-provider-veeam"
subcategory: ""
description: |-
  Manages the Veeam security analyzer scan schedule.
---

# veeam_security_analyzer_schedule (Resource)

Manages the Veeam security analyzer scan schedule singleton (`/api/v1/securityAnalyzer/schedule`). Controls when the built-in security analyzer automatically scans backup infrastructure for vulnerabilities and misconfigurations.

This is a singleton resource — only one instance may exist per provider configuration. Deleting the resource removes it from Terraform state only; the server-side schedule configuration is not reset.

## Example Usage

```hcl
resource "veeam_security_analyzer_schedule" "main" {
  run_automatically = true

  daily_enabled    = true
  daily_local_time = "03:00"

  monthly_enabled      = false
  monthly_day_of_month = 1
}
```

## Schema

### Optional

- `run_automatically` (Boolean) Whether the security analyzer runs on schedule automatically.
- `daily_enabled` (Boolean) Whether the daily scan schedule is enabled.
- `daily_local_time` (String) Local time for the daily scan in `HH:MM` format.
- `monthly_enabled` (Boolean) Whether the monthly scan schedule is enabled.
- `monthly_day_of_month` (Number) Day of the month for the monthly scan (1–31).

### Read-Only

- `id` (String) Always `"security-analyzer-schedule"`. Fixed singleton identifier.

## Import

This resource uses a fixed singleton ID:

```bash
terraform import veeam_security_analyzer_schedule.main security-analyzer-schedule
```
