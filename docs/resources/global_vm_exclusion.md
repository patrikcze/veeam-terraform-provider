---
page_title: "veeam_global_vm_exclusion Resource - terraform-provider-veeam"
subcategory: ""
description: |-
  Manages a global VM exclusion entry in Veeam Backup & Replication.
---

# veeam_global_vm_exclusion (Resource)

Manages a global VM exclusion entry in Veeam Backup & Replication (`/api/v1/globalExclusions/vm`). Objects added here are excluded from all backup jobs globally, regardless of individual job settings.

All identifying fields (`name`, `type`, `host_name`, `object_id`) carry `RequiresReplace` — any change to these fields results in a destroy and recreate.

## Example Usage

```hcl
# Exclude a specific VM
resource "veeam_global_vm_exclusion" "test_vm" {
  name        = "test-vm-01"
  type        = "VirtualMachine"
  host_name   = "vcenter.example.com"
  object_id   = "vm-1042"
  description = "Non-production test VM"
}

# Exclude an entire folder
resource "veeam_global_vm_exclusion" "dev_folder" {
  name      = "Development"
  type      = "Folder"
  host_name = "vcenter.example.com"
  object_id = "group-d128"
}
```

## Schema

### Required

- `name` (String) Display name of the excluded object. Changing this forces a destroy and recreate.
- `type` (String) Type of the excluded object. Allowed values: `VirtualMachine`, `Folder`, `Datacenter`, `Cluster`, `Host`, `Tag`, `VirtualDisk`. Changing this forces a destroy and recreate.

### Optional

- `host_name` (String) Hostname of the vCenter or ESXi server that owns the object. Changing this forces a destroy and recreate.
- `object_id` (String) vSphere MoRef identifier of the object (e.g. `vm-42`). Changing this forces a destroy and recreate.
- `description` (String) Optional description of the exclusion entry.

### Read-Only

- `id` (String) Exclusion entry identifier (assigned by the server).

## Import

```bash
terraform import veeam_global_vm_exclusion.main <uuid>
```

The UUID can be retrieved via `GET /api/v1/globalExclusions/vm`.
