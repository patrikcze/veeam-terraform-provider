# Basic Veeam Terraform Provider Example
#
# This example demonstrates basic usage of the Veeam Terraform provider
# including provider configuration and simple resource creation.

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

# Create a basic backup repository
resource "veeam_repository" "basic_repo" {
  name        = "Basic-Backup-Repository"
  description = "Basic backup repository for demonstration"
  path        = "/backup/basic"
  type        = "linux"
}

# Create a basic backup job
resource "veeam_backup_job" "basic_job" {
  name    = "Basic-Backup-Job"
  enabled = true
}

# Create a basic credential
resource "veeam_credential" "basic_cred" {
  name        = "Basic-Credential"
  description = "Basic credential for demonstration"
  username    = var.backup_username
  password    = var.backup_password
  type        = "linux"
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

output "credential_id" {
  description = "The ID of the created credential"
  value       = veeam_credential.basic_cred.id
}
