# ──────────────────────────────────────────────
# Provider connection (required)
# ──────────────────────────────────────────────

variable "veeam_host" {
  description = "Hostname or IP of the Veeam Backup & Replication server."
  type        = string
}

variable "veeam_username" {
  description = "VBR username (e.g. DOMAIN\\Administrator)."
  type        = string
}

variable "veeam_password" {
  description = "VBR password."
  type        = string
  sensitive   = true
}

variable "veeam_insecure" {
  description = "Skip TLS certificate verification."
  type        = bool
  default     = true
}

# ──────────────────────────────────────────────
# Infrastructure targets (Tier 3 — server-backed resources)
#
# Set these to real server names in your environment.
# Leave as "" to skip the corresponding resource blocks
# (use count = var.foo != "" ? 1 : 0 pattern in main.tf).
# ──────────────────────────────────────────────

variable "vcenter_host" {
  description = "vCenter or ESXi FQDN/IP (for veeam_managed_server ViHost). Leave empty to skip."
  type        = string
  default     = ""
}

variable "vcenter_thumbprint" {
  description = "SHA-1 thumbprint of the vCenter TLS certificate (required with vcenter_host)."
  type        = string
  default     = ""
}

variable "vcenter_username" {
  description = "vCenter service account username (e.g. VSPHERE.LOCAL\\svc-veeam)."
  type        = string
  default     = "VSPHERE.LOCAL\\svc-veeam"
}

variable "vcenter_password" {
  description = "vCenter service account password."
  type        = string
  sensitive   = true
  default     = ""
}

variable "windows_server" {
  description = "Windows server FQDN/IP (for proxy/repository host). Leave empty to skip."
  type        = string
  default     = ""
}

variable "windows_password" {
  description = "Windows administrator password."
  type        = string
  sensitive   = true
  default     = ""
}

variable "linux_server" {
  description = "Linux server FQDN/IP (for LinuxHost managed server / repository). Leave empty to skip."
  type        = string
  default     = ""
}

variable "linux_password" {
  description = "Linux user password."
  type        = string
  sensitive   = true
  default     = ""
}

variable "nfs_share_path" {
  description = "NFS share path for a standalone NFS repository (e.g. nas.example.com:/export/veeam)."
  type        = string
  default     = ""
}
