data "veeam_task_sessions" "all" {}

output "failed_task_names" {
  value = [for t in data.veeam_task_sessions.all.task_sessions : t.name if t.status == "Failed"]
}
