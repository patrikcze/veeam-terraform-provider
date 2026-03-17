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
  is_disabled = false

  computers = [
    {
      hostname        = "web01.example.com"
      connection_type = "PermanentCredentials"
      credentials_id  = veeam_credential.linux_user.id
    },
    {
      hostname        = "db01.example.com"
      connection_type = "PermanentCredentials"
      credentials_id  = veeam_credential.linux_user.id
    }
  ]

  options = [
    {
      install_backup_agent   = true
      distribution_server_id = veeam_managed_server.windows.id
      update_automatically   = true
      reboot_if_required     = false
    }
  ]
}
```

## Schema

### Required

- `name` (String) Protection group name.
- `type` (String) Protection group type. Currently this resource supports `IndividualComputers`.
- `computers` (List of Objects) List of computers in the protection group. Each object supports:
  - `hostname` (String, Required) FQDN or IP address of the computer.
  - `connection_type` (String, Required) Connection type: `PermanentCredentials`, `SingleUseCredentials`, `Certificate`.
  - `credentials_id` (String, Optional) Credential ID. Required with `PermanentCredentials`; must be omitted with `Certificate`.

### Optional

- `description` (String) Optional description.
- `is_disabled` (Boolean) Whether the protection group is disabled.
- `options` (List of Objects, max 1) Deployment options for Veeam backup agents. Supports:
  - `distribution_server_id` (String) ID of Windows distribution server for package deployment.
  - `distribution_repository_id` (String) ID of distribution object storage repository.
  - `install_backup_agent` (Boolean) Deploy backup agent from distribution source.
  - `install_cbt_driver` (Boolean) Deploy CBT driver for Windows protected computers.
  - `install_application_plugins` (Boolean) Deploy application plug-ins.
  - `application_plugins` (List of String) Application plugin names (for example `MSSQL`).
  - `update_automatically` (Boolean) Auto-upgrade agents/plugins on discovered computers.
  - `reboot_if_required` (Boolean) Reboot protected computer automatically if required.

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
- When `options.install_backup_agent = true`, set either `options.distribution_server_id` or `options.distribution_repository_id`.
- `SingleUseCredentials` connection type is defined by API but not yet exposed in Terraform schema for this resource.
- Deleting a protection group does not uninstall agents from the target computers.
