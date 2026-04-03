---
page_title: "veeam_services Data Source - terraform-provider-veeam"
subcategory: ""
description: |-
  Lists VBR services and their status.
---

# veeam_services (Data Source)

Lists all Veeam Backup & Replication services and their operational status.

## Example Usage

```hcl
data "veeam_services" "all" {}

output "running_services" {
  value = [for s in data.veeam_services.all.services : s.name if s.status == "Running"]
}
```

## Schema

### Read-Only

- `id` (String) Data source state identifier.
- `services` (List of Object) Service records with `id`, `name`, `status`, `version`.
