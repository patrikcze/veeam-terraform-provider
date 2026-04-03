---
page_title: "veeam_backup_objects Data Source - terraform-provider-veeam"
subcategory: ""
description: |-
  Lists backup objects (VMs, machines) or reads one object by ID.
---

# veeam_backup_objects (Data Source)

Lists objects contained within backups (virtual machines, physical computers, etc.), or reads a single object by ID.

## Example Usage

```hcl
data "veeam_backup_objects" "all" {}

output "vm_names" {
  value = [for o in data.veeam_backup_objects.all.objects : o.name if o.type == "VirtualMachine"]
}
```

## Schema

### Optional

- `object_id` (String) Reads a single backup object by ID.

### Read-Only

- `id` (String) Data source state identifier.
- `objects` (List of Object) Backup object records with `id`, `name`, `type`, `backup_id`, `restore_point_count`.
