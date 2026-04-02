---
page_title: "veeam_task_sessions Data Source - terraform-provider-veeam"
subcategory: ""
description: |-
  Lists granular task-level session details or reads one task session by ID.
---

# veeam_task_sessions (Data Source)

Lists task-level session details for Veeam jobs, or reads a single task session by ID. Task sessions represent individual object-level processing units within a job session.

## Example Usage

```hcl
data "veeam_task_sessions" "all" {}

output "failed_tasks" {
  value = [for t in data.veeam_task_sessions.all.task_sessions : t.name if t.status == "Failed"]
}
```

## Schema

### Optional

- `task_session_id` (String) Reads a single task session by ID.

### Read-Only

- `id` (String) Data source state identifier.
- `task_sessions` (List of Object) Task session records with `id`, `name`, `session_id`, `status`, `start_time`, `end_time`.
