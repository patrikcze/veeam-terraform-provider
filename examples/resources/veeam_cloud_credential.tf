# AWS S3 cloud credential
resource "veeam_cloud_credential" "aws_s3" {
  name        = "AWS-S3-Prod"
  description = "AWS S3 storage credential for capacity tier"
  type        = "AmazonS3"
  access_key  = var.aws_access_key
  secret_key  = var.aws_secret_key
}

# Azure Blob Storage cloud credential
resource "veeam_cloud_credential" "azure_blob" {
  name         = "Azure-Blob-Prod"
  description  = "Azure Blob Storage credential for capacity tier"
  type         = "AzureBlob"
  account_name = "mystorageaccount"
  shared_key   = var.azure_shared_key
}

output "aws_credential_id" {
  value = veeam_cloud_credential.aws_s3.id
}

output "azure_credential_id" {
  value = veeam_cloud_credential.azure_blob.id
}
