---
page_title: "Resources - terraform-provider-veeam"
subcategory: ""
description: |-
  Index of Terraform resources supported by the Veeam provider.
---

# Resources

Reference for all Veeam Terraform Provider resources.

## Available Resources

### [veeam_ad_domain](ad_domain.md)
Manages Active Directory domain registration in the Veeam inventory.

### [veeam_backup_job](backup_job.md)
Manages backup jobs for VMware (VSphereBackup), Hyper-V (HyperVBackup), and Veeam Agent workloads.

### [veeam_cloud_credential](cloud_credential.md)
Manages cloud credentials for AWS, Azure Blob, Azure Compute, and Google Cloud integrations.

### [veeam_configuration_backup](configuration_backup.md)
Manages VBR configuration backup settings and can trigger configuration backup runs.

### [veeam_credential](credential.md)
Manages standard (Windows/domain) and Linux SSH credentials.

### [veeam_email_settings](email_settings.md)
Manages SMTP email notification settings (singleton).

### [veeam_encryption_password](encryption_password.md)
Manages encryption passwords used by backup jobs and configuration backup.

### [veeam_entra_id_tenant](entra_id_tenant.md)
Manages a Microsoft Entra ID (Azure AD) tenant in the Veeam inventory.

### [veeam_event_forwarding](event_forwarding.md)
Manages SNMP trap and syslog event forwarding configuration (singleton).

### [veeam_general_options](general_options.md)
Manages server-level general options: storage latency, email, SNMP, syslog (singleton).

### [veeam_global_vm_exclusion](global_vm_exclusion.md)
Manages global VM exclusion entries (VirtualMachine, Folder, Tag, and more).

### [veeam_kms_server](kms_server.md)
Manages KMS (Key Management Service) server registration for encryption.

### [veeam_managed_server](managed_server.md)
Manages servers in the Veeam infrastructure: ViHost, WindowsHost, LinuxHost.

### [veeam_mount_server](mount_server.md)
Manages mount server registration in the backup infrastructure.

### [veeam_notification_settings](notification_settings.md)
Manages global job notification rules for email, SNMP, and syslog (singleton).

### [veeam_protection_group](protection_group.md)
Manages agent-based protection groups (IndividualComputers, CloudMachines).

### [veeam_proxy](proxy.md)
Manages backup proxies: ViProxy (vSphere), HvProxy (Hyper-V), GeneralPurposeProxy.

### [veeam_recovery_token](recovery_token.md)
Manages agent recovery tokens issued for managed servers.

### [veeam_repository](repository.md)
Manages backup repositories: WinLocal, LinuxLocal, Nfs, Smb with task/rate limits.

### [veeam_scale_out_repository](scale_out_repository.md)
Manages scale-out backup repositories (SOBR) with performance extents.

### [veeam_security_analyzer_schedule](security_analyzer_schedule.md)
Manages the security analyzer scan schedule (singleton).

### [veeam_security_settings](security_settings.md)
Manages server security hardening: SSL, MFA, lockout, password expiration (singleton).

### [veeam_security_user](security_user.md)
Manages security user accounts with RBAC role assignment.

### [veeam_storage_latency](storage_latency.md)
Manages storage latency control thresholds (singleton).

### [veeam_traffic_rules](traffic_rules.md)
Manages network traffic throttling rules (singleton).

### [veeam_unstructured_data_server](unstructured_data_server.md)
Manages unstructured data server registration for NAS backup.

## Common Patterns

### Resource Dependencies
Many resources work together. A typical pattern is repository + credential + backup job:

```hcl
resource "veeam_repository" "main" {
  name = "Main-Repository"
  path = "/backup/main"
  type = "linux"
}

resource "veeam_credential" "backup_user" {
  name     = "Backup-User"
  username = "backup"
  password = var.backup_password
  type     = "linux"
}

resource "veeam_backup_job" "daily" {
  name    = "Daily-Backup"
  enabled = true
  
  depends_on = [
    veeam_repository.main,
    veeam_credential.backup_user
  ]
}
```

### Import Existing Resources
All resources support importing existing Veeam objects:

```bash
# Import by name for backup jobs
terraform import veeam_backup_job.existing "Existing-Job-Name"

# Import by ID for repositories and credentials
terraform import veeam_repository.existing "repo-id-123"
terraform import veeam_credential.existing "cred-id-456"
```

### Variable Usage
Use variables for sensitive values and reusable inputs:

```hcl
variable "repository_path" {
  description = "Base path for repositories"
  type        = string
  default     = "/backup"
}

variable "admin_password" {
  description = "Administrator password"
  type        = string
  sensitive   = true
}

resource "veeam_repository" "example" {
  name = "Example-Repository"
  path = "${var.repository_path}/example"
  type = "linux"
}

resource "veeam_credential" "admin" {
  name     = "Admin-Credential"
  username = "admin"
  password = var.admin_password
  type     = "linux"
}
```

## Best Practices

- Use descriptive names for long-term maintainability.
- Keep secrets in variables or environment variables.
- Run `terraform plan` before `terraform apply`.
- Import existing infrastructure to avoid duplicate object errors.

## See Also

- [Data Sources Documentation](../data-sources/)
- [Examples](../../examples/)
- [Provider Configuration](../../README.md#provider-configuration)
