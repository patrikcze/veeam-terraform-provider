data "veeam_security_analyzer" "results" {}

output "security_last_run_time" {
  value = data.veeam_security_analyzer.results.last_run_time
}

output "security_failed_checks" {
  value = [for bp in data.veeam_security_analyzer.results.best_practices : bp.name if bp.status != "Passed"]
}
