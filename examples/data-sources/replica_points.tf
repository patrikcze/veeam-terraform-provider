data "veeam_replica_points" "all" {}

output "replica_point_count" {
  value = length(data.veeam_replica_points.all.replica_points)
}
