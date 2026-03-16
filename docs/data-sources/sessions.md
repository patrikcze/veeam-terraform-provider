---
page_title: "veeam_sessions Data Source - terraform-provider-veeam"
subcategory: ""
description: |-
  Lists sessions or reads a single Veeam session by ID.
---

# veeam_sessions (Data Source)

Lists Veeam sessions or reads a single session by ID.

## Example Usage

```hcl
data "veeam_sessions" "all" {}
```

## Schema

### Optional

- `session_id` (String) Reads a single session by ID.

### Read-Only

- `id` (String) Data source state identifier.
- `sessions` (List of Object) Session records with:
	- `id`, `name`, `job_id`, `session_type`, `state`, `result`, `creation_time`, `end_time`.

## Notes

- Without `session_id`, the data source lists sessions from the standard sessions endpoint.
