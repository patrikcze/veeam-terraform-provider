---
page_title: "veeam_security_analyzer Data Source - terraform-provider-veeam"
subcategory: ""
description: |-
  Reads security best practices check results and last analyzer run metadata.
---

# veeam_security_analyzer (Data Source)

Reads the Security Analyzer results from Veeam Backup & Replication. Returns the list of best practice checks with their pass/fail status and metadata about the most recent analyzer run.

## Example Usage

```hcl
data "veeam_security_analyzer" "results" {}

output "last_run_time" {
  value = data.veeam_security_analyzer.results.last_run_time
}

output "failed_checks" {
  value = [for bp in data.veeam_security_analyzer.results.best_practices : bp.name if bp.status != "Passed"]
}
```

## Schema

### Read-Only

- `id` (String) Always set to `"security-analyzer"`.
- `last_run_time` (String) Timestamp of the most recent security analyzer run.
- `last_run_status` (String) Overall result of the most recent security analyzer run.
- `best_practices` (List of Object) Security best practice check records with `id`, `name`, `status`, `description`.
