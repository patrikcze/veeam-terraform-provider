data "veeam_proxy_states" "all" {}

output "available_proxy_names" {
  value = [for s in data.veeam_proxy_states.all.states : s.name if s.status == "Available"]
}
