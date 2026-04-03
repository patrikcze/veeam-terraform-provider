# Exclude a specific VM from all backup jobs
resource "veeam_global_vm_exclusion" "test_vm" {
  name        = "test-vm-01"
  type        = "VirtualMachine"
  host_name   = "vcenter.example.com"
  object_id   = "vm-1042"
  description = "Non-production test VM — excluded from all backup jobs"
}

# Exclude an entire folder from backup
resource "veeam_global_vm_exclusion" "dev_folder" {
  name      = "Development"
  type      = "Folder"
  host_name = "vcenter.example.com"
  object_id = "group-d128"
}

output "test_vm_exclusion_id" {
  value = veeam_global_vm_exclusion.test_vm.id
}
