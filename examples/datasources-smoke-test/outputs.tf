# ──────────────────────────────────────────────
# Singleton outputs
# ──────────────────────────────────────────────

output "server_info" {
  description = "VBR server details."
  value = {
    installation_id = data.veeam_server_info.this.installation_id
    server_name     = data.veeam_server_info.this.server_name
    build_number    = data.veeam_server_info.this.build_number
    version         = data.veeam_server_info.this.version
  }
}

output "server_time" {
  description = "Current server time."
  value = {
    server_time = data.veeam_server_time.this.server_time
    time_zone   = data.veeam_server_time.this.time_zone
    utc_offset  = data.veeam_server_time.this.utc_offset
  }
}

output "server_certificate" {
  description = "Active TLS certificate on the VBR server."
  value = {
    subject       = data.veeam_server_certificate.this.subject
    issued_by     = data.veeam_server_certificate.this.issued_by
    thumbprint    = data.veeam_server_certificate.this.thumbprint
    valid_from    = data.veeam_server_certificate.this.valid_from
    valid_to      = data.veeam_server_certificate.this.valid_to
    serial_number = data.veeam_server_certificate.this.serial_number
  }
}

output "license" {
  description = "License information."
  value = {
    type                  = data.veeam_license.this.type
    status                = data.veeam_license.this.status
    licensed_to           = data.veeam_license.this.licensed_to
    expiration_date       = data.veeam_license.this.expiration_date
    licensed_instances    = data.veeam_license.this.licensed_instances
    consumed_instances    = data.veeam_license.this.consumed_instances
    licensed_sockets      = data.veeam_license.this.licensed_sockets
    consumed_sockets      = data.veeam_license.this.consumed_sockets
    licensed_capacity_tb  = data.veeam_license.this.licensed_capacity_tb
    consumed_capacity_tb  = data.veeam_license.this.consumed_capacity_tb
  }
}

output "security_analyzer_summary" {
  description = "Security best-practice check results."
  value = {
    last_run_time   = data.veeam_security_analyzer.this.last_run_time
    last_run_status = data.veeam_security_analyzer.this.last_run_status
    check_count     = length(data.veeam_security_analyzer.this.best_practices)
    failed_checks   = [
      for bp in data.veeam_security_analyzer.this.best_practices : bp.name
      if bp.status != "Success"
    ]
  }
}

# ──────────────────────────────────────────────
# Infrastructure outputs
# ──────────────────────────────────────────────

output "managed_server_names" {
  description = "Names of all managed servers registered in VBR."
  value       = [for s in data.veeam_managed_servers.all.servers : s.name]
}

output "managed_server_count" {
  value = length(data.veeam_managed_servers.all.servers)
}

output "repository_names" {
  description = "Names of all backup repositories."
  value       = [for r in data.veeam_repositories.all.repositories : r.name]
}

output "repository_states" {
  description = "Repository capacity/free-space overview."
  value = [
    for s in data.veeam_repository_states.all.states : {
      name       = s.name
      status     = s.status
      capacity   = s.capacity
      free_space = s.free_space
      used_space = s.used_space
    }
  ]
}

output "proxy_names" {
  description = "Names of all configured backup proxies."
  value       = [for p in data.veeam_proxies.all.proxies : p.name]
}

output "proxy_states" {
  description = "Online/offline status of each proxy."
  value = [
    for s in data.veeam_proxy_states.all.states : {
      name   = s.name
      type   = s.type
      status = s.status
    }
  ]
}

output "wan_accelerator_names" {
  description = "Names of all WAN accelerators."
  value       = [for w in data.veeam_wan_accelerators.all.accelerators : w.name]
}

output "credential_count" {
  description = "Number of stored credentials."
  value       = length(data.veeam_credentials.all.credentials)
}

output "vbr_services" {
  description = "Status of all VBR services."
  value = [
    for svc in data.veeam_services.all.services : {
      name    = svc.name
      status  = svc.status
      version = svc.version
    }
  ]
}

# ──────────────────────────────────────────────
# Jobs & sessions outputs
# ──────────────────────────────────────────────

output "backup_job_names" {
  description = "Names of all configured backup jobs."
  value       = [for j in data.veeam_backup_jobs.all.backup_jobs : j.name]
}

output "running_jobs" {
  description = "Names of jobs currently running."
  value = [
    for j in data.veeam_job_states.all.states : j.name
    if j.status == "Running"
  ]
}

output "job_state_summary" {
  description = "Last result for every job."
  value = [
    for j in data.veeam_job_states.all.states : {
      name        = j.name
      status      = j.status
      last_result = j.last_result
      last_run    = j.last_run
    }
  ]
}

output "recent_sessions" {
  description = "All sessions with their outcome."
  value = [
    for s in data.veeam_sessions.all.sessions : {
      name          = s.name
      session_type  = s.session_type
      state         = s.state
      result        = s.result
      creation_time = s.creation_time
      end_time      = s.end_time
    }
  ]
}

output "task_session_count" {
  description = "Total number of task sessions."
  value       = length(data.veeam_task_sessions.all.task_sessions)
}

# ──────────────────────────────────────────────
# Backup data outputs
# ──────────────────────────────────────────────

output "backup_names" {
  description = "Names of all backup chains."
  value       = [for b in data.veeam_backups.all.backups : b.name]
}

output "backup_object_count" {
  description = "Number of protected objects across all backups."
  value       = length(data.veeam_backup_objects.all.objects)
}

output "restore_point_count" {
  description = "Total number of restore points."
  value       = length(data.veeam_restore_points.all.restore_points)
}

# ──────────────────────────────────────────────
# Replication outputs
# ──────────────────────────────────────────────

output "replica_names" {
  description = "Names of all VM replicas."
  value       = [for r in data.veeam_replicas.all.replicas : r.name]
}

output "replica_point_count" {
  description = "Total number of replica restore points."
  value       = length(data.veeam_replica_points.all.replica_points)
}

# ──────────────────────────────────────────────
# Agent / protection outputs
# ──────────────────────────────────────────────

output "protection_group_names" {
  description = "Names of all protection groups."
  value       = [for pg in data.veeam_protection_groups.all.protection_groups : pg.name]
}

output "protected_computer_count" {
  description = "Number of computers under agent protection."
  value       = length(data.veeam_protected_computers.all.computers)
}

output "protected_computers_by_status" {
  description = "Protected computers grouped by status."
  value = {
    healthy  = [for c in data.veeam_protected_computers.all.computers : c.name if c.status == "Healthy"]
    warning  = [for c in data.veeam_protected_computers.all.computers : c.name if c.status == "Warning"]
    error    = [for c in data.veeam_protected_computers.all.computers : c.name if c.status == "Error"]
    inactive = [for c in data.veeam_protected_computers.all.computers : c.name if c.status == "Inactive"]
  }
}

# ──────────────────────────────────────────────
# Security & compliance outputs
# ──────────────────────────────────────────────

output "security_role_names" {
  description = "Names of all security roles defined in VBR."
  value       = [for r in data.veeam_security_roles.all.roles : r.name]
}

output "security_user_logins" {
  description = "Logins of all VBR users."
  value       = [for u in data.veeam_security_users.all.users : u.login]
}

output "malware_event_count" {
  description = "Number of malware detection events."
  value       = length(data.veeam_malware_events.all.events)
}

output "active_malware_events" {
  description = "Malware events that are still active."
  value = [
    for e in data.veeam_malware_events.all.events : {
      name           = e.name
      severity       = e.severity
      detection_time = e.detection_time
    }
    if e.state == "Active"
  ]
}
