---
page_title: "veeam_backup_job Resource - terraform-provider-veeam"
subcategory: ""
description: |-
  Manages a Veeam Backup & Replication job for VMware vSphere, Hyper-V, or Veeam Agent workloads.
---

# veeam_backup_job (Resource)

Manages backup jobs in Veeam Backup & Replication. Supports VMware vSphere, Microsoft Hyper-V, and Veeam Agent (Windows/Linux) job types.

## Example Usage

### VMware vSphere Backup

```hcl
resource "veeam_backup_job" "vsphere" {
  name        = "Daily-VM-Backup"
  type        = "VSphereBackup"
  description = "Daily backup of production VMs"

  virtual_machines {
    includes {
      platform  = "VSphere"
      type      = "VirtualMachine"
      host_name = "vcenter.example.com"
      name      = "web-prod-01"
      object_id = "vm-101"
    }
    exclude_templates = false
  }

  storage {
    repository_id      = veeam_repository.primary.id
    proxy_auto_select  = true
    retention_type     = "RestorePoints"
    retention_quantity = 14
  }

  schedule {
    run_automatically = true
    daily_enabled     = true
    daily_local_time  = "22:00"
    daily_kind        = "WeekDays"
    retry_enabled     = true
    retry_count       = 3
    retry_await_minutes = 10
  }
}
```

### Hyper-V Backup

```hcl
resource "veeam_backup_job" "hyperv" {
  name        = "HyperV-VM-Backup"
  type        = "HyperVBackup"
  description = "Backup of Hyper-V virtual machines"

  virtual_machines {
    includes {
      platform  = "HyperV"
      type      = "VirtualMachine"
      host_name = "hyperv-host.example.com"
      name      = "app-server-01"
      object_id = "hyperv-vm-guid"
    }
    exclude_templates = false
  }

  storage {
    repository_id      = veeam_repository.primary.id
    proxy_auto_select  = true
    retention_type     = "RestorePoints"
    retention_quantity = 7
  }
}
```

### Windows Agent Backup

```hcl
resource "veeam_backup_job" "windows_agent" {
  name             = "Windows-Agent-Backup"
  type             = "WindowsAgentBackup"
  description      = "Agent-based backup for Windows servers"
  agent_backup_mode = "EntireComputer"

  agent_computers {
    id                  = "computer-uuid"
    name                = "server01.example.com"
    type                = "WindowsComputer"
    protection_group_id = veeam_protection_group.servers.id
  }

  storage {
    repository_id      = veeam_repository.primary.id
    retention_type     = "RestorePoints"
    retention_quantity = 7
  }
}
```

### Job Chaining (After Another Job)

```hcl
resource "veeam_backup_job" "secondary" {
  name        = "Secondary-Backup"
  type        = "VSphereBackup"
  description = "Runs after the primary backup job completes"

  virtual_machines {
    includes {
      platform  = "VSphere"
      type      = "VirtualMachine"
      host_name = "vcenter.example.com"
      name      = "db-prod-01"
      object_id = "vm-202"
    }
    exclude_templates = false
  }

  storage {
    repository_id      = veeam_repository.primary.id
    proxy_auto_select  = true
    retention_type     = "RestorePoints"
    retention_quantity = 7
  }

  schedule {
    run_automatically = true
    after_job_enabled = true
    after_job_name    = "Daily-VM-Backup"
  }
}
```

## Schema

### Required

- `name` (String) Unique job name as it appears in the Veeam console.
- `type` (String) Job type discriminator. Supported values: `VSphereBackup`, `HyperVBackup`, `WindowsAgentBackup`, `LinuxAgentBackup`.
- `description` (String) Human-readable description. Required by the Veeam REST API.

### Optional

