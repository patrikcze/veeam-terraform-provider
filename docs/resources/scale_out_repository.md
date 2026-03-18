---
page_title: "veeam_scale_out_repository Resource - terraform-provider-veeam"
subcategory: ""
description: |-
  Manages a Veeam scale-out backup repository (SOBR).
---

# veeam_scale_out_repository (Resource)

Manages scale-out backup repositories (SOBR) in Veeam Backup & Replication. A SOBR aggregates one or more standard backup repositories (performance extents) into a single logical target. Jobs write to the SOBR and Veeam automatically distributes workloads across all extents.

## Example Usage

### Basic SOBR with two performance extents

```hcl
resource "veeam_scale_out_repository" "sobr" {
  name        = "SOBR-Production"
  description = "Primary scale-out repository"

  performance_extent_ids = [
    veeam_repository.repo_a.id,
    veeam_repository.repo_b.id,
  ]
}
```

### SOBR with capacity tier enabled

```hcl
resource "veeam_scale_out_repository" "sobr_with_capacity" {
  name        = "SOBR-With-Capacity"
  description = "SOBR backed by object storage capacity tier"

  performance_extent_ids = [
    veeam_repository.primary.id,
  ]

  capacity_tier_enabled = true
}
```

## Schema

### Required

- `name` (String) Unique SOBR name as it appears in the Veeam console.
- `performance_extent_ids` (List of String) Ordered list of repository IDs to include as performance extents. At least one repository is required. Obtain repository IDs from `veeam_repository` resources or data sources.

### Optional

- `description` (String) Human-readable description.
- `capacity_tier_enabled` (Boolean) Enable the capacity tier. Requires object storage configured in the Veeam console before enabling. Defaults to `false`.

### Read-Only

- `id` (String) Scale-out repository identifier assigned by the server (UUID).

---

## Import

Scale-out repositories can be imported using their ID:

```bash
terraform import veeam_scale_out_repository.example <sobr-id>
```

## Notes

- Performance extents (standard repositories) must already exist in Veeam before they can be added to a SOBR. Create them first using `veeam_repository`.
- The order of `performance_extent_ids` is preserved and sent to the API as-is. Reordering IDs in the list triggers an update.
- `capacity_tier_enabled = true` requires that an object storage repository has been configured in the Veeam console and linked to the SOBR separately. The API only sets a flag — it does not configure the storage target.
- Deleting a SOBR removes the logical container but does not delete the underlying performance extent repositories or their data.
- Sealed mode and maintenance mode for individual extents must be managed via the Veeam console or separate API calls — they are not modelled as Terraform attributes.
