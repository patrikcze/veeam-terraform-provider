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
- `access_key` (String, Sensitive) AWS access key ID (`Amazon` type).
- `account` (String) Storage account name (`AzureStorage` type).
- `shared_key` (String, Sensitive) Storage account shared key (`AzureStorage` type).
- `connection_name` (String) Connection display name (`AzureCompute` type).
- `creation_mode` (String) Azure compute creation mode. Supported value: `ExistingAccount`.
- `deployment_type` (String) Azure compute deployment type: `MicrosoftAzure` or `MicrosoftAzureStack`.
- `deployment_region` (String) Azure region for compute deployment (`AzureCompute` type, optional).
- `account_name` (String) Alias accepted for `account` (`AzureStorage`) or `access_key` (`Amazon`).
- `secret_key` (String, Sensitive) AWS secret access key (`Amazon` type) or alias for `shared_key` (`AzureStorage`).
- `tenant_id` (String, Sensitive) Azure Active Directory tenant ID (`AzureCompute` type).
- `application_id` (String, Sensitive) Azure application (client) ID (`AzureCompute` type).
- `application_key` (String, Sensitive) Azure application client secret (`AzureCompute` type).
- `project_id` (String, Sensitive) Google Cloud project ID (`Google` or `GoogleService` type).
- `service_account` (String, Sensitive) Google service account key JSON (`GoogleService` type).

### Read-Only

- `id` (String) Cloud credential identifier assigned by Veeam.

## Import

Import by credential ID:

```bash
terraform import veeam_cloud_credential.example "cloud-credential-id-123"
```

## Notes

- For `AzureStorage`, provide `account` (or `account_name`) and `shared_key` (or `secret_key`).
- For `AzureCompute`, only the `ExistingAccount` creation mode is supported. Provide `tenant_id`, `application_id`, and `application_key`.
- For `Amazon`, provide `access_key` and `secret_key`.
- For `GoogleService`, provide `service_account`. `project_id` is optional.
- `account_name` is accepted as an alias for `account` (AzureStorage) or `access_key` (Amazon).
- The `type` value must be one of the exact enum strings: `Amazon`, `AzureStorage`, `AzureCompute`, `Google`, `GoogleService`. Values such as `MicrosoftAzure` are deployment types, not credential types.
- The Veeam API does not return secret values on read. Secret fields are preserved from the last applied configuration.
