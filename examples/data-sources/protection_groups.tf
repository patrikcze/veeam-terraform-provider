# List all protection groups
data "veeam_protection_groups" "all" {}

# Look up a specific protection group by ID
data "veeam_protection_groups" "prod" {
  protection_group_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}

output "protection_group_names" {
  value = [for g in data.veeam_protection_groups.all.protection_groups : g.name]
}

output "prod_group_type" {
  value = data.veeam_protection_groups.prod.protection_groups[0].type
}

# Reference a protection group ID in a backup job resource
output "first_protection_group_id" {
  value = data.veeam_protection_groups.all.protection_groups[0].id
}
