terraform {
  required_version = ">= 1.6.0"

  required_providers {
    veeam = {
      source  = "patrikcze/veeam"
      version = "0.1.0"
    }
  }
}

provider "veeam" {
  host     = var.veeam_host
  username = var.veeam_username
  password = var.veeam_password
  insecure = var.veeam_insecure
}

# All data sources health-check
data "veeam_server_info" "all" {}
data "veeam_license" "all" {}
data "veeam_credentials" "all" {}
data "veeam_managed_servers" "all" {}
data "veeam_proxies" "all" {}
data "veeam_repositories" "all" {}
data "veeam_repository_states" "all" {}
data "veeam_backup_jobs" "all" {}
data "veeam_job_states" "all" {}
data "veeam_backups" "all" {}
data "veeam_restore_points" "all" {}
data "veeam_sessions" "all" {}
data "veeam_protection_groups" "all" {}
data "veeam_wan_accelerators" "all" {}

output "datasource_counts" {
  value = {
    credentials       = length(data.veeam_credentials.all.credentials)
    managed_servers   = length(data.veeam_managed_servers.all.servers)
    proxies           = length(data.veeam_proxies.all.proxies)
    repositories      = length(data.veeam_repositories.all.repositories)
    repository_states = length(data.veeam_repository_states.all.states)
    backup_jobs       = length(data.veeam_backup_jobs.all.backup_jobs)
    job_states        = length(data.veeam_job_states.all.states)
    backups           = length(data.veeam_backups.all.backups)
    restore_points    = length(data.veeam_restore_points.all.restore_points)
    sessions          = length(data.veeam_sessions.all.sessions)
    protection_groups = length(data.veeam_protection_groups.all.protection_groups)
    wan_accelerators  = length(data.veeam_wan_accelerators.all.accelerators)
  }
}

output "server_info" {
  value = {
    server_name  = data.veeam_server_info.all.server_name
    version      = data.veeam_server_info.all.version
    build_number = data.veeam_server_info.all.build_number
  }
}

output "license_summary" {
  value = {
    type            = data.veeam_license.all.type
    status          = data.veeam_license.all.status
    licensed_to     = data.veeam_license.all.licensed_to
    expiration_date = data.veeam_license.all.expiration_date
  }
}
