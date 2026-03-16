---
page_title: "veeam_job_states Data Source - terraform-provider-veeam"
subcategory: ""
description: |-
  Returns aggregated Veeam job state information.
---

# veeam_job_states (Data Source)

Returns aggregated job state overview.

## Example Usage

```hcl
data "veeam_job_states" "all" {}
```

## Schema

### Optional

- `job_id` (String) Filters state output to one job ID.

### Read-Only

- `id` (String) Data source state identifier.
- `states` (List of Object) Job state records with `job_id`, `name`, `type`, `status`, `last_result`, `last_run`.
