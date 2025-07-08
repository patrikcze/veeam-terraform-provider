# Variables for Advanced Veeam Provider Example

# Veeam Server Configuration
variable "veeam_host" {
  description = "Veeam Backup & Replication server hostname or IP address"
  type        = string
  default     = "https://veeam.example.com"
}

variable "veeam_username" {
  description = "Username for Veeam server authentication"
  type        = string
  default     = "admin"
}

variable "veeam_password" {
  description = "Password for Veeam server authentication"
  type        = string
  sensitive   = true
}

variable "veeam_insecure" {
  description = "Skip TLS certificate verification"
  type        = bool
  default     = false
}

# Linux Credentials
variable "linux_admin_username" {
  description = "Linux administrator username"
  type        = string
  default     = "root"
}

variable "linux_admin_password" {
  description = "Linux administrator password"
  type        = string
  sensitive   = true
}

# Windows Credentials
variable "windows_domain" {
  description = "Windows domain name"
  type        = string
  default     = "DOMAIN"
}

variable "windows_admin_username" {
  description = "Windows administrator username"
  type        = string
  default     = "administrator"
}

variable "windows_admin_password" {
  description = "Windows administrator password"
  type        = string
  sensitive   = true
}

# vCenter Credentials
variable "vcenter_admin_username" {
  description = "vCenter administrator username"
  type        = string
  default     = "administrator@vsphere.local"
}

variable "vcenter_admin_password" {
  description = "vCenter administrator password"
  type        = string
  sensitive   = true
}
