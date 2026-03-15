# Provider connection
variable "veeam_host" {
  description = "Veeam B&R server hostname or IP"
  type        = string
}

variable "veeam_port" {
  description = "Veeam REST API port"
  type        = number
  default     = 9419
}

variable "veeam_username" {
  description = "Veeam admin username"
  type        = string
}

variable "veeam_password" {
  description = "Veeam admin password"
  type        = string
  sensitive   = true
}

variable "veeam_insecure" {
  description = "Skip TLS verification (dev only)"
  type        = bool
  default     = false
}

# vCenter
variable "vcenter_host" {
  description = "vCenter FQDN or IP"
  type        = string
}

variable "vcenter_username" {
  description = "vCenter admin username"
  type        = string
}

variable "vcenter_password" {
  description = "vCenter admin password"
  type        = string
  sensitive   = true
}

variable "vcenter_thumbprint" {
  description = "vCenter TLS certificate thumbprint"
  type        = string
  default     = ""
}

# Linux repo host
variable "linux_repo_host" {
  description = "Linux repository server FQDN or IP"
  type        = string
}

variable "linux_password" {
  description = "Linux backup user password"
  type        = string
  sensitive   = true
}

variable "linux_ssh_fingerprint" {
  description = "SSH host key fingerprint for Linux server"
  type        = string
  default     = ""
}

# Encryption
variable "encryption_password" {
  description = "Backup encryption password"
  type        = string
  sensitive   = true
}
