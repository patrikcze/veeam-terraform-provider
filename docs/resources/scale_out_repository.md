---
page_title: "veeam_scale_out_repository Resource - terraform-provider-veeam"
subcategory: ""
description: |-
  Manages a Veeam scale-out backup repository.
---

# veeam_scale_out_repository (Resource)

Manages scale-out backup repositories (SOBR) in Veeam.

## Example Usage

```hcl
resource "veeam_scale_out_repository" "sobr" {
  name                     = "sobr-main"
  description              = "Primary SOBR"
  capacity_tier_enabled    = true
  maintenance_mode_enabled = false
  sealed_mode_enabled      = false
}
```

## Schema

### Required

- `name` (String) SOBR name.

### Optional

- `description` (String) Optional description.
- `capacity_tier_enabled` (Boolean) Enables capacity tier usage.
- `maintenance_mode_enabled` (Boolean) Enables maintenance mode.
- `sealed_mode_enabled` (Boolean) Enables sealed mode.

### Read-Only

- `id` (String) Scale-out repository identifier.

## Import

Import by repository ID:

```bash
terraform import veeam_scale_out_repository.example "sobr-id-123"
```

## Notes

- A scale-out backup repository (SOBR) aggregates multiple standard repositories (performance extents) into a single logical target. Performance extents must be configured in VBR before enabling advanced tiers.
- `capacity_tier_enabled`, `maintenance_mode_enabled`, and `sealed_mode_enabled` are applied as separate API operations after the repository object is created or updated. A plan that only changes these values will still issue an update call.
