resource "veeam_ad_domain" "corp" {
  name        = "corp.example.com"
  username    = "CORP\\veeam-svc"
  password    = var.ad_password
  description = "Corporate Active Directory domain"
}

output "ad_domain_id" {
  value = veeam_ad_domain.corp.id
}
