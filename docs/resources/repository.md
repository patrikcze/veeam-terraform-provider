---
page_title: "veeam_repository Resource - terraform-provider-veeam"
subcategory: ""
description: |-
  Manages a Veeam backup repository.
---

# veeam_repository (Resource)

Manages backup repositories in Veeam Backup & Replication. Supports Windows local (`WinLocal`), Linux local (`LinuxLocal`), NFS (`Nfs`), and SMB (`Smb`) repository types. Repository IDs can be used as performance extents in `veeam_scale_out_repository`.

## Example Usage

### Windows local repository

```hcl
resource "veeam_repository" "windows" {
  name               = "Windows-Backup-Repository"
  description        = "Primary Windows backup target"
  type               = "WinLocal"
  host_id            = veeam_managed_server.win.id
  path               = "D:\\Backups"
  max_task_count     = 4
  task_limit_enabled = true
}
```

### Linux local repository

```hcl
resource "veeam_repository" "linux" {
  name               = "Linux-Backup-Repository"
  description        = "Primary Linux backup target"
  type               = "LinuxLocal"
  host_id            = veeam_managed_server.linux.id
  path               = "/mnt/backups"
  max_task_count     = 4
  task_limit_enabled = true
}
```

### NFS share repository

```hcl
resource "veeam_repository" "nfs" {
  name       = "NFS-Repository"
  description = "NFS network share backup target"
  type       = "Nfs"
  share_path = "nfs://nas.example.com/backups"
}
```

### SMB share repository

```hcl
resource "veeam_repository" "smb" {
  name           = "SMB-Repository"
  description    = "SMB network share backup target"
  type           = "Smb"
  share_path     = "\\\\nas.example.com\\backups"
  credentials_id = veeam_credential.smb_cred.id
}
```

## Schema

### Required

- `name` (String) Unique repository name.
- `type` (String) Repository type: `WinLocal`, `LinuxLocal`, `Nfs`, or `Smb`.

### Optional

- `description` (String) Optional repository description.
- `host_id` (String) Managed server ID. Required for `WinLocal` and `LinuxLocal` types.
- `path` (String) Filesystem path on the managed server. Required for `WinLocal` and `LinuxLocal`.
- `max_task_count` (Number) Maximum concurrent backup tasks. Set `task_limit_enabled = true` to activate this limit.
- `task_limit_enabled` (Boolean) Enable the concurrent task limit. Must be `true` when `max_task_count` is set.
- `read_write_rate` (Number) Maximum read/write throughput in MB/s. Set `read_write_limit_enabled = true` to activate this limit.
- `read_write_limit_enabled` (Boolean) Enable the read/write rate limit. Must be `true` when `read_write_rate` is set.
- `share_path` (String) Network share path. Required for `Nfs` and `Smb` types.
- `credentials_id` (String) Credential ID for authenticated SMB share access.

### Read-Only

- `id` (String) Repository identifier assigned by Veeam (UUID).

---

## Import

Repositories can be imported using their ID:

```bash
terraform import veeam_repository.example <repository-id>
```

## Notes

- `host_id` and `path` are required for `WinLocal` and `LinuxLocal` repositories. The provider uses them to populate the mount-server settings that VBR requires for local repository types.
- `share_path` is required for `Nfs` and `Smb` types.
- `credentials_id` is required for SMB shares that need authentication.
- `task_limit_enabled` must be `true` for `max_task_count` to take effect in VBR. Sending only `max_task_count` without the enable flag is ignored by the API.
- `read_write_limit_enabled` must be `true` for `read_write_rate` to take effect.
- Repository IDs can be referenced by `veeam_scale_out_repository.performance_extent_ids` to build a SOBR.
- Deleting a repository removes it from the Veeam infrastructure but does not delete existing backup files stored on disk.
