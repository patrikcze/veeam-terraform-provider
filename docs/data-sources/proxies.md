---
page_title: "veeam_proxies Data Source - terraform-provider-veeam"
subcategory: ""
description: |-
  Lists backup proxies or reads one proxy by ID.
---

# veeam_proxies (Data Source)

Lists backup proxies or reads one proxy by ID.

## Example Usage

```hcl
data "veeam_proxies" "all" {}
```

## Schema

### Optional

- `proxy_id` (String) Reads a single proxy by ID.

### Read-Only

- `id` (String) Data source state identifier.
- `proxies` (List of Object) Proxy records with `id`, `name`, `type`, and `description`.
