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
# Standard (Windows/domain style) credential
resource "veeam_credential" "standard" {
  username    = "DOMAIN\\administrator"
  password    = var.windows_admin_password
  description = "Domain administrator credentials"
  type        = "Standard"
}

# Linux credential with password authentication
resource "veeam_credential" "linux_user" {
  username            = "backup"
  password            = var.linux_backup_password
  description         = "Linux backup user credentials"
  type                = "Linux"
  authentication_type = "Password"
  ssh_port            = 22
}
```

## Schema

### Required

- `username` (String) Username used for authentication.
- `password` (String, Sensitive) Secret password value.
- `type` (String) Credential type: `Standard` or `Linux`.

### Optional

- `description` (String) Optional description.
- `ssh_port` (Number) SSH port (Linux only).
- `elevate_to_root` (Boolean) Elevate to root via sudo (Linux only).
- `add_to_sudoers` (Boolean) Automatically add to sudoers (Linux only).
- `use_su` (Boolean) Use `su` instead of `sudo` (Linux only).
- `authentication_type` (String) Linux authentication type: `Password` or `PrivateKey`.
- `private_key` (String, Sensitive) SSH private key (Linux + PrivateKey auth).
- `passphrase` (String, Sensitive) Private key passphrase.
- `root_password` (String, Sensitive) Root password for `su` elevation.

### Read-Only

- `id` (String) Credential identifier assigned by Veeam.

## Import

Credentials can be imported using their ID:

```bash
terraform import veeam_credential.example "credential-id-123"
```

## Notes

- Password values are never returned by the Veeam API.
- For Linux credentials on VBR v13 rev1, set `authentication_type` explicitly (for example `Password`).
- Deleting a credential can impact jobs or infrastructure objects that reference it.
