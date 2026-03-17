---
page_title: "veeam_proxy Resource - terraform-provider-veeam"
subcategory: ""
description: |-
  Manages a Veeam backup proxy (ViProxy, HvProxy, or FileProxy).
---

# veeam_proxy (Resource)

Manages a Veeam backup proxy. Supports vSphere (`ViProxy`), Hyper-V (`HvProxy`), and file (`FileProxy`) proxy types.

## Example Usage

```hcl
resource "veeam_proxy" "vsphere" {
  type                     = "ViProxy"
  description              = "Primary vSphere backup proxy"
  host_id                  = veeam_managed_server.esxi.id
  transport_mode           = "Auto"
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
- `transport_mode` (String) Data transport mode: `Auto`, `DirectAccess`, `VirtualAppliance`, or `Network`.
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

- The proxy `name` is derived from the associated managed server and cannot be set directly.
- Transport mode `Auto` allows Veeam to select the most efficient mode for each backup task.
- Deleting a proxy removes it from the backup infrastructure but does not affect the underlying managed server.
- `HvProxy` and `FileProxy` types follow the same schema as `ViProxy`; validate settings against your Veeam environment.
