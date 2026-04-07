terraform {
  required_version = ">= 1.6.0"

  required_providers {
    veeam = {
      source  = "patrikcze/veeam"
      version = "0.1.0"
    }
  }
}

provider "veeam" {
  host     = var.veeam_host
  username = var.veeam_username
  password = var.veeam_password
  insecure = var.veeam_insecure
}

# ══════════════════════════════════════════════════════════════════════════════
# TIER 1 — SINGLETON RESOURCES
# Safe to always apply. These only update the VBR server's running configuration;
# destroying them merely removes them from Terraform state (no VBR data is lost).
# ══════════════════════════════════════════════════════════════════════════════

# veeam_general_options — server-wide storage latency + email + syslog
resource "veeam_general_options" "smoke" {
  storage_latency_control_enabled = false
  storage_latency_limit_ms        = 20

  email_notifications_enabled = false
  email_smtp_server           = "smtp.smoke-test.local"
  email_smtp_port             = 25
  email_from                  = "veeam@smoke-test.local"
  email_to                    = "alerts@smoke-test.local"
  email_subject               = "[Veeam] %JobResult% - %JobName% (smoke-test)"

  snmp_notifications_enabled   = false
  syslog_notifications_enabled = false
  syslog_server                = ""
  syslog_port                  = 514
}

# veeam_email_settings — granular SMTP configuration
resource "veeam_email_settings" "smoke" {
  enabled     = false
  smtp_server = "smtp.smoke-test.local"
  port        = 25
  use_ssl     = false
  from        = "veeam@smoke-test.local"
  to          = "alerts@smoke-test.local"
  subject     = "[Veeam] %JobResult% - %JobName%"

  send_on_success    = false
  send_on_warning    = true
  send_on_error      = true
  send_daily_summary = false
}

# veeam_notification_settings — per-channel notification toggles
resource "veeam_notification_settings" "smoke" {
  notify_on_success                = false
  notify_on_warning                = true
  notify_on_error                  = true
  suppress_repeating_notifications = true
  notify_on_last_retry_only        = true

  send_snmp_on_success = false
  send_snmp_on_warning = false
  send_snmp_on_error   = false

  send_syslog_on_success = false
  send_syslog_on_warning = false
  send_syslog_on_error   = false
}

# veeam_traffic_rules — network throttling
resource "veeam_traffic_rules" "smoke" {
  throttling_enabled = false
  throttling_rules   = "[]"
}

# veeam_security_settings — MFA, lockout, inactivity, password expiry
resource "veeam_security_settings" "smoke" {
  require_ssl       = true
  require_mfa       = false
  block_first_login = false

  login_attempt_limit    = 5
  inactivity_timeout_min = 30

  password_expiration_enabled = false
  password_expiration_days    = 90
}

# veeam_configuration_backup — configuration backup schedule + target
resource "veeam_configuration_backup" "smoke" {
  enabled              = false
  restore_points_count = 3
  # repository_id is optional; omit to use the default VBR configuration backup location
}

# veeam_storage_latency — standalone storage latency singleton
resource "veeam_storage_latency" "smoke" {
  latency_limit_ms      = 20
  throttling_io_enabled = false
  stop_jobs_io_enabled  = false
}

# veeam_event_forwarding — syslog / SIEM event forwarding
resource "veeam_event_forwarding" "smoke" {
  enabled = false
}

# veeam_security_analyzer_schedule — periodic best practices scan schedule
resource "veeam_security_analyzer_schedule" "smoke" {
  enabled       = false
  schedule_type = "Weekly"
  day_of_week   = "Sunday"
  start_time    = "02:00"
}

# ══════════════════════════════════════════════════════════════════════════════
# TIER 2 — STANDALONE CRUD RESOURCES
# These create real VBR objects but have no external server dependencies.
# Safe to apply on any VBR instance.
# ══════════════════════════════════════════════════════════════════════════════

# veeam_credential — Windows service account
resource "veeam_credential" "smoke_windows" {
  username    = "SMOKE\\svc-veeam-test"
  password    = "P@ssw0rd-smoke-test!"
  description = "smoke-test: Windows credential"
  type        = "Standard"
}

