---
page_title: "veeam_repositories Data Source - terraform-provider-veeam"
subcategory: ""
description: |-
  Queries repositories from Veeam Backup & Replication.
---

# veeam_repositories (Data Source)

Queries repositories from Veeam Backup & Replication.

## Example Usage

```hcl
# List all repositories
data "veeam_repositories" "all" {}

output "repository_names" {
  value = [for r in data.veeam_repositories.all.repositories : r.name]
}

# Look up a specific repository by name
data "veeam_repositories" "primary" {
  repository_name = "Linux-Primary-Repo"
}

output "primary_repo_id" {
  value = data.veeam_repositories.primary.repositories[0].id
}
```

## Schema

### Optional

- `repository_id` (String) Filters to a single repository by ID (conflicts with `repository_name`).
- `repository_name` (String) Filters to a single repository by name (conflicts with `repository_id`).

### Read-Only

- `id` (String) Data source state identifier.
- `repositories` (List of Object) Repository objects. Each item includes:
  - `id` (String) Repository identifier.
  - `name` (String) Repository name.
  - `description` (String) Description.
  - `type` (String) Repository type (`WinLocal`, `LinuxLocal`, `Nfs`, `Smb`, etc.).
  - `path` (String) Filesystem or share path.
  - `capacity` (Number) Total capacity in bytes.
  - `free_space` (Number) Available free space in bytes.
  - `used_space` (Number) Used space in bytes.
  - `status` (String) Repository health status.
  - `created_at` (String) Creation timestamp.
  - `updated_at` (String) Last update timestamp.

## Usage Examples

### Filter by available free space

```hcl
data "veeam_repositories" "all" {}

locals {
  repos_with_space = [
    for r in data.veeam_repositories.all.repositories : r
    if r.free_space > 10737418240  # more than 10 GB free
  ]
}

output "repositories_with_space" {
  value = local.repos_with_space
}
```

## Notes

- Without filters, all repositories are returned.
- Results are always returned as a list, even for a single match.
- Use `repository_id` on `veeam_backup_job` to reference a repository returned by this data source.
