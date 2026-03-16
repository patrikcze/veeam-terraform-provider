---
page_title: "veeam_wan_accelerators Data Source - terraform-provider-veeam"
subcategory: ""
description: |-
  Lists WAN accelerators or reads one accelerator by ID.
---

# veeam_wan_accelerators (Data Source)

Lists WAN accelerators or reads one by ID.

## Example Usage

```hcl
data "veeam_wan_accelerators" "all" {}
```

## Schema

### Optional

- `accelerator_id` (String) Reads a single WAN accelerator by ID.

### Read-Only

- `id` (String) Data source state identifier.
- `accelerators` (List of Object) WAN accelerator records with `id`, `name`, `type`, `description`.
