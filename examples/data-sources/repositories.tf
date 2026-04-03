# List all repositories
data "veeam_repositories" "all" {}

# Filter by name
data "veeam_repositories" "primary" {
  repository_name = "Linux-Primary-Repo"
}

# Look up by ID
data "veeam_repositories" "by_id" {
  repository_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}

output "repository_names" {
  value = [for r in data.veeam_repositories.all.repositories : r.name]
}

output "primary_repo_id" {
  value = data.veeam_repositories.primary.repositories[0].id
}

# Find repositories with available free space (> 100 GB = 107374182400 bytes)
output "repos_with_space" {
  value = [
    for r in data.veeam_repositories.all.repositories : r.name
    if r.free_space > 107374182400
  ]
}

# Use a repository ID as a backup job target
output "first_repo_id" {
  value = data.veeam_repositories.all.repositories[0].id
}
