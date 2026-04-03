# Singleton resource — only one instance per provider configuration.
resource "veeam_storage_latency" "main" {
  enabled          = true
  latency_limit_ms = 20

  # Throttle IOPS when latency exceeds the limit
  throttling_io_enabled = true
  throttling_io_limit   = 512

  # Stop jobs when datastore latency is critically high
  stop_jobs_enabled  = true
  stop_jobs_limit_ms = 40
}

output "storage_latency_id" {
  value = veeam_storage_latency.main.id
}
