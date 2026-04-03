---
page_title: "veeam_storage_latency Resource - terraform-provider-veeam"
subcategory: ""
description: |-
  Manages the Veeam storage latency control singleton.
---

# veeam_storage_latency (Resource)

Manages the Veeam storage latency control singleton (`/api/v1/generalOptions/storageLatency`). Controls how Veeam responds to high storage latency on datastores during backup operations — it can throttle IOPS or stop jobs entirely when thresholds are exceeded.

This is a singleton resource — only one instance may exist per provider configuration. Deleting the resource removes it from Terraform state only; the server configuration is not reset.

## Example Usage

```hcl
resource "veeam_storage_latency" "main" {
  enabled          = true
  latency_limit_ms = 20

  throttling_io_enabled = true
  throttling_io_limit   = 512

  stop_jobs_enabled  = true
  stop_jobs_limit_ms = 40
}
```

## Schema

### Optional

- `enabled` (Boolean) Whether global storage latency control is enabled.
- `latency_limit_ms` (Number) Global latency threshold in milliseconds. Jobs are throttled when the datastore latency exceeds this value.
- `throttling_io_enabled` (Boolean) Whether IOPS throttling is enabled when the latency limit is exceeded.
- `throttling_io_limit` (Number) Maximum IOPS allowed for backup operations when throttling is active.
- `stop_jobs_enabled` (Boolean) Whether backup jobs should be stopped when the stop-jobs latency threshold is exceeded.
- `stop_jobs_limit_ms` (Number) Latency threshold in milliseconds above which backup jobs are stopped.

### Read-Only

- `id` (String) Always `"storage-latency"`. Fixed singleton identifier.

## Import

This resource uses a fixed singleton ID:

```bash
terraform import veeam_storage_latency.main storage-latency
```
