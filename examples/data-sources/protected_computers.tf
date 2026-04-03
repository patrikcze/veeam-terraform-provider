data "veeam_protected_computers" "all" {}

output "protected_computer_names" {
  value = [for c in data.veeam_protected_computers.all.computers : c.name]
}