# veeam_credential — Linux credential
resource "veeam_credential" "smoke_linux" {
  username        = "smoke-linux-user"
  password        = "P@ssw0rd-smoke-test!"
  description     = "smoke-test: Linux credential"
  type            = "Linux"
  ssh_port        = 22
  elevate_to_root = true
  add_to_sudoers  = true
}

# veeam_encryption_password — backup encryption key
resource "veeam_encryption_password" "smoke" {
  hint     = "smoke-test encryption key"
  password = "EncryptMe-smoke-1!"
}

# veeam_kms_server — KMS server entry (no outbound connection attempted on create)
resource "veeam_kms_server" "smoke" {
  name        = "kms-smoke-test.local"
  description = "smoke-test: KMS server"
  hostname    = "kms-smoke-test.local"
  port        = 5696
}

# veeam_security_user — VBR RBAC user (requires a local or AD account)
# NOTE: The login must match a real Windows/AD account accessible to VBR.
# Comment out if no suitable account is available.
# resource "veeam_security_user" "smoke" {
#   login       = "SMOKE\\veeam-readonly"
#   password    = "P@ssw0rd-smoke!"
#   description = "smoke-test: read-only VBR user"
#   role        = "ReadOnlyUser"
# }

# veeam_ad_domain — Active Directory domain registration
# NOTE: Requires a real reachable domain controller.
# Comment out if no AD environment is available.
# resource "veeam_ad_domain" "smoke" {
#   name        = "smoke.local"
#   username    = "SMOKE\\Administrator"
#   password    = "DomainP@ss!"
#   description = "smoke-test: AD domain"
# }

# veeam_recovery_token — agent recovery token
resource "veeam_recovery_token" "smoke" {
  description     = "smoke-test: recovery token"
  expiration_days = 7
}

# ══════════════════════════════════════════════════════════════════════════════
# TIER 3 — INFRASTRUCTURE-BACKED CRUD RESOURCES
# These require real server connections. Controlled via variables:
#   vcenter_host, windows_server, linux_server, nfs_share_path
# Resources are skipped (count = 0) when the corresponding variable is empty.
# ══════════════════════════════════════════════════════════════════════════════

# ── VMware vSphere / Virtual Infrastructure ──────────────────────────────────
#
# veeam_managed_server with type = "ViHost" is the VMware vSphere resource.
# It registers a vCenter Server or standalone ESXi host in VBR's backup
# infrastructure (equivalent to "Add vCenter/ESXi" in the VBR console under
# Backup Infrastructure > Managed Servers > VMware vSphere).
#
# After this is applied, the vCenter/ESXi appears in VBR's Virtual Infrastructure
# tree and all its VMs/datastores become available for backup jobs.

resource "veeam_credential" "vcenter" {
  count       = var.vcenter_host != "" && var.vcenter_password != "" ? 1 : 0
  username    = var.vcenter_username
  password    = var.vcenter_password
  description = "smoke-test: vCenter service account"
  type        = "Standard"
}

# Dedicated VMware vSphere resource — veeam_vsphere_server
# Equivalent to "Add vCenter/ESXi" in VBR console under Backup Infrastructure > VMware vSphere.
resource "veeam_vsphere_server" "vcenter" {
  count                  = var.vcenter_host != "" && var.vcenter_password != "" ? 1 : 0
  name                   = var.vcenter_host
  description            = "smoke-test: VMware vCenter (veeam_vsphere_server)"
  credentials_id         = veeam_credential.vcenter[0].id
  port                   = 443
  certificate_thumbprint = var.vcenter_thumbprint
}

# ── Linux infrastructure ──────────────────────────────────────────────────────

resource "veeam_credential" "linux" {
  count           = var.linux_server != "" && var.linux_password != "" ? 1 : 0
  username        = "smoke-backup"
  password        = var.linux_password
  description     = "smoke-test: Linux credential"
  type            = "Linux"
  ssh_port        = 22
  elevate_to_root = true
  add_to_sudoers  = true
}

resource "veeam_managed_server" "linux" {
  count          = var.linux_server != "" && var.linux_password != "" ? 1 : 0
  name           = var.linux_server
  description    = "smoke-test: Linux managed server"
  type           = "LinuxHost"
  credentials_id = veeam_credential.linux[0].id
}

