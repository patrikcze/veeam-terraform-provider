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

## Schema

### Optional

- `repository_id` (String) Filters to a single repository by ID (conflicts with `repository_name`).
- `repository_name` (String) Filters to a single repository by name (conflicts with `repository_id`).

### Read-Only

- `id` (String) Data source state identifier.
- `repositories` (List of Object) Repository objects with capacity and status fields.

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

- Without filters, all repositories are returned.
- Results are always returned as a list, even for a single match.
