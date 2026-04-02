data "veeam_replicas" "all" {}

output "replica_names" {
  value = [for r in data.veeam_replicas.all.replicas : r.name]
}
