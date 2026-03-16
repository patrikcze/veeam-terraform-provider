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
  name         = "aws-prod"
  type         = "Amazon"
  account_name = var.aws_access_key
  secret_key   = var.aws_secret_key
}
```

## Schema

### Required

- `name` (String) Credential name.
- `type` (String) Cloud credential type (for example `Amazon`, `Azure`, `Google`).

### Optional

- `description` (String) Optional credential description.
- `account_name` (String) Account or access key identifier.
- `secret_key` (String, Sensitive) Secret key value.
- `tenant_id` (String) Tenant ID for Azure-style credentials.
- `application_id` (String) Application/client ID.
- `application_key` (String, Sensitive) Application/client secret.
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
