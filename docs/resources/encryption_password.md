---
page_title: "veeam_encryption_password Resource - terraform-provider-veeam"
subcategory: ""
description: |-
  Manages a Veeam encryption password used for backup encryption.
---

# veeam_encryption_password (Resource)

Manages a Veeam encryption password used for backup encryption.

## Example Usage

```hcl
resource "veeam_encryption_password" "backups" {
  password = var.encryption_password
  hint     = "Backup encryption key"
}
```

## Schema

### Required

- `password` (String, Sensitive) The encryption password.
- `hint` (String) Hint to help remember the password.

### Read-Only

- `id` (String) Encryption password identifier (assigned by the server).

## Import

Encryption passwords can be imported using their ID:

```bash
terraform import veeam_encryption_password.example "encryption-password-id-123"
```

## Notes

- The password is stored securely in Veeam and is not returned by the API.
- The hint is visible to Veeam administrators.
- Deleting an encryption password may affect backup jobs that reference it.
