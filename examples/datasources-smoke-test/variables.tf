variable "veeam_host" {
  description = "Hostname or IP of the Veeam Backup & Replication server."
  type        = string
}

variable "veeam_username" {
  description = "VBR username."
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
