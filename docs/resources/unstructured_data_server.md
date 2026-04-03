---
page_title: "veeam_unstructured_data_server Resource - terraform-provider-veeam"
subcategory: ""
description: |-
  Manages an unstructured data server (NAS/file share backup source) in the Veeam inventory.
---

# veeam_unstructured_data_server (Resource)

Manages an unstructured data server in the Veeam inventory (`/api/v1/inventory/unstructuredDataServers`). Unstructured data servers represent NAS devices and file shares that can be backed up using Veeam's file share backup capabilities.

The `type` attribute is immutable — changing it forces a destroy and recreate.

## Example Usage

```hcl
resource "veeam_credential" "nas_cred" {
  username    = "CORP\\svc-nasbackup"
  password    = var.nas_password
  description = "NAS backup credential"
  type        = "Standard"
}

# CIFS/SMB file server
resource "veeam_unstructured_data_server" "cifs_share" {
  name           = "Corp File Server"
  description    = "Corporate CIFS file server"
  type           = "CifsShare"
  host_name      = "fileserver.example.com"
  credentials_id = veeam_credential.nas_cred.id
}

# NFS share (no credentials required for public NFS)
resource "veeam_unstructured_data_server" "nfs_share" {
  name      = "NFS Data Store"
  type      = "NfsShare"
  host_name = "nas.example.com"
}
```

## Schema

### Required

- `name` (String) Display name of the unstructured data server.
- `type` (String) Server type. Allowed values: `CifsShare`, `NfsShare`, `FileServer`. Changing this forces a destroy and recreate.
- `host_name` (String) FQDN or IP address of the NAS device or file server.

### Optional

- `description` (String) Optional description of the unstructured data server.
- `credentials_id` (String) UUID of the credential used to connect to the server.
- `access_credentials_id` (String) UUID of the credential used for share-level access (CIFS/NFS authentication).

### Read-Only

- `id` (String) Unstructured data server identifier (assigned by the server).

## Import

```bash
terraform import veeam_unstructured_data_server.main <uuid>
```

The UUID can be retrieved via `GET /api/v1/inventory/unstructuredDataServers`.
