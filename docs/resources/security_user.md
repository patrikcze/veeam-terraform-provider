---
page_title: "veeam_security_user Resource - terraform-provider-veeam"
subcategory: ""
description: |-
  Manages a Veeam security user account with RBAC role assignment.
---

# veeam_security_user (Resource)

Manages a Veeam Backup & Replication security user account (`/api/v1/security/users`) with RBAC role assignment.

This resource supports **Create, Read, and Delete** only. Changing `login` or `role` forces a destroy and recreate.

## Example Usage

```hcl
resource "veeam_security_user" "operator" {
  login       = "veeam-operator"
  password    = var.operator_password
  description = "Restore operator account"
  role        = "RestoreOperator"
}

resource "veeam_security_user" "admin" {
  login = "veeam-admin"
  password = var.admin_password
  role  = "PortalAdministrator"
}
```

## Schema

### Required

- `login` (String) Login name for the user. Changing this forces a new resource.
- `password` (String, Sensitive) User password. Write-only — never read back from the API.
- `role` (String) RBAC role assigned to the user. Changing this forces a new resource. Common values: `PortalAdministrator`, `PortalUser`, `PortalReadOnlyUser`, `RestoreOperator`.

### Optional

- `description` (String) Optional description.

### Read-Only

- `id` (String) User identifier (assigned by the server).

## Import

```bash
terraform import veeam_security_user.operator <user-id>
```

## Notes

- `password` is write-only. The API never returns it, so changes to the password in config will always trigger an update, but the prior value cannot be detected via plan diff.
- There is no Update API for security users. Changing `login` or `role` replaces the resource entirely.
