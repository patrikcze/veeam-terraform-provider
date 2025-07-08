# Data Sources Example for Veeam Terraform Provider
#
# This example demonstrates how to use data sources to query
# existing Veeam resources and use them in your configuration.

terraform {
  required_providers {
    veeam = {
      source  = "patrikcze/veeam"
      version = "~> 1.0"
    }
  }
  required_version = ">= 1.0"
}

# Configure the Veeam Provider
provider "veeam" {
  host     = var.veeam_host
  username = var.veeam_username
  password = var.veeam_password
  insecure = var.veeam_insecure
}

# Get all backup jobs
data "veeam_backup_jobs" "all_jobs" {}

# Get a specific backup job by name
data "veeam_backup_jobs" "critical_job" {
  job_name = "Critical-Production-Backup"
}

# Get a specific backup job by ID
data "veeam_backup_jobs" "specific_job" {
  job_id = var.backup_job_id
}

# Get all repositories
data "veeam_repositories" "all_repos" {}

# Get a specific repository by name
data "veeam_repositories" "production_repo" {
  repository_name = "Production-Repository"
}

# Get a specific repository by ID
data "veeam_repositories" "specific_repo" {
  repository_id = var.repository_id
}

# Use data sources to create new resources
resource "veeam_backup_job" "new_job" {
  name    = "New-Backup-Job-${formatdate("YYYY-MM-DD", timestamp())}"
  enabled = true
  
  # Only create if the production repository exists
  count = length(data.veeam_repositories.production_repo.repositories) > 0 ? 1 : 0
}

# Create a repository with capacity based on existing repository
resource "veeam_repository" "scaled_repo" {
  name        = "Scaled-Repository"
  description = "Repository scaled based on existing repository"
  path        = "/backup/scaled"
  type        = "linux"
  
  # Set capacity to 150% of the first repository found
  capacity = length(data.veeam_repositories.all_repos.repositories) > 0 ? (
    data.veeam_repositories.all_repos.repositories[0].capacity * 1.5
  ) : 1099511627776  # Default to 1TB if no repositories found
}

# Use locals for complex data processing
locals {
  # Filter enabled backup jobs
  enabled_jobs = [
    for job in data.veeam_backup_jobs.all_jobs.backup_jobs : job
    if job.enabled
  ]
  
  # Calculate total repository capacity
  total_repository_capacity = sum([
    for repo in data.veeam_repositories.all_repos.repositories : repo.capacity
  ])
  
  # Calculate total free space
  total_free_space = sum([
    for repo in data.veeam_repositories.all_repos.repositories : repo.free_space
  ])
  
  # Calculate total used space
  total_used_space = sum([
    for repo in data.veeam_repositories.all_repos.repositories : repo.used_space
  ])
  
  # Find repositories with low free space (less than 10%)
  low_space_repos = [
    for repo in data.veeam_repositories.all_repos.repositories : repo
    if repo.capacity > 0 && (repo.free_space / repo.capacity) < 0.1
  ]
  
  # Create a map of job names to their enabled status
  job_status_map = {
    for job in data.veeam_backup_jobs.all_jobs.backup_jobs : job.name => job.enabled
  }
}

# Output comprehensive information
output "backup_jobs_summary" {
  description = "Summary of all backup jobs"
  value = {
    total_jobs    = length(data.veeam_backup_jobs.all_jobs.backup_jobs)
    enabled_jobs  = length(local.enabled_jobs)
    disabled_jobs = length(data.veeam_backup_jobs.all_jobs.backup_jobs) - length(local.enabled_jobs)
    job_names     = [for job in data.veeam_backup_jobs.all_jobs.backup_jobs : job.name]
    job_status    = local.job_status_map
  }
}

output "repositories_summary" {
  description = "Summary of all repositories"
  value = {
    total_repositories = length(data.veeam_repositories.all_repos.repositories)
    total_capacity     = local.total_repository_capacity
    total_free_space   = local.total_free_space
    total_used_space   = local.total_used_space
    utilization_percent = local.total_repository_capacity > 0 ? (
      (local.total_used_space / local.total_repository_capacity) * 100
    ) : 0
    low_space_repos = [for repo in local.low_space_repos : repo.name]
  }
}

output "specific_resources" {
  description = "Information about specific resources"
  value = {
    critical_job_exists = length(data.veeam_backup_jobs.critical_job.backup_jobs) > 0
    production_repo_exists = length(data.veeam_repositories.production_repo.repositories) > 0
    
    critical_job_details = length(data.veeam_backup_jobs.critical_job.backup_jobs) > 0 ? {
      name        = data.veeam_backup_jobs.critical_job.backup_jobs[0].name
      enabled     = data.veeam_backup_jobs.critical_job.backup_jobs[0].enabled
      description = data.veeam_backup_jobs.critical_job.backup_jobs[0].description
    } : null
    
    production_repo_details = length(data.veeam_repositories.production_repo.repositories) > 0 ? {
      name        = data.veeam_repositories.production_repo.repositories[0].name
      path        = data.veeam_repositories.production_repo.repositories[0].path
      capacity    = data.veeam_repositories.production_repo.repositories[0].capacity
      free_space  = data.veeam_repositories.production_repo.repositories[0].free_space
      used_space  = data.veeam_repositories.production_repo.repositories[0].used_space
    } : null
  }
}

output "repository_details" {
  description = "Detailed information about all repositories"
  value = {
    for repo in data.veeam_repositories.all_repos.repositories : repo.name => {
      id           = repo.id
      path         = repo.path
      type         = repo.type
      capacity     = repo.capacity
      free_space   = repo.free_space
      used_space   = repo.used_space
      utilization  = repo.capacity > 0 ? (repo.used_space / repo.capacity) * 100 : 0
      status       = repo.status
      created_at   = repo.created_at
      updated_at   = repo.updated_at
    }
  }
}

output "backup_job_details" {
  description = "Detailed information about all backup jobs"
  value = {
    for job in data.veeam_backup_jobs.all_jobs.backup_jobs : job.name => {
      id          = job.id
      enabled     = job.enabled
      description = job.description
      repository  = job.repository
      schedule    = job.schedule
      job_type    = job.job_type
      created_at  = job.created_at
      updated_at  = job.updated_at
    }
  }
}
