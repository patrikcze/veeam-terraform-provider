# Variables for Basic Veeam Provider Example

variable "veeam_host" {
  description = "Veeam Backup & Replication server hostname or IP address"
  type        = string
  default     = "veeam.example.com"
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

variable "backup_username" {
  description = "Username for backup operations"
  type        = string
  default     = "backup"
}

variable "backup_password" {
  description = "Password for backup operations"
  type        = string
  sensitive   = true
}