resource "veeam_repository" "linux" {
  count          = var.linux_server != "" && var.linux_password != "" ? 1 : 0
  name           = "smoke-linux-repo"
  description    = "smoke-test: Linux local repository"
  type           = "LinuxLocal"
  host_id        = veeam_managed_server.linux[0].id
  path           = "/mnt/veeam-smoke"
  max_task_count = 2
}

# ── NFS repository (standalone — no managed server required) ─────────────────

resource "veeam_repository" "nfs" {
  count       = var.nfs_share_path != "" ? 1 : 0
  name        = "smoke-nfs-repo"
  description = "smoke-test: NFS share repository"
  type        = "Nfs"
  share_path  = var.nfs_share_path
}

# ── Windows proxy ─────────────────────────────────────────────────────────────

resource "veeam_credential" "windows" {
  count       = var.windows_server != "" && var.windows_password != "" ? 1 : 0
  username    = "SMOKE\\Administrator"
  password    = var.windows_password
  description = "smoke-test: Windows administrator"
  type        = "Standard"
}

resource "veeam_managed_server" "windows" {
  count          = var.windows_server != "" && var.windows_password != "" ? 1 : 0
  name           = var.windows_server
  description    = "smoke-test: Windows managed server"
  type           = "WindowsHost"
  credentials_id = veeam_credential.windows[0].id
}

resource "veeam_proxy" "vsphere" {
  count                    = var.windows_server != "" && var.windows_password != "" ? 1 : 0
  description              = "smoke-test: vSphere proxy"
  type                     = "ViProxy"
  host_id                  = veeam_managed_server.windows[0].id
  transport_mode           = "Auto"
  failover_to_network      = true
  host_to_proxy_encryption = false
  max_task_count           = 2
}

resource "veeam_mount_server" "smoke" {
  count             = var.windows_server != "" && var.windows_password != "" ? 1 : 0
  managed_server_id = veeam_managed_server.windows[0].id
  description       = "smoke-test: mount server"
  type              = "WinServer"
  credentials_id    = veeam_credential.windows[0].id
}

# ── Scale-out repository (requires at least one extent/repository) ────────────

resource "veeam_scale_out_repository" "smoke" {
  count       = var.nfs_share_path != "" ? 1 : 0
  name        = "smoke-sobr"
  description = "smoke-test: scale-out backup repository"
  policy      = "DataLocality"
  extents     = [veeam_repository.nfs[0].id]
}

# ── Global VM exclusion (vSphere MoRef — requires vCenter connected) ──────────

resource "veeam_global_vm_exclusion" "smoke" {
  count     = var.vcenter_host != "" && var.vcenter_password != "" ? 1 : 0
  name      = "smoke-excluded-vm"
  type      = "VirtualMachine"
  object_id = "vm-smoke-0001"
  host_id   = veeam_vsphere_server.vcenter[0].id
}

# ── Unstructured data server (NAS/object) ────────────────────────────────────

resource "veeam_unstructured_data_server" "smoke" {
  count       = var.nfs_share_path != "" ? 1 : 0
  name        = "smoke-nas-server"
  description = "smoke-test: unstructured data (NAS) server"
  type        = "CifsShare"
  share_path  = var.nfs_share_path
}

# ── Protection group (agent-based computers) ─────────────────────────────────

resource "veeam_protection_group" "smoke" {
  count            = var.windows_server != "" ? 1 : 0
  name             = "smoke-protection-group"
  description      = "smoke-test: Windows computers protection group"
  type             = "IndividualComputers"
  credentials_id   = veeam_credential.windows[0].id
  schedule_enabled = false

  computers {
    host_name = var.windows_server
    type      = "Host"
  }
}

# ── Backup job (requires repository + proxy + vCenter) ───────────────────────
# Only created when all three infrastructure tiers are available.

resource "veeam_backup_job" "smoke" {
  count            = (var.vcenter_host != "" && var.linux_server != "" && var.windows_server != "") ? 1 : 0
  name             = "smoke-backup-job"
  description      = "smoke-test: VMware backup job"
  type             = "VSphereBackup"
  repository_id    = veeam_repository.linux[0].id
  is_enabled       = false
  schedule_enabled = false

  objects {
    type      = "VirtualMachine"
    name      = "smoke-vm-01"
    object_id = "vm-smoke-0001"
    host_name = var.vcenter_host
  }
}
