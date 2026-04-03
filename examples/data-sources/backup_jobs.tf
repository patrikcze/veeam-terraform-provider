# List all backup jobs
data "veeam_backup_jobs" "all" {}

# Look up a specific job by name
data "veeam_backup_jobs" "daily" {
  job_name = "Daily-VM-Backup"
}

# Look up a specific job by ID
data "veeam_backup_jobs" "by_id" {
  job_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}

output "all_job_names" {
  value = [for j in data.veeam_backup_jobs.all.backup_jobs : j.name]
}

output "daily_job_id" {
  value = data.veeam_backup_jobs.daily.backup_jobs[0].id
}

# Reference a job's repository in another resource
output "daily_job_repository" {
  value = data.veeam_backup_jobs.daily.backup_jobs[0].repository
}
