# Resources

This section contains documentation for all Veeam Terraform Provider resources.

## Available Resources

### [veeam_backup_job](backup_job.md)
Manages Veeam backup jobs. Use this resource to create, update, and delete backup jobs in your Veeam environment.

**Key features:**
- Create and manage backup jobs
- Enable/disable backup jobs
- Import existing backup jobs

### [veeam_repository](repository.md)
Manages Veeam backup repositories. Use this resource to create, update, and delete backup repositories for storing backup data.

**Key features:**
- Create Linux and Windows repositories
- Configure repository capacity limits
- Manage repository paths and types
- Import existing repositories

### [veeam_credential](credential.md)
Manages Veeam credentials. Use this resource to create, update, and delete authentication credentials for various systems.

**Key features:**
- Create Linux, Windows, and standard credentials
- Manage domain credentials
- Secure password handling
- Import existing credentials

## Common Patterns

### Resource Dependencies
Many resources work together. For example, backup jobs typically depend on repositories and credentials:

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
All resources support importing existing Veeam resources:

```bash
# Import by name for backup jobs
terraform import veeam_backup_job.existing "Existing-Job-Name"

# Import by ID for repositories and credentials
terraform import veeam_repository.existing "repo-id-123"
terraform import veeam_credential.existing "cred-id-456"
```

### Variable Usage
Use variables for sensitive information and reusable values:

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

1. **Use descriptive names**: Make resource names clear and consistent
2. **Set dependencies**: Use `depends_on` to ensure proper resource creation order
3. **Use variables**: Store sensitive information in variables, not hardcoded values
4. **Plan before apply**: Always run `terraform plan` before applying changes
5. **Import existing resources**: Import existing Veeam resources to avoid conflicts

## Error Handling

Common errors and solutions:

- **Resource already exists**: Import the existing resource or choose a different name
- **Authentication failed**: Check your provider configuration and credentials
- **Path not accessible**: Ensure the specified paths exist and are accessible
- **Insufficient permissions**: Verify your Veeam user has the necessary permissions

## See Also

- [Data Sources Documentation](../data-sources/)
- [Examples](../../examples/)
- [Provider Configuration](../../README.md#provider-configuration)
