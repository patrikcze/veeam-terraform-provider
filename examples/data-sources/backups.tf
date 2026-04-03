# List all backups
data "veeam_backups" "all" {}

# List backups including file-level backups
data "veeam_backups" "with_files" {
  include_files = true
}

# Look up a specific backup by ID
data "veeam_backups" "specific" {
  backup_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}

output "backup_names" {
  value = [for b in data.veeam_backups.all.backups : b.name]
}

output "first_backup_id" {
  value = data.veeam_backups.all.backups[0].id
}

# Use the job ID from a backup to look up the associated job state
output "first_backup_job_id" {
  value = data.veeam_backups.all.backups[0].job_id
}
