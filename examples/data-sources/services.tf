data "veeam_services" "all" {}

output "running_service_names" {
  value = [for s in data.veeam_services.all.services : s.name if s.status == "Running"]
}
