---
page_title: "veeam_traffic_rules Resource - terraform-provider-veeam"
subcategory: ""
description: |-
  Manages Veeam Backup & Replication network traffic throttling rules.
---

# veeam_traffic_rules (Resource)

Manages Veeam Backup & Replication network traffic throttling rules (`/api/v1/trafficRules`).

This is a **singleton** resource. Deleting it removes it from Terraform state only; the server configuration is not reset.

## Example Usage

```hcl
resource "veeam_traffic_rules" "main" {
  throttling_enabled = true
  throttling_rules   = jsonencode([
    {
      name             = "Office Hours"
      timeFrom         = "09:00"
      timeTo           = "18:00"
      throttlingValue  = 50
      throttlingUnit   = "Percent"
    }
  ])
}
```

## Schema

### Optional

- `throttling_enabled` (Boolean) Whether network traffic throttling is globally enabled.
- `throttling_rules` (String) JSON-encoded array of throttling rule objects as returned by the Veeam API (`rules` field). Use `jsonencode()` to construct the value. When not set, existing server rules are preserved.

### Read-Only

- `id` (String) Always `"traffic-rules"`. Fixed singleton identifier.

## Import

```bash
terraform import veeam_traffic_rules.main "traffic-rules"
```

## Notes

- `throttling_rules` is stored as a raw JSON string to accommodate the polymorphic structure of the Veeam rules array. Use `jsondecode()` in outputs to inspect individual fields.
- Setting `throttling_rules = "[]"` clears all rules on the server.
