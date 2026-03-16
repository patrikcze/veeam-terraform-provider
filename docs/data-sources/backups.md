---
page_title: "veeam_backups Data Source - terraform-provider-veeam"
subcategory: ""
description: |-
  Lists backups and optionally backup files.
---

# veeam_backups (Data Source)

Returns backup objects, optionally enriched with backup file details.

## Example Usage

```hcl
data "veeam_backups" "all" {
  include_files = true
}
```

## Schema

### Optional

- `backup_id` (String) Reads a single backup by ID.
- `include_files` (Boolean) Includes nested backup files when supported.

### Read-Only

- `id` (String) Data source state identifier.
- `backups` (List of Object) Backups with fields:
  - `id` (String) Backup ID.
  - `name` (String) Backup name.
  - `type` (String) Backup type.
  - `job_id` (String) Source job ID.
  - `job_name` (String) Source job name.
  - `files` (List of Object) Backup files (`id`, `name`, `type`, `size`).

## Notes

- Without `backup_id`, all available backups are returned.
