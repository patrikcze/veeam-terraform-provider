---
page_title: "veeam_replicas Data Source - terraform-provider-veeam"
subcategory: ""
description: |-
  Lists VM replicas or reads one replica by ID.
---

# veeam_replicas (Data Source)

Lists VM replicas managed by Veeam Backup & Replication, or reads a single replica by ID.

## Example Usage

```hcl
data "veeam_replicas" "all" {}

output "replica_names" {
  value = [for r in data.veeam_replicas.all.replicas : r.name]
}
```

## Schema

### Optional

- `replica_id` (String) Reads a single replica by ID.

### Read-Only

- `id` (String) Data source state identifier.
- `replicas` (List of Object) Replica records with `id`, `name`, `type`, `state`, `platform`.
