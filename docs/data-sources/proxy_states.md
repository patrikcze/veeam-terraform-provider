---
page_title: "veeam_proxy_states Data Source - terraform-provider-veeam"
subcategory: ""
description: |-
  Lists proxy health and state information.
---

# veeam_proxy_states (Data Source)

Lists the current health and operational state for all backup proxies in the Veeam infrastructure.

## Example Usage

```hcl
data "veeam_proxy_states" "all" {}

output "available_proxies" {
  value = [for s in data.veeam_proxy_states.all.states : s.name if s.status == "Available"]
}
```

## Schema

### Read-Only

- `id` (String) Data source state identifier.
- `states` (List of Object) Proxy state records with `id`, `name`, `status`, `type`.
