resource "veeam_credential" "nas_cred" {
  username    = "CORP\\svc-nasbackup"
  password    = var.nas_password
  description = "NAS backup credential"
  type        = "Standard"
}

# CIFS/SMB share
resource "veeam_unstructured_data_server" "cifs_share" {
  name           = "Corp File Server"
  description    = "Corporate CIFS file server"
  type           = "CifsShare"
  host_name      = "fileserver.example.com"
  credentials_id = veeam_credential.nas_cred.id
}

# NFS share
resource "veeam_unstructured_data_server" "nfs_share" {
  name      = "NFS Data Store"
  type      = "NfsShare"
  host_name = "nas.example.com"
}

output "cifs_server_id" {
  value = veeam_unstructured_data_server.cifs_share.id
}
