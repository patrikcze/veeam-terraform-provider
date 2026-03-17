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

## Known Veeam Behavior

- If an encryption password is currently (or very recently) referenced by Configuration Backup, Veeam may reject deletion with:

```text
Unable to delete selected password because it is in use by: Backup Configuration Job
```

- This is expected Veeam behavior. The provider surfaces this API error and recommends retrying destroy after Configuration Backup fully releases the password (or after switching Configuration Backup to another password in VBR).
- In practice, a second `terraform destroy` shortly after the first one often succeeds.
