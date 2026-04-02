---
page_title: "veeam_security_users Data Source - terraform-provider-veeam"
subcategory: ""
description: |-
  Lists security users or reads one user by ID.
---

# veeam_security_users (Data Source)

Lists configured RBAC users on the Veeam Backup & Replication server, or reads a single user by ID.

## Example Usage

```hcl
data "veeam_security_users" "all" {}

output "user_logins" {
  value = [for u in data.veeam_security_users.all.users : u.login]
}
```

## Schema

### Optional

- `user_id` (String) Reads a single security user by ID.

### Read-Only

- `id` (String) Data source state identifier.
- `users` (List of Object) Security user records with `id`, `login`, `description`, `role_id`.
