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

output "entra_tenant_id" {
  value = veeam_entra_id_tenant.corporate.id
}
