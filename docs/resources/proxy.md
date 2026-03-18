---
page_title: "veeam_proxy Resource - terraform-provider-veeam"
subcategory: ""
description: |-
  Manages a Veeam backup proxy (ViProxy, HvProxy, or GeneralPurposeProxy).
---

# veeam_proxy (Resource)

Manages a Veeam backup proxy. Supports vSphere (`ViProxy`), Hyper-V (`HvProxy`), and general-purpose (`GeneralPurposeProxy`) proxy types.

## Example Usage

### vSphere Proxy

```hcl
resource "veeam_proxy" "vsphere" {
  type                     = "ViProxy"
  description              = "Primary vSphere backup proxy"
  host_id                  = veeam_managed_server.esxi.id
  transport_mode           = "Auto"
  failover_to_network      = true
  host_to_proxy_encryption = false
  max_task_count           = 4
}
```

### Hyper-V Proxy

```hcl
resource "veeam_proxy" "hyperv" {
  type          = "HvProxy"
  description   = "Primary Hyper-V backup proxy"
  host_id       = veeam_managed_server.hvhost.id
  max_task_count = 2
}
```

### General-Purpose Proxy

```hcl
resource "veeam_proxy" "general" {
  type          = "GeneralPurposeProxy"
  description   = "File-level backup proxy"
  host_id       = veeam_managed_server.windows.id
  max_task_count = 2
}
```

## Schema

### Required

- `type` (String) Proxy type: `ViProxy`, `HvProxy`, or `GeneralPurposeProxy`.
- `host_id` (String) ID of the managed server to assign as the proxy. Obtain from a `veeam_managed_server` resource or data source.

### Optional

- `description` (String) Optional description.
- `transport_mode` (String) Data transport mode (`ViProxy` only). Supported values: `Auto`, `DirectAccess`, `VirtualAppliance`, `Network`. Defaults to `Auto`.
- `failover_to_network` (Boolean) Fall back to network transport if the primary mode fails (`ViProxy` only).
- `host_to_proxy_encryption` (Boolean) Encrypt data in transit between the host and the proxy (`ViProxy` only).
- `max_task_count` (Number) Maximum number of concurrent backup tasks.

### Read-Only

- `id` (String) Proxy identifier assigned by the server (UUID).
- `name` (String) Proxy name (derived from the associated managed server; cannot be set directly).

---

## Import

Proxies can be imported using their ID:

```bash
terraform import veeam_proxy.example <proxy-id>
```

## Notes

- The proxy `name` is derived from the associated managed server hostname and is read-only.
- `transport_mode`, `failover_to_network`, and `host_to_proxy_encryption` apply to `ViProxy` only. For `HvProxy` and `GeneralPurposeProxy`, set only `host_id` and `max_task_count`.
- Transport mode `Auto` lets Veeam choose the most efficient mode for each task.
- Deleting a proxy removes it from the backup infrastructure but does not affect the underlying managed server.
- The legacy `FileProxy` type is an alias for `GeneralPurposeProxy` in API v1.3. Use `GeneralPurposeProxy` in new configurations.
