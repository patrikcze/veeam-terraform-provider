---
page_title: "veeam_proxy Resource - terraform-provider-veeam"
subcategory: ""
description: |-
  Manages a Veeam backup proxy (ViProxy).
---

# veeam_proxy (Resource)

Manages a Veeam backup proxy. Currently supports vSphere proxies (`ViProxy`).

## Example Usage

```hcl
resource "veeam_proxy" "vsphere" {
  type                     = "ViProxy"
  description              = "Primary vSphere backup proxy"
  host_id                  = veeam_managed_server.esxi.id
  transport_mode           = "auto"
  failover_to_network      = true
  host_to_proxy_encryption = true
  max_task_count           = 4
}
```

## Schema

### Required

- `type` (String) Proxy type: `ViProxy`, `HvProxy`, or `FileProxy`.
- `host_id` (String) ID of the managed server used as the proxy.

### Optional

- `description` (String) Optional description.
- `transport_mode` (String) Data transport mode: `auto`, `directAccess`, `virtualAppliance`, or `network`.
- `failover_to_network` (Boolean) Failover to network transport if primary mode fails.
- `host_to_proxy_encryption` (Boolean) Encrypt data between host and proxy.
- `max_task_count` (Number) Maximum concurrent tasks.

### Read-Only

- `id` (String) Proxy identifier (assigned by the server).
- `name` (String) Proxy name (derived from the host).

## Import

Proxies can be imported using their ID:

```bash
terraform import veeam_proxy.example "proxy-id-123"
```

## Notes

- The proxy name is derived from the associated managed server and is read-only.
- Transport mode `auto` lets Veeam choose the best mode automatically.
- Deleting a proxy removes it from the backup infrastructure but does not affect the underlying managed server.
