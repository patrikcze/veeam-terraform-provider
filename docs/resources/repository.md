---
page_title: "veeam_repository Resource - terraform-provider-veeam"
subcategory: ""
description: |-
  Manages a Veeam backup repository.
---

# veeam_repository (Resource)

Manages backup repositories used by Veeam jobs.

## Example Usage

```hcl
# Windows local repository on an existing managed server
resource "veeam_repository" "windows_repo" {
  name        = "Windows-Backup-Repository"
  description = "Windows backup repository"
  type        = "WinLocal"
  host_id     = data.veeam_managed_servers.all.servers[0].id
  path        = "D:\\TESTBACKUP"
}

# Linux local repository
resource "veeam_repository" "linux_repo" {
  name        = "Linux-Backup-Repository"
  description = "Primary Linux backup repository"
  type        = "LinuxLocal"
  host_id     = data.veeam_managed_servers.all.servers[0].id
  path        = "/backup/linux"
}
```

## Schema

### Required

- `name` (String) Unique repository name.
- `type` (String) Repository type (`WinLocal`, `LinuxLocal`, `Nfs`, or `Smb`).

### Optional

- `description` (String) Optional repository description.
- `host_id` (String) Managed server ID for local repository types.
- `path` (String) Filesystem path for local repository types.
- `max_task_count` (Number) Maximum concurrent tasks.
- `share_path` (String) Network share path for NFS/SMB repository types.
- `credentials_id` (String) Credential ID for SMB access.

### Read-Only

- `id` (String) Repository identifier assigned by Veeam.

## Import

Repositories can be imported using their ID:

```bash
terraform import veeam_repository.example "repository-id-123"
```

## Notes

- `host_id` and `path` are listed as optional in the schema but are effectively required for `WinLocal` and `LinuxLocal` repositories. The provider uses them to populate the mount-server settings that VBR requires for local repository types.
- `share_path` is required for `Nfs` and `Smb` types.
- `credentials_id` is required for `Smb` when the share requires authentication.
