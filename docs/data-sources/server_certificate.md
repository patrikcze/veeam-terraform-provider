---
page_title: "veeam_server_certificate Data Source - terraform-provider-veeam"
subcategory: ""
description: |-
  Reads the server TLS certificate details.
---

# veeam_server_certificate (Data Source)

Reads the TLS certificate currently installed on the Veeam Backup & Replication server. Useful for certificate rotation auditing and compliance checks.

## Example Usage

```hcl
data "veeam_server_certificate" "current" {}

output "cert_thumbprint" {
  value = data.veeam_server_certificate.current.thumbprint
}

output "cert_valid_to" {
  value = data.veeam_server_certificate.current.valid_to
}
```

## Schema

### Read-Only

- `id` (String) Always set to `"server-certificate"`.
- `thumbprint` (String) SHA-1 thumbprint of the certificate.
- `subject` (String) Certificate subject distinguished name.
- `issued_by` (String) Certificate issuer distinguished name.
- `valid_from` (String) Certificate validity start date in ISO 8601 format.
- `valid_to` (String) Certificate expiry date in ISO 8601 format.
- `serial_number` (String) Certificate serial number.
