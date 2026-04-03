data "veeam_backup_objects" "all" {}

output "backup_object_names" {
  value = [for o in data.veeam_backup_objects.all.objects : o.name]
}
