---
page_title: "veeam_cloud_credential Resource - terraform-provider-veeam"
subcategory: ""
description: |-
  Manages a Veeam cloud credential (AWS, Azure, or GCP).
---

# veeam_cloud_credential (Resource)

Manages cloud credentials used by Veeam integrations (AWS, Azure, GCP).

## Example Usage

```hcl
resource "veeam_cloud_credential" "aws" {
  name       = "aws-prod"
  type       = "Amazon"
  access_key = var.aws_access_key
  secret_key = var.aws_secret_key
}

resource "veeam_cloud_credential" "azure_storage" {
  name       = "azure-storage-prod"
  type       = "AzureStorage"
  account    = var.azure_storage_account
  shared_key = var.azure_storage_shared_key
}

resource "veeam_cloud_credential" "azure_compute" {
  name            = "azure-compute-prod"
  type            = "AzureCompute"
  connection_name = "Azure Compute Connection"
  creation_mode   = "ExistingAccount"
  deployment_type = "MicrosoftAzure"
  tenant_id       = var.azure_tenant_id
  application_id  = var.azure_application_id
  application_key = var.azure_application_secret
}
```

## Schema

### Required

- `name` (String) Credential name.
- `type` (String) Cloud credential type: `Amazon`, `AzureStorage`, `AzureCompute`, `Google`, `GoogleService`.

### Optional

- `description` (String) Optional credential description.
- `access_key` (String) Access key for `Amazon` type.
- `account` (String) Account name for `AzureStorage` type.
- `shared_key` (String, Sensitive) Shared key for `AzureStorage` type.
- `connection_name` (String) Connection display name for `AzureCompute` type.
- `creation_mode` (String) Azure compute creation mode. Currently supported: `ExistingAccount`.
- `deployment_type` (String) Azure compute deployment type: `MicrosoftAzure` or `MicrosoftAzureStack`.
- `deployment_region` (String) Optional Azure region for Azure compute deployment payload.
- `account_name` (String) Legacy alias for account/access-key style fields.
- `secret_key` (String, Sensitive) Secret key value.
- `tenant_id` (String) Tenant ID for Azure-style credentials.
- `application_id` (String) Application/client ID.
- `application_key` (String, Sensitive) Application/client secret (used as Azure compute `secret` for `ExistingAccount`).
- `project_id` (String) Project ID for project-scoped clouds.
- `service_account` (String, Sensitive) Service account JSON or key material.

### Read-Only

- `id` (String) Cloud credential identifier assigned by Veeam.

## Import

Import by credential ID:

```bash
terraform import veeam_cloud_credential.example "cloud-credential-id-123"
```

## Notes

- Sensitive fields are marked as sensitive in Terraform state output.
- Use variables or environment variables for secret values.
- For `AzureStorage`, the Veeam API requires `account` and `sharedKey`.
- For `AzureCompute`, this resource currently supports `ExistingAccount` flow with tenant/app/secret credentials.

## Azure Storage Acceptance Example

```hcl
resource "veeam_cloud_credential" "azure_storage" {
  name       = "azure-storageaccount"
  type       = "AzureStorage"

  # Either 'account' or 'account_name' can be used as input.
  account_name = var.azure_storage_account

  # Prefer 'shared_key' for AzureStorage.
  shared_key = var.azure_storage_shared_key
}
```

## Troubleshooting

- Error: `Unsupported cloud credential type`  
  Use one of: `Amazon`, `AzureStorage`, `AzureCompute`, `Google`, `GoogleService`.  
  Values like `MicrosoftAzure` are deployment types for Azure compute, not cloud credential `type` values.

- Error: `Type 'AzureStorage' requires 'account' ... and 'shared_key' ...`  
  Provide `account` (or `account_name`) and `shared_key` (or `secret_key`).

- Error mentions Amazon requirements while testing Azure  
  Check your `type` value in Terraform state/plan output. If it shows `Amazon`, update config to `type = "AzureStorage"` and re-apply.
