# veeam_credential

Manages a Veeam Credential. This resource allows you to create, update, and delete authentication credentials in Veeam Backup & Replication.

## Example Usage

```hcl
# Windows domain credential
resource "veeam_credential" "windows_domain" {
  name        = "Windows-Domain-Admin"
  description = "Domain administrator credentials"
  username    = "DOMAIN\\administrator"
  password    = var.windows_admin_password
  type        = "windows"
  domain      = "DOMAIN"
}

# Linux credential
resource "veeam_credential" "linux_user" {
  name        = "Linux-Backup-User"
  description = "Linux backup user credentials"
  username    = "backup"
  password    = var.linux_backup_password
  type        = "linux"
}

# Standard credential without domain
resource "veeam_credential" "standard" {
  name     = "Standard-Credential"
  username = "admin"
  password = var.admin_password
  type     = "standard"
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name of the credential. Must be unique within the Veeam environment.
- `username` - (Required) The username for the credential.
- `password` - (Required) The password for the credential. This value is sensitive and will not be displayed in logs.
- `type` - (Required) The type of credential. Valid values are `windows`, `linux`, `standard`.
- `description` - (Optional) A description for the credential.
- `domain` - (Optional) The domain for the credential. Required for Windows domain credentials.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The unique identifier of the credential.

## Import

Credentials can be imported using their ID:

```bash
terraform import veeam_credential.example "credential-id-123"
```

## Notes

- Credential names must be unique within the Veeam environment
- The password is stored securely in Veeam and is not returned by the API
- For Windows domain credentials, include the domain in the username (e.g., `DOMAIN\\username`) or use the `domain` field
- The `type` field determines how the credential is used within Veeam
- Updating the password will change the stored credential in Veeam
- Deleting a credential will remove it from Veeam, but may affect backup jobs that use it
