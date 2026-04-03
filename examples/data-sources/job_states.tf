# List states for all jobs
data "veeam_job_states" "all" {}

# Filter states for a specific job
data "veeam_job_states" "daily_backup" {
  job_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}

output "all_job_statuses" {
  value = { for s in data.veeam_job_states.all.states : s.name => s.status }
}

output "daily_backup_last_result" {
  value = data.veeam_job_states.daily_backup.states[0].last_result
}

output "daily_backup_last_run" {
  value = data.veeam_job_states.daily_backup.states[0].last_run
}
