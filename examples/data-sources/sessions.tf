# List all recent sessions
data "veeam_sessions" "all" {}

# Look up a specific session by ID
data "veeam_sessions" "specific" {
  session_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}

output "recent_session_names" {
  value = [for s in data.veeam_sessions.all.sessions : s.name]
}

# Find failed sessions
output "failed_sessions" {
  value = [
    for s in data.veeam_sessions.all.sessions : {
      name   = s.name
      job_id = s.job_id
      ended  = s.end_time
    }
    if s.result == "Failed"
  ]
}

output "latest_session_state" {
  value = data.veeam_sessions.all.sessions[0].state
}
