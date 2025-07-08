# veeam_repository

Manages a Veeam Backup Repository. This resource allows you to create, update, and delete backup repositories in Veeam Backup & Replication.

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

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name of the repository. Must be unique within the Veeam environment.
- `path` - (Required) The file system path for the repository.
- `type` - (Required) The type of repository. Valid values are `linux`, `windows`, `nfs`, `smb`.
- `description` - (Optional) A description for the repository.
- `capacity` - (Optional) The maximum capacity of the repository in bytes. If not specified, the repository will use available disk space.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The unique identifier of the repository.

## Import

Repositories can be imported using their ID:

```bash
terraform import veeam_repository.example "repository-id-123"
```

## Notes

- Repository names must be unique within the Veeam environment
- The specified path must be accessible by the Veeam server
- For Windows repositories, use backslashes in the path (e.g., `C:\\Backup`)
- For Linux repositories, use forward slashes in the path (e.g., `/backup/linux`)
- The capacity is specified in bytes. Use calculations like `1073741824000` for 1TB
- Deleting a repository will remove it from Veeam configuration but will not delete backup files
