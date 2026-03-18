---
page_title: "veeam_managed_servers Data Source - terraform-provider-veeam"
subcategory: ""
description: |-
  Lists managed servers or reads one managed server by ID.
---

# veeam_managed_servers (Data Source)

Lists managed servers or reads one managed server by ID.

## Example Usage

```hcl
data "veeam_managed_servers" "all" {}
```

## Schema

### Optional

- `server_id` (String) Reads a single managed server by ID.

### Read-Only

- `id` (String) Data source state identifier.
- `servers` (List of Object) Managed server records with `id`, `name`, `type`, `description`, `status`.
