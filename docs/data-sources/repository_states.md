---
page_title: "veeam_repository_states Data Source - terraform-provider-veeam"
subcategory: ""
description: |-
  Returns repository capacity and health state information.
---

# veeam_repository_states (Data Source)

Returns repository capacity and health state information.

## Example Usage

```hcl
data "veeam_repository_states" "all" {}
```

## Schema

### Read-Only

- `id` (String) Data source state identifier.
- `states` (List of Object) Repository state records with `id`, `name`, `type`, `status`, `capacity`, `free_space`, `used_space`.
