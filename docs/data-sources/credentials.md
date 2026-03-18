---
page_title: "veeam_credentials Data Source - terraform-provider-veeam"
subcategory: ""
description: |-
  Lists all credentials configured in Veeam Backup & Replication.
---

# veeam_credentials (Data Source)

Lists all credentials configured in Veeam Backup & Replication.

## Example Usage

```hcl
data "veeam_credentials" "all" {}

output "credentials" {
  value = data.veeam_credentials.all.credentials
}
```

## Schema

### Read-Only

- `credentials` (List of Object) List of credentials. Each object includes:
  - `id` (String) Credential identifier.
  - `username` (String) Username.
  - `description` (String) Description.
  - `type` (String) Credential type (`Standard`, `Linux`).

## Notes

- Passwords are never returned by the API.
- Use this data source to look up the `id` of an existing credential for use in the `credentials_id` argument of resources such as `veeam_managed_server`, `veeam_repository`, and `veeam_protection_group`.
