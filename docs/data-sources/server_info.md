---
page_title: "veeam_server_info Data Source - terraform-provider-veeam"
subcategory: ""
description: |-
  Returns Veeam backup server installation and version information.
---

# veeam_server_info (Data Source)

Returns Veeam backup server installation and version information.

## Example Usage

```hcl
data "veeam_server_info" "this" {}
```

## Schema

### Read-Only

- `id` (String) Data source state identifier.
- `installation_id` (String) Backup server installation ID.
- `server_name` (String) Server host name.
- `build_number` (String) Installed build number.
- `version` (String) Installed Veeam version.
