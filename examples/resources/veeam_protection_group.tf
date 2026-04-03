resource "veeam_credential" "agent_cred" {
  username        = "backup-agent"
  password        = var.agent_password
  description     = "Agent deployment credential"
  type            = "Linux"
  ssh_port        = 22
  elevate_to_root = true
}

# IndividualComputers — explicit list of hosts
resource "veeam_protection_group" "servers" {
  name        = "Production-Servers"
  description = "Direct-connect production servers"
  type        = "IndividualComputers"

  computers = [
    {
      hostname        = "db01.example.com"
      connection_type = "PermanentCredentials"
      credentials_id  = veeam_credential.agent_cred.id
    },
    {
      hostname        = "app01.example.com"
      connection_type = "PermanentCredentials"
      credentials_id  = veeam_credential.agent_cred.id
    }
  ]

  options = [
    {
      install_backup_agent        = true
      install_cbt_driver          = false
      install_application_plugins = false
      application_plugins         = []
      update_automatically        = true
      reboot_if_required          = false
      distribution_server_id      = null
      distribution_repository_id  = null
    }
  ]
}

output "protection_group_id" {
  value = veeam_protection_group.servers.id
}
