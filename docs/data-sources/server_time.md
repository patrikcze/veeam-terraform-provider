---
page_title: "veeam_server_time Data Source - terraform-provider-veeam"
subcategory: ""
description: |-
  Reads the current server time and time zone.
---

# veeam_server_time (Data Source)

Reads the current date, time, and time zone from the Veeam Backup & Replication server. Useful for validating schedule configurations or for auditing purposes.

## Example Usage

```hcl
data "veeam_server_time" "current" {}

output "server_time" {
  value = data.veeam_server_time.current.server_time
}
```

## Schema

### Read-Only

- `id` (String) Always set to `"server-time"`.
- `server_time` (String) Current server date and time in ISO 8601 format.
- `time_zone` (String) Server time zone name (e.g. `UTC`, `Eastern Standard Time`).
- `utc_offset` (String) UTC offset for the server time zone (e.g. `+00:00`, `-05:00`).
