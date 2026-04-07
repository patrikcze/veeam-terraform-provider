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

# ──────────────────────────────────────────────
# Singleton datasources (no collection, no filter)
# ──────────────────────────────────────────────

data "veeam_server_info" "this" {}

data "veeam_server_time" "this" {}

data "veeam_server_certificate" "this" {}

data "veeam_license" "this" {}

data "veeam_security_analyzer" "this" {}

# ──────────────────────────────────────────────
# Infrastructure datasources
# ──────────────────────────────────────────────

data "veeam_managed_servers" "all" {}

data "veeam_repositories" "all" {}

data "veeam_repository_states" "all" {}

data "veeam_proxies" "all" {}

data "veeam_proxy_states" "all" {}

data "veeam_wan_accelerators" "all" {}

data "veeam_credentials" "all" {}

data "veeam_services" "all" {}

# ──────────────────────────────────────────────
# Jobs & sessions
# ──────────────────────────────────────────────

data "veeam_backup_jobs" "all" {}

data "veeam_job_states" "all" {}

data "veeam_sessions" "all" {}

data "veeam_task_sessions" "all" {}

# ──────────────────────────────────────────────
# Backup data
# ──────────────────────────────────────────────

data "veeam_backups" "all" {}

data "veeam_backup_objects" "all" {}

data "veeam_restore_points" "all" {}

# ──────────────────────────────────────────────
# Replication
# ──────────────────────────────────────────────

data "veeam_replicas" "all" {}

data "veeam_replica_points" "all" {}

# ──────────────────────────────────────────────
# Agent / protection
# ──────────────────────────────────────────────

data "veeam_protection_groups" "all" {}

data "veeam_protected_computers" "all" {}

# ──────────────────────────────────────────────
# Security & compliance
# ──────────────────────────────────────────────

data "veeam_security_roles" "all" {}

data "veeam_security_users" "all" {}

data "veeam_malware_events" "all" {}
