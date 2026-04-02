---
page_title: "veeam_kms_server Resource - terraform-provider-veeam"
subcategory: ""
description: |-
  Manages a KMS (Key Management Service) server registration in Veeam Backup & Replication.
---

# veeam_kms_server (Resource)

Manages a KMS (Key Management Service) server registration in Veeam Backup & Replication (`/api/v1/kmsServers`).

## Example Usage

```hcl
resource "veeam_kms_server" "main" {
  name        = "Corporate KMS"
  hostname    = "kms.example.com"
  port        = 9998
  description = "Primary KMS for backup encryption"
}
```

## Schema

### Required

- `name` (String) Display name for the KMS server.
- `hostname` (String) KMS server hostname or IP address.

### Optional

- `description` (String) Optional description.
- `port` (Number) KMS server port. Defaults to `9998`.
- `certificate_thumbprint` (String) Expected TLS certificate thumbprint for mutual verification.

### Read-Only

- `id` (String) KMS server identifier (assigned by the server).

## Import

```bash
terraform import veeam_kms_server.main <kms-server-id>
```
