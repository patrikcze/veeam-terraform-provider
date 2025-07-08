package tests

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccWorkflow_CredentialRepositoryBackupJob(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Skipping acceptance test - set TF_ACC=1 to run")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowConfig_full(),
				Check: resource.ComposeTestCheckFunc(
					// Check credential
					resource.TestCheckResourceAttr("veeam_credential.example", "name", "workflow-credential"),
					resource.TestCheckResourceAttr("veeam_credential.example", "type", "linux"),
					resource.TestCheckResourceAttrSet("veeam_credential.example", "id"),

					// Check repository
					resource.TestCheckResourceAttr("veeam_repository.example", "name", "workflow-repository"),
					resource.TestCheckResourceAttr("veeam_repository.example", "type", "local"),
					resource.TestCheckResourceAttrSet("veeam_repository.example", "id"),

					// Check backup job
					resource.TestCheckResourceAttr("veeam_backup_job.example", "name", "workflow-backup-job"),
					resource.TestCheckResourceAttr("veeam_backup_job.example", "enabled", "true"),
				),
			},
		},
	})
}

func testAccWorkflowConfig_full() string {
	return `
# Provider configuration
terraform {
  required_providers {
    veeam = {
      source = "patrikcze/veeam"
    }
  }
}

provider "veeam" {
  host     = var.veeam_host
  username = var.veeam_username
  password = var.veeam_password
  insecure = true
}

# Variables for configuration
variable "veeam_host" {
  description = "Veeam server host"
  type        = string
  default     = "https://veeam-server:9419"
}

variable "veeam_username" {
  description = "Veeam username"
  type        = string
  default     = "administrator"
}

variable "veeam_password" {
  description = "Veeam password"
  type        = string
  sensitive   = true
}

# Create a credential for backup operations
resource "veeam_credential" "example" {
  name        = "workflow-credential"
  description = "Credential for workflow testing"
  username    = "backup-user"
  password    = "secure-password"
  type        = "linux"
}

# Create a repository for backup storage
resource "veeam_repository" "example" {
  name        = "workflow-repository"
  description = "Repository for workflow testing"
  path        = "/backup/workflow"
  type        = "local"
  capacity    = 1073741824  # 1GB
}

# Create a backup job using the credential and repository
resource "veeam_backup_job" "example" {
  name    = "workflow-backup-job"
  enabled = true
  
  # Note: In a real implementation, you might reference the credential and repository
  # credential_id = veeam_credential.example.id
  # repository_id = veeam_repository.example.id
}

# Outputs for verification
output "credential_id" {
  value = veeam_credential.example.id
}

output "repository_id" {
  value = veeam_repository.example.id
}

output "backup_job_name" {
  value = veeam_backup_job.example.name
}
`
}
