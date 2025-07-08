# Advanced Veeam Terraform Provider Example
#
# This example demonstrates advanced usage of the Veeam Terraform provider
# including multiple repositories, complex credential management, and
# resource relationships.

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

# Create multiple repositories for different purposes
resource "veeam_repository" "primary_repo" {
  name        = "Primary-Backup-Repository"
  description = "Primary repository for daily backups"
  path        = "/backup/primary"
  type        = "linux"
  capacity    = 1099511627776  # 1TB
}

resource "veeam_repository" "secondary_repo" {
  name        = "Secondary-Backup-Repository"
  description = "Secondary repository for weekly backups"
  path        = "/backup/secondary"
  type        = "linux"
  capacity    = 2199023255552  # 2TB
}

resource "veeam_repository" "archive_repo" {
  name        = "Archive-Repository"
  description = "Archive repository for long-term retention"
  path        = "/backup/archive"
  type        = "linux"
  capacity    = 5497558138880  # 5TB
}

# Create different types of credentials
resource "veeam_credential" "linux_admin" {
  name        = "Linux-Admin-Credential"
  description = "Linux administrator credential for backup operations"
  username    = var.linux_admin_username
  password    = var.linux_admin_password
  type        = "linux"
}

resource "veeam_credential" "windows_domain" {
  name        = "Windows-Domain-Credential"
  description = "Windows domain credential for VM backups"
  username    = "${var.windows_domain}\\${var.windows_admin_username}"
  password    = var.windows_admin_password
  type        = "windows"
  domain      = var.windows_domain
}

resource "veeam_credential" "vcenter_admin" {
  name        = "vCenter-Admin-Credential"
  description = "vCenter administrator credential"
  username    = var.vcenter_admin_username
  password    = var.vcenter_admin_password
  type        = "standard"
}

# Create backup jobs with different configurations
resource "veeam_backup_job" "daily_vm_backup" {
  name    = "Daily-VM-Backup"
  enabled = true
  
  depends_on = [
    veeam_repository.primary_repo,
    veeam_credential.vcenter_admin
  ]
}

resource "veeam_backup_job" "weekly_full_backup" {
  name    = "Weekly-Full-Backup"
  enabled = true
  
  depends_on = [
    veeam_repository.secondary_repo,
    veeam_credential.vcenter_admin
  ]
}

resource "veeam_backup_job" "monthly_archive" {
  name    = "Monthly-Archive-Backup"
  enabled = true
  
  depends_on = [
    veeam_repository.archive_repo,
    veeam_credential.vcenter_admin
  ]
}

# Use locals for complex calculations
locals {
  total_repository_capacity = (
    veeam_repository.primary_repo.capacity +
    veeam_repository.secondary_repo.capacity +
    veeam_repository.archive_repo.capacity
  )
  
  repository_names = [
    veeam_repository.primary_repo.name,
    veeam_repository.secondary_repo.name,
    veeam_repository.archive_repo.name
  ]
  
  backup_job_names = [
    veeam_backup_job.daily_vm_backup.name,
    veeam_backup_job.weekly_full_backup.name,
    veeam_backup_job.monthly_archive.name
  ]
}

# Output comprehensive information
output "repository_summary" {
  description = "Summary of all created repositories"
  value = {
    repositories = {
      primary = {
        id       = veeam_repository.primary_repo.id
        name     = veeam_repository.primary_repo.name
        path     = veeam_repository.primary_repo.path
        capacity = veeam_repository.primary_repo.capacity
      }
      secondary = {
        id       = veeam_repository.secondary_repo.id
        name     = veeam_repository.secondary_repo.name
        path     = veeam_repository.secondary_repo.path
        capacity = veeam_repository.secondary_repo.capacity
      }
      archive = {
        id       = veeam_repository.archive_repo.id
        name     = veeam_repository.archive_repo.name
        path     = veeam_repository.archive_repo.path
        capacity = veeam_repository.archive_repo.capacity
      }
    }
    total_capacity = local.total_repository_capacity
  }
}

output "credential_summary" {
  description = "Summary of all created credentials"
  value = {
    linux_admin = {
      id   = veeam_credential.linux_admin.id
      name = veeam_credential.linux_admin.name
      type = veeam_credential.linux_admin.type
    }
    windows_domain = {
      id     = veeam_credential.windows_domain.id
      name   = veeam_credential.windows_domain.name
      type   = veeam_credential.windows_domain.type
      domain = veeam_credential.windows_domain.domain
    }
    vcenter_admin = {
      id   = veeam_credential.vcenter_admin.id
      name = veeam_credential.vcenter_admin.name
      type = veeam_credential.vcenter_admin.type
    }
  }
}

output "backup_job_summary" {
  description = "Summary of all created backup jobs"
  value = {
    jobs = {
      daily = {
        name    = veeam_backup_job.daily_vm_backup.name
        enabled = veeam_backup_job.daily_vm_backup.enabled
      }
      weekly = {
        name    = veeam_backup_job.weekly_full_backup.name
        enabled = veeam_backup_job.weekly_full_backup.enabled
      }
      monthly = {
        name    = veeam_backup_job.monthly_archive.name
        enabled = veeam_backup_job.monthly_archive.enabled
      }
    }
    total_jobs = length(local.backup_job_names)
  }
}
