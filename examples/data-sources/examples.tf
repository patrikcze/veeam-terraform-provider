# Example Terraform configuration showing how to use the Veeam data sources

# Fetch all backup jobs
data "veeam_backup_jobs" "all_jobs" {
  # No parameters needed to fetch all jobs
}

# Fetch a specific backup job by ID
data "veeam_backup_jobs" "specific_job_by_id" {
  job_id = "backup-job-123"
}

# Fetch a specific backup job by name
data "veeam_backup_jobs" "specific_job_by_name" {
  job_name = "Daily Backup"
}

# Fetch all repositories
data "veeam_repositories" "all_repos" {
  # No parameters needed to fetch all repositories
}

# Fetch a specific repository by ID
data "veeam_repositories" "specific_repo_by_id" {
  repository_id = "repo-456"
}

# Fetch a specific repository by name
data "veeam_repositories" "specific_repo_by_name" {
  repository_name = "Primary Storage"
}

# Example outputs showing how to use the data
output "all_backup_jobs" {
  value = data.veeam_backup_jobs.all_jobs.backup_jobs
}

output "specific_job_name" {
  value = data.veeam_backup_jobs.specific_job_by_id.backup_jobs[0].name
}

output "repo_capacity" {
  value = data.veeam_repositories.specific_repo_by_name.repositories[0].capacity
}

output "repo_free_space" {
  value = data.veeam_repositories.specific_repo_by_name.repositories[0].free_space
}
