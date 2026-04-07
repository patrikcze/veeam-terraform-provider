# ──────────────────────────────────────────────
# Tier 1 — Singleton resource IDs (always populated)
# ──────────────────────────────────────────────

output "general_options_id" {
  description = "Singleton ID of veeam_general_options."
  value       = veeam_general_options.smoke.id
}

output "email_settings_id" {
  description = "Singleton ID of veeam_email_settings."
  value       = veeam_email_settings.smoke.id
}

output "notification_settings_id" {
  description = "Singleton ID of veeam_notification_settings."
  value       = veeam_notification_settings.smoke.id
}

output "traffic_rules_id" {
  description = "Singleton ID of veeam_traffic_rules."
  value       = veeam_traffic_rules.smoke.id
}

output "security_settings_id" {
  description = "Singleton ID of veeam_security_settings."
  value       = veeam_security_settings.smoke.id
}

output "configuration_backup_id" {
  description = "Singleton ID of veeam_configuration_backup."
  value       = veeam_configuration_backup.smoke.id
}

output "storage_latency_id" {
  description = "Singleton ID of veeam_storage_latency."
  value       = veeam_storage_latency.smoke.id
}

output "event_forwarding_id" {
  description = "Singleton ID of veeam_event_forwarding."
  value       = veeam_event_forwarding.smoke.id
}

output "security_analyzer_schedule_id" {
  description = "Singleton ID of veeam_security_analyzer_schedule."
  value       = veeam_security_analyzer_schedule.smoke.id
}

# ──────────────────────────────────────────────
# Tier 2 — Standalone CRUD (always created)
# ──────────────────────────────────────────────

output "credential_windows_id" {
  description = "ID of the smoke-test Windows credential."
  value       = veeam_credential.smoke_windows.id
}

output "credential_linux_id" {
  description = "ID of the smoke-test Linux credential."
  value       = veeam_credential.smoke_linux.id
}

output "encryption_password_id" {
  description = "ID of the smoke-test encryption password."
  value       = veeam_encryption_password.smoke.id
}

output "kms_server_id" {
  description = "ID of the smoke-test KMS server."
  value       = veeam_kms_server.smoke.id
}

output "recovery_token_id" {
  description = "ID of the smoke-test recovery token."
  value       = veeam_recovery_token.smoke.id
}

output "recovery_token_value" {
  description = "Token value (sensitive — captured on Create only)."
  value       = veeam_recovery_token.smoke.token_value
  sensitive   = true
}

# ──────────────────────────────────────────────
# Tier 3 — Infrastructure resources (populated when variables are set)
# ──────────────────────────────────────────────

output "vcenter_vsphere_server_id" {
  description = "ID of the registered vCenter (veeam_vsphere_server). Empty if vcenter_host not set."
  value       = length(veeam_vsphere_server.vcenter) > 0 ? veeam_vsphere_server.vcenter[0].id : ""
}

output "vcenter_vsphere_server_status" {
  description = "VBR connection status of the vCenter."
  value       = length(veeam_vsphere_server.vcenter) > 0 ? veeam_vsphere_server.vcenter[0].status : "not created"
}

output "linux_managed_server_id" {
  description = "ID of the Linux managed server (empty if linux_server not set)."
  value       = length(veeam_managed_server.linux) > 0 ? veeam_managed_server.linux[0].id : ""
}

output "linux_repository_id" {
  description = "ID of the Linux local repository (empty if linux_server not set)."
  value       = length(veeam_repository.linux) > 0 ? veeam_repository.linux[0].id : ""
}

output "nfs_repository_id" {
  description = "ID of the NFS repository (empty if nfs_share_path not set)."
  value       = length(veeam_repository.nfs) > 0 ? veeam_repository.nfs[0].id : ""
}

output "windows_managed_server_id" {
  description = "ID of the Windows managed server (empty if windows_server not set)."
  value       = length(veeam_managed_server.windows) > 0 ? veeam_managed_server.windows[0].id : ""
}

output "vsphere_proxy_id" {
  description = "ID of the vSphere proxy (empty if windows_server not set)."
  value       = length(veeam_proxy.vsphere) > 0 ? veeam_proxy.vsphere[0].id : ""
}

output "scale_out_repository_id" {
  description = "ID of the scale-out repository (empty if nfs_share_path not set)."
  value       = length(veeam_scale_out_repository.smoke) > 0 ? veeam_scale_out_repository.smoke[0].id : ""
}

output "protection_group_id" {
  description = "ID of the protection group (empty if windows_server not set)."
  value       = length(veeam_protection_group.smoke) > 0 ? veeam_protection_group.smoke[0].id : ""
}

output "backup_job_id" {
  description = "ID of the backup job (empty if full infrastructure not provided)."
  value       = length(veeam_backup_job.smoke) > 0 ? veeam_backup_job.smoke[0].id : ""
}
