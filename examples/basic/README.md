# Basic Veeam Provider Example

This example demonstrates basic usage of the Veeam Terraform provider, including:

- Provider configuration
- Creating a backup repository
- Creating a backup job
- Creating a credential
- Using variables for configuration

## Prerequisites

- Terraform >= 1.0
- Access to a Veeam Backup & Replication server
- Valid credentials for the Veeam server

## Usage

1. **Set up variables**: Copy `terraform.tfvars.example` to `terraform.tfvars` and update with your values:

```hcl
veeam_host        = "your-veeam-server.com"
veeam_username    = "admin"
veeam_password    = "your-password"
veeam_insecure    = false
backup_username   = "backup"
backup_password   = "backup-password"
```

2. **Initialize Terraform**:

```bash
terraform init
```

3. **Plan the deployment**:

```bash
terraform plan
```

4. **Apply the configuration**:

```bash
terraform apply
```

5. **View outputs**:

```bash
terraform output
```

6. **Clean up resources**:

```bash
terraform destroy
```

## Configuration

The example creates the following resources:

- `veeam_repository.basic_repo` - A basic Linux backup repository
- `veeam_backup_job.basic_job` - A basic backup job
- `veeam_credential.basic_cred` - A basic Linux credential

## Outputs

- `repository_id` - The ID of the created repository
- `backup_job_name` - The name of the created backup job
- `credential_id` - The ID of the created credential

## Security Notes

- Use environment variables or a secure variable management system for passwords
- Consider using Terraform Cloud or Enterprise for state management
- Enable TLS certificate verification in production environments
