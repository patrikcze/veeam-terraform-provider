# Focused data source lookups

data "veeam_backup_jobs" "specific_job_by_name" {
  job_name = "Daily-VM-Backup"
}

data "veeam_repositories" "specific_repo_by_name" {
  repository_name = "Default Backup Repository"
}

data "veeam_proxies" "specific_proxy" {
  proxy_id = "00000000-0000-0000-0000-000000000000"
}

output "specific_lookup_results" {
  value = {
    matched_jobs         = length(data.veeam_backup_jobs.specific_job_by_name.backup_jobs)
    matched_repositories = length(data.veeam_repositories.specific_repo_by_name.repositories)
    matched_proxies      = length(data.veeam_proxies.specific_proxy.proxies)
  }
}
