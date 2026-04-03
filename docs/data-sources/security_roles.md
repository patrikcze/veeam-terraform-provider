---
page_title: "veeam_security_roles Data Source - terraform-provider-veeam"
subcategory: ""
description: |-
  Lists RBAC roles or reads one role by ID.
---

# veeam_security_roles (Data Source)

Lists available RBAC roles defined on the Veeam Backup & Replication server, or reads a single role by ID.

## Example Usage

```hcl
data "veeam_security_roles" "all" {}

output "role_names" {
  value = [for r in data.veeam_security_roles.all.roles : r.name]
}
```

## Schema

### Optional

- `role_id` (String) Reads a single security role by ID.

### Read-Only

- `id` (String) Data source state identifier.
- `roles` (List of Object) RBAC role records with `id`, `name`, `description`.
