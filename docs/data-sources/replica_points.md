---
page_title: "veeam_replica_points Data Source - terraform-provider-veeam"
subcategory: ""
description: |-
  Lists replica restore points or reads one by ID.
---

# veeam_replica_points (Data Source)

Lists replica restore points, or reads a single replica point by ID.

## Example Usage

```hcl
data "veeam_replica_points" "all" {}

output "replica_point_count" {
  value = length(data.veeam_replica_points.all.replica_points)
}
```

## Schema

### Optional

- `replica_point_id` (String) Reads a single replica point by ID.

### Read-Only

- `id` (String) Data source state identifier.
- `replica_points` (List of Object) Replica point records with `id`, `name`, `replica_id`, `creation_time`.
