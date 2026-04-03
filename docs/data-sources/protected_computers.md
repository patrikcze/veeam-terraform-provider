---
page_title: "veeam_protected_computers Data Source - terraform-provider-veeam"
subcategory: ""
description: |-
  Lists agent-protected computers or filters by ID.
---

# veeam_protected_computers (Data Source)

Lists computers protected by Veeam agents, or filters to a single computer by ID.

## Example Usage

```hcl
data "veeam_protected_computers" "all" {}

output "protected_computer_names" {
  value = [for c in data.veeam_protected_computers.all.computers : c.name]
}
```

## Schema

### Optional

- `computer_id` (String) Filters results to a single protected computer by ID.

### Read-Only

- `id` (String) Data source state identifier.
- `computers` (List of Object) Protected computer records with `id`, `name`, `type`, `status`, `platform`.
