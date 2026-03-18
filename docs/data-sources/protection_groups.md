---
page_title: "veeam_protection_groups Data Source - terraform-provider-veeam"
subcategory: ""
description: |-
  Lists protection groups or reads one protection group by ID.
---

# veeam_protection_groups (Data Source)

Lists protection groups or reads one by ID.

## Example Usage

```hcl
data "veeam_protection_groups" "all" {}
```

## Schema

### Optional

- `protection_group_id` (String) Reads a single protection group by ID.

### Read-Only

- `id` (String) Data source state identifier.
- `protection_groups` (List of Object) Protection group records with `id`, `name`, `type`, `description`.
