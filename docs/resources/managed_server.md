---
page_title: "veeam_managed_server Resource - terraform-provider-veeam"
subcategory: ""
description: |-
  Manages a Veeam managed server (ViHost, WindowsHost, LinuxHost).
---

# veeam_managed_server (Resource)

Manages a Veeam managed server. Supports vSphere hosts (`ViHost`), Windows hosts (`WindowsHost`), and Linux hosts (`LinuxHost`).

## Example Usage

### vSphere Host

```hcl
resource "veeam_managed_server" "esxi" {
  name           = "esxi01.example.com"
  description    = "Primary ESXi host"
  type           = "ViHost"
  credentials_id = veeam_credential.vcenter.id
  port           = 443
}
```

### Linux Host

```hcl
resource "veeam_managed_server" "linux" {
  name            = "backup-repo.example.com"
  description     = "Linux backup repository server"
  type            = "LinuxHost"
  credentials_id  = veeam_credential.linux_user.id
  ssh_fingerprint = "ssh-rsa 3072 KCR7SAh1JAqLNVgFaPcL6uEBS72dX+brNJKsjXov22M"
}
```

## Schema

### Required

- `name` (String) FQDN or IP address of the managed server.
- `type` (String) Server type: `ViHost`, `WindowsHost`, or `LinuxHost`.
- `credentials_id` (String) ID of the saved credential used to connect.

### Optional

- `description` (String) Optional description.
- `port` (Number) Connection port (e.g. 443 for ViHost).
- `certificate_thumbprint` (String) TLS certificate thumbprint (ViHost only).
- `ssh_fingerprint` (String) SSH host key fingerprint (LinuxHost only). Use the Veeam/OpenSSH style value (for example `ssh-rsa 3072 ...`), not `SHA256:...`.

### Read-Only

- `id` (String) Server identifier (assigned by the server).
- `status` (String) Server availability status.

## Import

Managed servers can be imported using their ID:

```bash
terraform import veeam_managed_server.example "server-id-123"
```

## Notes

- Server creation may be asynchronous (the API returns 202 Accepted).
- Deleting a managed server removes it from the Veeam infrastructure.
- The `status` field is computed and reflects the current availability of the server.
- For Linux hosts, if `ssh_fingerprint` is omitted, empty, or provided in `SHA256:` format, the provider automatically requests the fingerprint from VBR (`/api/v1/connectionCertificate`) and uses that value for creation.
