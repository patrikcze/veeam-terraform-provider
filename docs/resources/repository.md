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
# Basic Linux repository
resource "veeam_repository" "linux_repo" {
  name        = "Linux-Backup-Repository"
  description = "Primary Linux backup repository"
  path        = "/backup/linux"
  type        = "linux"
}

# Windows repository with capacity limit
resource "veeam_repository" "windows_repo" {
  name        = "Windows-Backup-Repository"
  description = "Windows backup repository with capacity limit"
  path        = "C:\\Backup\\Windows"
  type        = "windows"
  capacity    = 1073741824000  # 1TB in bytes
}

# Repository with minimal configuration
resource "veeam_repository" "minimal" {
  name = "Minimal-Repository"
  path = "/backup/minimal"
  type = "linux"
}
```

## Schema

### Required

- `name` (String) Unique repository name.
- `path` (String) Filesystem path used by the repository.
- `type` (String) Repository type (`linux`, `windows`, `nfs`, `smb`).

### Optional

- `description` (String) Optional repository description.
- `capacity` (Number) Maximum repository capacity in bytes.

### Read-Only

- `id` (String) Repository identifier assigned by Veeam.

## Import

Repositories can be imported using their ID:

```bash
terraform import veeam_repository.example "repository-id-123"
```

## Notes

- Repository paths must be accessible by the Veeam backup server.
- Capacity is defined in bytes.
