resource "veeam_repository" "extent1" {
  name           = "SOBR-Extent-01"
  type           = "LinuxLocal"
  host_id        = var.linux_host_id
  path           = "/mnt/extent01"
  max_task_count = 4
}

resource "veeam_repository" "extent2" {
  name           = "SOBR-Extent-02"
  type           = "LinuxLocal"
  host_id        = var.linux_host_id
  path           = "/mnt/extent02"
  max_task_count = 4
}

resource "veeam_scale_out_repository" "production" {
  name        = "Production-SOBR"
  description = "Scale-out backup repository for production workloads"

  # Extents are assigned in priority order
  performance_extent_ids = [
    veeam_repository.extent1.id,
    veeam_repository.extent2.id,
  ]

  capacity_tier_enabled = false

  placement_policy = {
    # DataLocality keeps restore-point chains on the same extent
    type = "DataLocality"
  }
}

output "sobr_id" {
  value = veeam_scale_out_repository.production.id
}