- `is_high_priority` (Boolean) If `true`, the resource scheduler prioritises this job over other jobs of the same type. Optional, Computed. Defaults to `false`.
- `virtual_machines` (Block) Defines which VMs or containers are protected. Required for `VSphereBackup` and `HyperVBackup`. See [virtual\_machines](#nested-virtual_machines) below.
- `agent_computers` (List of Blocks) Agent-managed computers or protection groups to include. Required for `WindowsAgentBackup` and `LinuxAgentBackup`. See [agent\_computers](#nested-agent_computers) below.
- `agent_backup_mode` (String) Agent backup scope. Optional, Computed. Required in practice for agent job types. Supported values: `EntireComputer`, `Volumes`, `FileLevel`.
- `include_usb_drives` (Boolean) If `true`, periodically connected USB drives are included in the backup. Optional, Computed. Applies to `WindowsAgentBackup` job type only.
- `agent_type` (String) Protected computer type for Windows agent jobs. Optional, Computed. Supported values: `Workstation`, `Server`, `FailoverCluster`. Applies to `WindowsAgentBackup` job type only.
- `use_snapshotless_file_level_backup` (Boolean) If `true`, creates a crash-consistent file-level backup without a snapshot. Optional, Computed. Applies to `LinuxAgentBackup` job type only, when `agent_backup_mode = "FileLevel"`.
- `storage` (Block) Backup storage configuration. Strongly recommended to set explicitly. See [storage](#nested-storage) below.
- `guest_processing` (Block) Application-aware processing and guest file indexing. Applies to `VSphereBackup` and `HyperVBackup` only. See [guest\_processing](#nested-guest_processing) below.
- `schedule` (Block) Job scheduling configuration. When omitted, the job must be started manually. See [schedule](#nested-schedule) below.

### Read-Only

- `id` (String) Job identifier assigned by the server (UUID).
- `is_disabled` (Boolean) Whether the job is currently disabled. Use the Veeam console or the enable/disable API endpoints to toggle this state.

---

<a id="nested-virtual_machines"></a>
### Nested Block: `virtual_machines`

Defines which inventory objects (VMs, folders, datastores, clusters) are included in the job.

#### Required

- `includes` (List of Blocks) One or more inventory objects to protect. Each block supports:
  - `platform` (String, Required) Hypervisor platform: `VSphere` or `HyperV`.
  - `name` (String, Required) Display name of the inventory object as shown in the hypervisor console.
  - `type` (String, Optional, Computed) Inventory object type. Common values: `VirtualMachine`, `Datastore`, `Folder`, `ResourcePool`, `Cluster`, `Host`, `Tag`.
  - `host_name` (String, Optional, Computed) Hostname or FQDN of the vCenter / Hyper-V host that owns the object.
  - `object_id` (String, Optional, Computed) Managed object reference ID (for example `vm-101` for vSphere, or GUID for Hyper-V).

#### Optional

- `exclude_templates` (Boolean) If `true`, virtual machine templates are automatically excluded from the job. Defaults to `false`.

---

<a id="nested-agent_computers"></a>
### Nested Block: `agent_computers`

Lists the agent-managed computers or protection group members to back up. Use together with `agent_backup_mode`.

Each block supports:

- `id` (String, Required) UUID of the agent-managed object.
- `name` (String, Required) Display name of the computer or protection group.
- `type` (String, Required) Object class. Common values: `ProtectionGroup`, `WindowsComputer`, `LinuxComputer`, `WindowsCluster`, `Domain`, `OrganizationUnit`.
- `protection_group_id` (String, Required) UUID of the protection group that contains this object. Obtain from the `veeam_protection_group` resource or data source.

---

<a id="nested-storage"></a>
### Nested Block: `storage`

Configures where and how long backups are stored.

#### Optional

- `repository_id` (String) UUID of the target backup repository.
- `proxy_auto_select` (Boolean) If `true`, Veeam automatically selects the most suitable backup proxy. Defaults to `true`.
- `retention_type` (String) Retention policy type: `RestorePoints` or `Days`. Defaults to `RestorePoints`.
- `retention_quantity` (Number) Number of restore points or days to retain.

---

<a id="nested-guest_processing"></a>
### Nested Block: `guest_processing`

Controls application-aware processing and guest OS file indexing. Applies to `VSphereBackup` and `HyperVBackup` job types only.

#### Optional

- `app_aware_enabled` (Boolean) Enable application-aware processing. Requires VMware Tools or Hyper-V Integration Services in each guest.
- `fs_indexing_enabled` (Boolean) Enable guest OS file indexing for file-level search inside backup archives.
- `interaction_proxy_auto_select` (Boolean) Automatically select the guest interaction proxy. Defaults to `true`.

---

<a id="nested-schedule"></a>
### Nested Block: `schedule`

Configures when the job runs automatically. When this block is omitted the job must be started manually.

#### Optional

- `run_automatically` (Boolean) Master switch to enable automatic scheduling.
- `daily_enabled` (Boolean) Run on a daily schedule.
- `daily_local_time` (String) Daily start time in `HH:MM` format (server local time).
- `daily_kind` (String) Which days to run. Supported values: `Everyday`, `WeekDays`, `SelectedDays`.
- `monthly_enabled` (Boolean) Run on a monthly schedule.
- `monthly_local_time` (String) Monthly start time in `HH:MM` format.
- `monthly_day_of_month` (Number) Day of the month (1–28) on which the job runs. Note: `monthly_day_of_week`, `monthly_day_number_in_month`, and `monthly_months` are not yet exposed by the provider; configure those via the Veeam console.
- `periodically_enabled` (Boolean) Run at a repeating interval.
- `periodically_kind` (String) Interval unit. Supported values: `Hours`, `Minutes`.
- `periodically_frequency` (Number) Interval value (for example `4` with `Hours` = every 4 hours).
- `after_job_enabled` (Boolean) Start this job automatically after another job completes.
- `after_job_name` (String) Display **name** of the preceding job. The Veeam API v1.3 identifies chained jobs by name, not UUID.
- `retry_enabled` (Boolean) Retry the job automatically on failure.
- `retry_count` (Number) Number of retry attempts.
- `retry_await_minutes` (Number) Minutes to wait between retry attempts.

---

## Import

Backup jobs can be imported using their ID:

```bash
terraform import veeam_backup_job.example <job-id>
```

## Notes

- Job names must be unique within the Veeam environment.
- The `type` attribute uses `RequiresReplace` — changing it destroys and recreates the job.
- `virtual_machines` is required for `VSphereBackup` and `HyperVBackup`; `agent_computers` and `agent_backup_mode` are required for `WindowsAgentBackup` and `LinuxAgentBackup`.
- Object IDs for `virtual_machines.includes.object_id` can be obtained from the vSphere Client (MoRef ID, for example `vm-101`) or via the Veeam REST API inventory endpoints.
- `after_job_name` must be the **display name** of the preceding job, not its UUID — this is an API requirement in Veeam REST API v1.3.
- The `is_disabled` attribute is read-only. Use the Veeam console or the REST API enable/disable endpoints to toggle job state.
- Deleting a backup job does not delete existing backups or restore points stored in the repository.
- For agent jobs, set `storage.repository_id` explicitly when you need a specific repository; otherwise Veeam may use the server default backup repository.
- For agent jobs, `storage.proxy_auto_select` is preserved from Terraform configuration/state because agent storage API responses do not include backup proxy selection fields.
- Veeam can display retry interval defaults in UI even when retry is disabled; in Terraform, `schedule.retry_enabled = false` remains authoritative for behavior.
