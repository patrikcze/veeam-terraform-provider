---
page_title: "veeam_restore_points Data Source - terraform-provider-veeam"
subcategory: ""
description: |-
  Lists restore points globally or per backup object.
---

# veeam_restore_points (Data Source)

Lists restore points globally or per backup object.

## Example Usage

```hcl
data "veeam_restore_points" "all" {}
```

## Schema

### Optional

- `restore_point_id` (String) Reads a single restore point by ID.
- `backup_object_id` (String) Lists restore points for a specific backup object.

### Read-Only

- `id` (String) Data source state identifier.
- `restore_points` (List of Object) Restore point records with `id`, `name`, `backup_id`, `creation_time`, `type`.
