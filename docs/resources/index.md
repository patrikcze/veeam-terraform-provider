---
page_title: "Resources - terraform-provider-veeam"
subcategory: ""
description: |-
  Index of Terraform resources supported by the Veeam provider.
---

# Resources

Reference for all Veeam Terraform Provider resources.

## Available Resources

### [veeam_backup_job](backup_job.md)
Manages backup jobs.

### [veeam_repository](repository.md)
Manages backup repositories.

### [veeam_credential](credential.md)
Manages standard and Linux/Windows credentials.

### [veeam_cloud_credential](cloud_credential.md)
Manages cloud credentials used by Veeam for AWS, Azure, and GCP integrations.

### [veeam_scale_out_repository](scale_out_repository.md)
Manages Veeam scale-out backup repositories.

### [veeam_configuration_backup](configuration_backup.md)
Manages configuration backup settings and can trigger configuration backup runs.

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
