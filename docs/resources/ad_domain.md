---
page_title: "veeam_ad_domain Resource - terraform-provider-veeam"
subcategory: ""
description: |-
  Manages an Active Directory domain registration in Veeam Backup & Replication.
---

# veeam_ad_domain (Resource)

Manages an Active Directory domain registration in Veeam Backup & Replication (`/api/v1/adDomains`).

This resource supports **Create, Read, and Delete** only. Changing `name` or `username` forces a destroy and recreate.

## Example Usage

```hcl
resource "veeam_ad_domain" "corp" {
  name        = "corp.example.com"
  username    = "CORP\\veeam-svc"
  password    = var.ad_password
  description = "Corporate Active Directory domain"
}
```

## Schema

### Required

- `name` (String) Fully qualified domain name (e.g. `corp.example.com`). Changing this forces a new resource.
- `username` (String) Domain administrator account (e.g. `DOMAIN\\user`). Changing this forces a new resource.
- `password` (String, Sensitive) Domain account password. Write-only — never read back from the API.

### Optional

- `description` (String) Optional description.

### Read-Only

- `id` (String) AD domain identifier (assigned by the server).

## Import

```bash
terraform import veeam_ad_domain.corp <domain-id>
```

## Notes

- `password` is write-only. The API never returns it.
- There is no Update API for AD domain registrations. Changing `name` or `username` replaces the resource entirely.
