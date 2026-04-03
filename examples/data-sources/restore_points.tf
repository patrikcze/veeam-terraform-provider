# List all restore points
data "veeam_restore_points" "all" {}

# Filter restore points for a specific backup object (VM)
data "veeam_restore_points" "vm_points" {
  backup_object_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}

# Look up a specific restore point by ID
data "veeam_restore_points" "specific" {
  restore_point_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}

output "restore_point_count" {
  value = length(data.veeam_restore_points.all.restore_points)
}

output "latest_restore_point" {
  value = data.veeam_restore_points.vm_points.restore_points[0].creation_time
}

output "restore_point_ids" {
  value = [for rp in data.veeam_restore_points.vm_points.restore_points : rp.id]
}
