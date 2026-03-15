---
page_title: "veeam_protection_group Resource - terraform-provider-veeam"
subcategory: ""
description: |-
  Manages a Veeam agent protection group (IndividualComputers).
---

# veeam_protection_group (Resource)

Manages a Veeam agent protection group. Currently supports `IndividualComputers` type.

## Example Usage

```hcl
resource "veeam_protection_group" "servers" {
  name        = "Production-Servers"
  description = "Agent-based protection for production servers"
  type        = "IndividualComputers"

  computers {
    hostname       = "web01.example.com"
    credentials_id = veeam_credential.linux_user.id
  }

  computers {
    hostname       = "db01.example.com"
    credentials_id = veeam_credential.linux_user.id
  }
}
```

## Schema

### Required

- `name` (String) Protection group name.
- `type` (String) Protection group type: `IndividualComputers`, `CloudMachines`, etc.

### Optional

- `description` (String) Optional description.
- `computers` (Block List) List of computers in the protection group. Each block supports:
  - `hostname` (String, Required) FQDN or IP address of the computer.
  - `credentials_id` (String, Required) Credential ID for the computer.

### Read-Only

- `id` (String) Protection group identifier (assigned by the server).

## Import

Protection groups can be imported using their ID:

```bash
terraform import veeam_protection_group.example "group-id-123"
```

## Notes

- Protection groups are used with Veeam Agent for Linux/Windows.
- The `IndividualComputers` type allows specifying computers by hostname.
- Deleting a protection group does not uninstall agents from the target computers.
