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

variable "repository_host_name" {
  description = "Managed server name in VBR used as repository host"
  type        = string
}
