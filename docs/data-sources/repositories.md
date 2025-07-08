# veeam_repositories

Use this data source to query repositories from Veeam Backup & Replication. You can retrieve all repositories or filter by specific criteria.

## Example Usage

```hcl
# Get all repositories
data "veeam_repositories" "all" {}

# Get a specific repository by ID
data "veeam_repositories" "specific_by_id" {
  repository_id = "repo-456"
}

# Get a specific repository by name
data "veeam_repositories" "specific_by_name" {
  repository_name = "Main Repository"
}

# Use the data in outputs
output "all_repositories" {
  value = data.veeam_repositories.all.repositories
}

output "specific_repo_details" {
  value = data.veeam_repositories.specific_by_name.repositories[0]
}

# Use repository data in a resource
resource "veeam_backup_job" "repo_job" {
  name       = "Backup-with-${data.veeam_repositories.specific_by_name.repositories[0].name}-repo"
  repository = data.veeam_repositories.specific_by_name.repositories[0].id
}
```

## Argument Reference

The following arguments are supported:

- `repository_id` - (Optional) The ID of a specific repository to retrieve. Conflicts with `repository_name`.
- `repository_name` - (Optional) The name of a specific repository to retrieve. Conflicts with `repository_id`.

## Attributes Reference

The following attributes are exported:

- `id` - The identifier of the data source.
- `repositories` - A list of repositories. Each repository has the following attributes:
  - `id` - The unique identifier of the repository.
  - `name` - The name of the repository.
  - `description` - The description of the repository.
  - `path` - The file system path of the repository.
  - `type` - The type of the repository (e.g., `linux`, `windows`).
  - `capacity` - The total capacity of the repository in bytes.
  - `free_space` - The free space in the repository in bytes.
  - `used_space` - The used space in the repository in bytes.
  - `status` - The current status of the repository.
  - `created_at` - The timestamp when the repository was created.
  - `updated_at` - The timestamp when the repository was last updated.

## Usage Examples

### Filtering and Processing

```hcl
# Get all repositories and filter based on free space
data "veeam_repositories" "all" {}

locals {
  large_repos = [
    for repo in data.veeam_repositories.all.repositories : repo
    if repo.free_space > 10737418240  # Filter repositories with more than 10GB free space
  ]
}

output "large_repositories" {
  value = local.large_repos
}
```

### Conditional Resource Creation

```hcl
# Create a backup job only if a specific repository exists
data "veeam_repositories" "check_repo" {
  repository_name = "Critical-Repo"
}

resource "veeam_backup_job" "conditional" {
  count = length(data.veeam_repositories.check_repo.repositories) > 0 ? 1 : 0
  
  name = "Critical-Repo-Backup"
  repository = data.veeam_repositories.check_repo.repositories[0].id
}
```

## Notes

- If no arguments are provided, all repositories will be returned
- When using `repository_id` or `repository_name`, only one repository will be returned in the list
- The `repositories` attribute is always a list, even when filtering by ID or name
- If no repositories match the criteria, an empty list will be returned
- The data source provides read-only access to repository information
