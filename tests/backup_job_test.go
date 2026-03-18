package tests

// ---------------------------------------------------------------------------
// Acceptance tests for veeam_backup_job resource.
//
// These tests require a live Veeam Backup & Replication V13 server and must be
// activated by setting TF_ACC=1.  All required environment variables:
//   VEEAM_HOST, VEEAM_USERNAME, VEEAM_PASSWORD, VEEAM_INSECURE (optional)
//   TF_VAR_test_repo_id      — UUID of an existing backup repository
//   TF_VAR_test_vcenter_host — hostname of the vCenter Server used in includes
//   TF_VAR_test_vm_name      — display name of a VM that exists in vCenter
//   TF_VAR_test_vm_object_id — MoRef ID of the VM (e.g. "vm-101")
//
// Run with:
//   TF_ACC=1 VEEAM_HOST=... VEEAM_USERNAME=... VEEAM_PASSWORD=... \
//   TF_VAR_test_repo_id=... TF_VAR_test_vcenter_host=... \
//   TF_VAR_test_vm_name=... TF_VAR_test_vm_object_id=... \
//   go test ./tests/ -run TestAccBackupJob -v -timeout 30m
// ---------------------------------------------------------------------------

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/patrikcze/terraform-provider-veeam/internal/client"
)

// ---------------------------------------------------------------------------
// Acceptance test cases
// ---------------------------------------------------------------------------

// TestAccBackupJob_Basic creates a minimal VSphereBackup job with one VM include,
// verifies it was created, then verifies destroy removes it from the server.
func TestAccBackupJob_Basic(t *testing.T) {
	ctx := context.Background()

	if os.Getenv("TF_ACC") == "" {
		t.Skip("Skipping acceptance test — set TF_ACC=1 to run")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckBackupJobDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBackupJobConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBackupJobExists(ctx, "veeam_backup_job.test"),
					resource.TestCheckResourceAttr("veeam_backup_job.test", "name", "tf-acc-backup-job"),
					resource.TestCheckResourceAttr("veeam_backup_job.test", "type", "VSphereBackup"),
					resource.TestCheckResourceAttr("veeam_backup_job.test", "is_disabled", "false"),
				),
			},
		},
	})
}

// TestAccBackupJob_Update verifies that name / description changes are applied
// without recreating the job (type must not change to avoid RequiresReplace).
func TestAccBackupJob_Update(t *testing.T) {
	ctx := context.Background()

	if os.Getenv("TF_ACC") == "" {
		t.Skip("Skipping acceptance test — set TF_ACC=1 to run")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckBackupJobDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBackupJobConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBackupJobExists(ctx, "veeam_backup_job.test"),
					resource.TestCheckResourceAttr("veeam_backup_job.test", "name", "tf-acc-backup-job"),
				),
			},
			{
				Config: testAccBackupJobConfig_updated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBackupJobExists(ctx, "veeam_backup_job.test"),
					resource.TestCheckResourceAttr("veeam_backup_job.test", "name", "tf-acc-backup-job-updated"),
					resource.TestCheckResourceAttr("veeam_backup_job.test", "description", "Updated description"),
				),
			},
		},
	})
}

// TestAccBackupJob_WithSchedule verifies that daily schedule settings are stored
// and read back correctly.
func TestAccBackupJob_WithSchedule(t *testing.T) {
	ctx := context.Background()

	if os.Getenv("TF_ACC") == "" {
		t.Skip("Skipping acceptance test — set TF_ACC=1 to run")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckBackupJobDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBackupJobConfig_withSchedule(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBackupJobExists(ctx, "veeam_backup_job.scheduled"),
					resource.TestCheckResourceAttr("veeam_backup_job.scheduled", "schedule.run_automatically", "true"),
					resource.TestCheckResourceAttr("veeam_backup_job.scheduled", "schedule.daily_enabled", "true"),
					resource.TestCheckResourceAttr("veeam_backup_job.scheduled", "schedule.daily_local_time", "22:00"),
					resource.TestCheckResourceAttr("veeam_backup_job.scheduled", "schedule.daily_kind", "Weekdays"),
				),
			},
		},
	})
}

// TestAccBackupJob_Import verifies that an existing job can be imported via
// `terraform import veeam_backup_job.name <uuid>`.
func TestAccBackupJob_Import(t *testing.T) {
	ctx := context.Background()

	if os.Getenv("TF_ACC") == "" {
		t.Skip("Skipping acceptance test — set TF_ACC=1 to run")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckBackupJobDestroy(ctx),
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
				// schedule and guest_processing may differ after import because some
				// computed fields default on the server side.
				ImportStateVerifyIgnore: []string{
					"schedule",
					"guest_processing",
					"storage.proxy_auto_select",
				},
			},
		},
	})
}

// ---------------------------------------------------------------------------
// Helper functions
// ---------------------------------------------------------------------------

func testAccCheckBackupJobExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found in state: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no Backup Job ID is set for %s", n)
		}

		host := os.Getenv("VEEAM_HOST")
		username := os.Getenv("VEEAM_USERNAME")
		password := os.Getenv("VEEAM_PASSWORD")

		port := 9419
		if p := os.Getenv("VEEAM_PORT"); p != "" {
			port, _ = strconv.Atoi(p)
		}

		c, err := client.NewVeeamClient(ctx, host, port, username, password, false)
		if err != nil {
			return fmt.Errorf("failed to create Veeam client: %s", err)
		}

		var result map[string]any
		endpoint := fmt.Sprintf("/api/v1/jobs/%s", rs.Primary.ID)
		if err := c.GetJSON(ctx, endpoint, &result); err != nil {
			return fmt.Errorf("backup job not found on server (id=%s): %s", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccCheckBackupJobDestroy(ctx context.Context) func(*terraform.State) error {
	return func(s *terraform.State) error {
		host := os.Getenv("VEEAM_HOST")
		username := os.Getenv("VEEAM_USERNAME")
		password := os.Getenv("VEEAM_PASSWORD")

		port := 9419
		if p := os.Getenv("VEEAM_PORT"); p != "" {
			port, _ = strconv.Atoi(p)
		}

		c, err := client.NewVeeamClient(ctx, host, port, username, password, false)
		if err != nil {
			return fmt.Errorf("failed to create Veeam client: %s", err)
		}

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "veeam_backup_job" {
				continue
			}

			var result map[string]any
			endpoint := fmt.Sprintf("/api/v1/jobs/%s", rs.Primary.ID)
			if err := c.GetJSON(ctx, endpoint, &result); err == nil {
				return fmt.Errorf("backup job still exists on server: %s", rs.Primary.ID)
			}
		}

		return nil
	}
}

// ---------------------------------------------------------------------------
// HCL configurations used by the acceptance tests.
//
// Variable references (TF_VAR_*) allow the tests to be run against different
// vCenter / VM combinations without modifying the source code.
// ---------------------------------------------------------------------------

func testAccBackupJobConfig_basic() string {
	repoID := os.Getenv("TF_VAR_test_repo_id")
	vCenterHost := os.Getenv("TF_VAR_test_vcenter_host")
	vmName := os.Getenv("TF_VAR_test_vm_name")
	vmObjectID := os.Getenv("TF_VAR_test_vm_object_id")

	return fmt.Sprintf(`
resource "veeam_backup_job" "test" {
  name        = "tf-acc-backup-job"
  type        = "VSphereBackup"
  description = "Terraform acceptance test backup job"

  virtual_machines {
    includes {
      platform  = "VSphere"
      type      = "VirtualMachine"
      host_name = %q
      name      = %q
      object_id = %q
    }
    exclude_templates = false
  }

  storage {
    repository_id      = %q
    proxy_auto_select  = true
    retention_type     = "RestorePoints"
    retention_quantity = 7
  }
}
`, vCenterHost, vmName, vmObjectID, repoID)
}

func testAccBackupJobConfig_updated() string {
	repoID := os.Getenv("TF_VAR_test_repo_id")
	vCenterHost := os.Getenv("TF_VAR_test_vcenter_host")
	vmName := os.Getenv("TF_VAR_test_vm_name")
	vmObjectID := os.Getenv("TF_VAR_test_vm_object_id")

	return fmt.Sprintf(`
resource "veeam_backup_job" "test" {
  name        = "tf-acc-backup-job-updated"
  type        = "VSphereBackup"
  description = "Updated description"

  virtual_machines {
    includes {
      platform  = "VSphere"
      type      = "VirtualMachine"
      host_name = %q
      name      = %q
      object_id = %q
    }
    exclude_templates = false
  }

  storage {
    repository_id      = %q
    proxy_auto_select  = true
    retention_type     = "RestorePoints"
    retention_quantity = 7
  }
}
`, vCenterHost, vmName, vmObjectID, repoID)
}

func testAccBackupJobConfig_withSchedule() string {
	repoID := os.Getenv("TF_VAR_test_repo_id")
	vCenterHost := os.Getenv("TF_VAR_test_vcenter_host")
	vmName := os.Getenv("TF_VAR_test_vm_name")
	vmObjectID := os.Getenv("TF_VAR_test_vm_object_id")

	return fmt.Sprintf(`
resource "veeam_backup_job" "scheduled" {
  name        = "tf-acc-backup-scheduled"
  type        = "VSphereBackup"
  description = "Terraform acceptance test with schedule"

  virtual_machines {
    includes {
      platform  = "VSphere"
      type      = "VirtualMachine"
      host_name = %q
      name      = %q
      object_id = %q
    }
    exclude_templates = false
  }

  storage {
    repository_id      = %q
    proxy_auto_select  = true
    retention_type     = "RestorePoints"
    retention_quantity = 14
  }

  schedule {
    run_automatically = true
    daily_enabled     = true
    daily_local_time  = "22:00"
    daily_kind        = "Weekdays"
    retry_enabled     = true
    retry_count       = 3
    retry_await_minutes = 10
  }
}
`, vCenterHost, vmName, vmObjectID, repoID)
}
