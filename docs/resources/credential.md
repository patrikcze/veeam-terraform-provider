---
page_title: "veeam_credential Resource - terraform-provider-veeam"
subcategory: ""
description: |-
  Manages a Veeam credential.
---

# veeam_credential (Resource)

Manages authentication credentials used by Veeam components.

## Example Usage

```hcl
# Windows domain credential
resource "veeam_credential" "windows_domain" {
  name        = "Windows-Domain-Admin"
  description = "Domain administrator credentials"
  username    = "DOMAIN\\administrator"
  password    = var.windows_admin_password
  type        = "windows"
  domain      = "DOMAIN"
}

# Linux credential
resource "veeam_credential" "linux_user" {
  name        = "Linux-Backup-User"
  description = "Linux backup user credentials"
  username    = "backup"
  password    = var.linux_backup_password
  type        = "linux"
}

# Standard credential without domain
resource "veeam_credential" "standard" {
  name     = "Standard-Credential"
  username = "admin"
  password = var.admin_password
  type     = "standard"
}
```

## Schema

### Required

- `name` (String) Unique credential name.
- `username` (String) Username used for authentication.
- `password` (String, Sensitive) Secret password value.
- `type` (String) Credential type: `windows`, `linux`, or `standard`.

### Optional

- `description` (String) Optional description.
- `domain` (String) Domain for Windows-style credentials.

### Read-Only

- `id` (String) Credential identifier assigned by Veeam.

## Import

Credentials can be imported using their ID:

```bash
terraform import veeam_credential.example "credential-id-123"
```

## Notes

- Password values are never returned by the Veeam API.
- Deleting a credential can impact jobs or infrastructure objects that reference it.
