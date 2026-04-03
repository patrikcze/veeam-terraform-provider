---
page_title: "veeam_mount_server Resource - terraform-provider-veeam"
subcategory: ""
description: |-
  Manages a mount server in the Veeam backup infrastructure.
---

# veeam_mount_server (Resource)

Manages a mount server in the Veeam backup infrastructure (`/api/v1/backupInfrastructure/mountServers`). Mount servers are used for instant VM recovery, file-level restore, and application item recovery.

Mount servers do not have a dedicated delete endpoint in the Veeam REST API — their lifecycle is bound to the managed server they belong to. Deleting this resource removes it from Terraform state only; the actual mount server configuration is not removed from the Veeam server.

## Example Usage

```hcl
resource "veeam_managed_server" "win_host" {
  name           = "backup01.example.com"
  type           = "WindowsHost"
  credentials_id = var.windows_credential_id
}

resource "veeam_mount_server" "backup01" {
  name              = "backup01-mount"
  description       = "Mount server for instant VM recovery"
  managed_server_id = veeam_managed_server.win_host.id
  type              = "WinServer"
  credentials_id    = var.windows_credential_id
}
```

## Schema

### Required

- `name` (String) Display name of the mount server.
- `managed_server_id` (String) UUID of the managed server (host) that owns this mount server. Changing this forces a destroy and recreate.
- `type` (String) Mount server type (e.g. `WinServer`, `LinuxServer`). Changing this forces a destroy and recreate.

### Optional

- `description` (String) Optional description of the mount server.
- `credentials_id` (String) UUID of the credential used to connect to the mount server.

### Read-Only

- `id` (String) Mount server identifier (assigned by the server).

## Import

```bash
terraform import veeam_mount_server.main <uuid>
```

The UUID can be retrieved via `GET /api/v1/backupInfrastructure/mountServers`.

Note: After import, the `managed_server_id` and `type` fields must be supplied in the configuration to prevent destroy and recreate on the next apply.
