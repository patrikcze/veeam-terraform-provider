# List all WAN accelerators
data "veeam_wan_accelerators" "all" {}

# Look up a specific WAN accelerator by ID
data "veeam_wan_accelerators" "hq" {
  accelerator_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}

output "wan_accelerator_names" {
  value = [for a in data.veeam_wan_accelerators.all.accelerators : a.name]
}

output "first_accelerator_id" {
  value = data.veeam_wan_accelerators.all.accelerators[0].id
}
