---
page_title: "veeam_entra_id_tenant Resource - terraform-provider-veeam"
subcategory: ""
description: |-
  Manages a Microsoft Entra ID (Azure AD) tenant in the Veeam inventory.
---

# veeam_entra_id_tenant (Resource)

Manages a Microsoft Entra ID (Azure AD) tenant in the Veeam backup inventory (`/api/v1/inventory/entraId/tenants`). The tenant registration is used as the target for Microsoft 365 backup jobs.

The `tenant_id` attribute is immutable — changing it forces a destroy and recreate of the resource.

## Example Usage

```hcl
resource "veeam_cloud_credential" "entra_app" {
  name        = "Entra ID App Registration"
  description = "OAuth2 app credential for tenant backup"
  type        = "AzureServiceAccount"
}

resource "veeam_entra_id_tenant" "corporate" {
  name           = "Contoso Azure Tenant"
  description    = "Primary Azure AD tenant for M365 backup"
  tenant_id      = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  credentials_id = veeam_cloud_credential.entra_app.id
}
```

## Schema

### Required

- `name` (String) Display name for this tenant entry.
- `tenant_id` (String) Azure AD / Entra ID tenant GUID (e.g. `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`). Changing this forces a destroy and recreate.
- `credentials_id` (String) UUID of the OAuth2 application credential used to authenticate to this tenant.

### Optional

- `description` (String) Optional description of the tenant entry.

### Read-Only

- `id` (String) Entra ID tenant resource identifier (assigned by the server).

## Import

```bash
terraform import veeam_entra_id_tenant.main <uuid>
```

The UUID can be retrieved from the Veeam console under **Inventory > Microsoft Entra ID**, or via `GET /api/v1/inventory/entraId/tenants`.
