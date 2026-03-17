# Basic Veeam Terraform Provider Example
#
# This example demonstrates basic usage of the Veeam Terraform provider
# including provider configuration and simple resource creation.

terraform {
  required_providers {
    veeam = {
      source  = "patrikcze/veeam"
      version = "0.1.0"
    }
  }
  required_version = ">= 1.6.0"
}

# Configure the Veeam Provider
provider "veeam" {
  host     = var.veeam_host
  username = var.veeam_username
  password = var.veeam_password
  insecure = var.veeam_insecure
}

# Create a basic backup repository
resource "veeam_repository" "basic_repo" {
  name        = "Basic-Backup-Repository"
  description = "Basic backup repository for demonstration"
  type        = "WinLocal"
  host_id     = local.repository_host_id
  path        = "D:\\TESTBACKUP"
}

# Create a basic backup job
resource "veeam_backup_job" "basic_job" {
  name          = "Basic-Backup-Job"
  type          = "VSphereBackup"
  repository_id = veeam_repository.basic_repo.id
}

# Read managed servers and pick one by name
data "veeam_managed_servers" "all" {}

locals {
  repository_host_id = one([
    for s in data.veeam_managed_servers.all.servers : s.id
    if lower(s.name) == lower(var.repository_host_name)
  ])
}

# Output the created resources
output "repository_id" {
  description = "The ID of the created repository"
  value       = veeam_repository.basic_repo.id
}

output "backup_job_name" {
  description = "The name of the created backup job"
  value       = veeam_backup_job.basic_job.name
}

