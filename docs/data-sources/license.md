---
page_title: "veeam_license Data Source - terraform-provider-veeam"
subcategory: ""
description: |-
  Returns installed Veeam license details and usage summary.
---

# veeam_license (Data Source)

Returns installed license details and high-level consumption stats.

## Example Usage

```hcl
data "veeam_license" "current" {}
```

## Schema

### Read-Only

- `id` (String) Data source state identifier.
- `type` (String) License type.
- `status` (String) License status.
- `licensed_to` (String) Licensed organization or user.
- `expiration_date` (String) License expiration date.
- `licensed_sockets` (Number) Licensed socket count.
- `consumed_sockets` (Number) Consumed socket count.
- `licensed_instances` (Number) Licensed instance count.
- `consumed_instances` (Number) Consumed instance count.
- `licensed_capacity_tb` (Number) Licensed capacity in TB.
- `consumed_capacity_tb` (Number) Consumed capacity in TB.
