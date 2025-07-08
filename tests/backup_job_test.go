package tests

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/patrikcze/terraform-provider-veeam/internal/client"
)

func TestAccBackupJob_Basic(t *testing.T) {
	ctx := context.Background()

	if os.Getenv("TF_ACC") == "" {
		t.Skip("Skipping acceptance test - set TF_ACC=1 to run")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckBackupJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupJobConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBackupJobExists(ctx, "veeam_backup_job.test"),
					resource.TestCheckResourceAttr("veeam_backup_job.test", "name", "test-backup-job"),
					resource.TestCheckResourceAttr("veeam_backup_job.test", "enabled", "true"),
				),
			},
		},
	})
}

func TestAccBackupJob_Disabled(t *testing.T) {
	ctx := context.Background()

	if os.Getenv("TF_ACC") == "" {
		t.Skip("Skipping acceptance test - set TF_ACC=1 to run")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckBackupJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupJobConfig_disabled(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBackupJobExists(ctx, "veeam_backup_job.test"),
					resource.TestCheckResourceAttr("veeam_backup_job.test", "name", "test-backup-job-disabled"),
					resource.TestCheckResourceAttr("veeam_backup_job.test", "enabled", "false"),
				),
			},
		},
	})
}

func TestAccBackupJob_Update(t *testing.T) {
	ctx := context.Background()

	if os.Getenv("TF_ACC") == "" {
		t.Skip("Skipping acceptance test - set TF_ACC=1 to run")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckBackupJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupJobConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBackupJobExists(ctx, "veeam_backup_job.test"),
					resource.TestCheckResourceAttr("veeam_backup_job.test", "name", "test-backup-job"),
					resource.TestCheckResourceAttr("veeam_backup_job.test", "enabled", "true"),
				),
			},
			{
				Config: testAccBackupJobConfig_updated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBackupJobExists(ctx, "veeam_backup_job.test"),
					resource.TestCheckResourceAttr("veeam_backup_job.test", "name", "test-backup-job-updated"),
					resource.TestCheckResourceAttr("veeam_backup_job.test", "enabled", "false"),
				),
			},
		},
	})
}

func TestAccBackupJob_Import(t *testing.T) {
	ctx := context.Background()

	if os.Getenv("TF_ACC") == "" {
		t.Skip("Skipping acceptance test - set TF_ACC=1 to run")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckBackupJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupJobConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBackupJobExists(ctx, "veeam_backup_job.test"),
				),
			},
			{
				ResourceName:      "veeam_backup_job.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckBackupJobExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Backup Job ID is set")
		}

		// Get client from environment for testing
		host := os.Getenv("VEEAM_HOST")
		username := os.Getenv("VEEAM_USERNAME")
		password := os.Getenv("VEEAM_PASSWORD")

		client, err := client.NewVeeamClient(ctx, host, username, password, false)
		if err != nil {
			return fmt.Errorf("Failed to create client: %s", err)
		}

		var result map[string]interface{}
		err = client.GetJSON(ctx, fmt.Sprintf("/backupJobs/%s", rs.Primary.ID), &result)
		if err != nil {
			return fmt.Errorf("Backup Job not found: %s", err)
		}

		return nil
	}
}

func testAccCheckBackupJobDestroy(s *terraform.State) error {
	ctx := context.Background()

	// Get client from environment for testing
	host := os.Getenv("VEEAM_HOST")
	username := os.Getenv("VEEAM_USERNAME")
	password := os.Getenv("VEEAM_PASSWORD")

	client, err := client.NewVeeamClient(ctx, host, username, password, false)
	if err != nil {
		return fmt.Errorf("Failed to create client: %s", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "veeam_backup_job" {
			continue
		}

		var result map[string]interface{}
		err := client.GetJSON(ctx, fmt.Sprintf("/backupJobs/%s", rs.Primary.ID), &result)
		if err == nil {
			return fmt.Errorf("Backup Job still exists: %s", rs.Primary.ID)
		}
	}

	return nil
}

func testAccBackupJobConfig_basic() string {
	return `
resource "veeam_backup_job" "test" {
  name    = "test-backup-job"
  enabled = true
}
`
}

func testAccBackupJobConfig_disabled() string {
	return `
resource "veeam_backup_job" "test" {
  name    = "test-backup-job-disabled"
  enabled = false
}
`
}

func testAccBackupJobConfig_updated() string {
	return `
resource "veeam_backup_job" "test" {
  name    = "test-backup-job-updated"
  enabled = false
}
`
}
