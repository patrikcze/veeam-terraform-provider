package tests

// ---------------------------------------------------------------------------
// Acceptance tests for veeam_repository resource.
//
// These tests require a live Veeam Backup & Replication V13 server and must be
// activated by setting TF_ACC=1.  All required environment variables:
//   VEEAM_HOST, VEEAM_USERNAME, VEEAM_PASSWORD, VEEAM_INSECURE (optional)
//   TF_VAR_test_host_id  — UUID of a managed Linux server already registered in VBR
//   TF_VAR_test_repo_path — (optional) filesystem path on that server, defaults to /tmp/tf-acc-repo
//
// Run with:
//   TF_ACC=1 VEEAM_HOST=... VEEAM_USERNAME=... VEEAM_PASSWORD=... \
//   TF_VAR_test_host_id=<linux-server-uuid> \
//   go test ./tests/ -run TestAccRepository -v -timeout 15m
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
// Pre-check helpers
// ---------------------------------------------------------------------------

func testAccPreCheckRepository(t *testing.T) {
	testAccPreCheck(t)
	if v := os.Getenv("TF_VAR_test_host_id"); v == "" {
		t.Fatal("TF_VAR_test_host_id must be set for repository acceptance tests")
	}
}

// ---------------------------------------------------------------------------
// Acceptance test cases
// ---------------------------------------------------------------------------

func TestAccRepository_Basic(t *testing.T) {
	ctx := context.Background()

	if os.Getenv("TF_ACC") == "" {
		t.Skip("Skipping acceptance test — set TF_ACC=1 to run")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckRepository(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, "veeam_repository.test"),
					resource.TestCheckResourceAttr("veeam_repository.test", "name", "tf-acc-repository"),
					resource.TestCheckResourceAttr("veeam_repository.test", "type", "LinuxLocal"),
					resource.TestCheckResourceAttrSet("veeam_repository.test", "id"),
				),
			},
		},
	})
}

func TestAccRepository_Update(t *testing.T) {
	ctx := context.Background()

	if os.Getenv("TF_ACC") == "" {
		t.Skip("Skipping acceptance test — set TF_ACC=1 to run")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckRepository(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, "veeam_repository.test"),
					resource.TestCheckResourceAttr("veeam_repository.test", "name", "tf-acc-repository"),
					resource.TestCheckResourceAttr("veeam_repository.test", "description", "Terraform acceptance test repository"),
				),
			},
			{
				Config: testAccRepositoryConfig_updated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, "veeam_repository.test"),
					resource.TestCheckResourceAttr("veeam_repository.test", "name", "tf-acc-repository-updated"),
					resource.TestCheckResourceAttr("veeam_repository.test", "description", "Updated description"),
				),
			},
		},
	})
}

func TestAccRepository_Import(t *testing.T) {
	ctx := context.Background()

	if os.Getenv("TF_ACC") == "" {
		t.Skip("Skipping acceptance test — set TF_ACC=1 to run")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckRepository(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, "veeam_repository.test"),
				),
			},
			{
				ResourceName:      "veeam_repository.test",
				ImportState:       true,
				ImportStateVerify: true,
				// mount server and task limit flags may differ after import
				// (API may return defaults not sent in the original create)
				ImportStateVerifyIgnore: []string{
					"task_limit_enabled",
					"read_write_limit_enabled",
					"read_write_rate",
				},
			},
		},
	})
}

// ---------------------------------------------------------------------------
// Check helpers
// ---------------------------------------------------------------------------

func testAccCheckRepositoryExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found in state: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no Repository ID is set for %s", n)
		}

		host := os.Getenv("VEEAM_HOST")
		username := os.Getenv("VEEAM_USERNAME")
		password := os.Getenv("VEEAM_PASSWORD")
		insecure := os.Getenv("VEEAM_INSECURE") == "true"

		port := 9419
		if p := os.Getenv("VEEAM_PORT"); p != "" {
			port, _ = strconv.Atoi(p)
		}

		c, err := client.NewVeeamClient(ctx, host, port, username, password, insecure)
		if err != nil {
			return fmt.Errorf("failed to create Veeam client: %s", err)
		}

		var result map[string]any
		endpoint := fmt.Sprintf("/api/v1/backupInfrastructure/repositories/%s", rs.Primary.ID)
		if err := c.GetJSON(ctx, endpoint, &result); err != nil {
			return fmt.Errorf("repository not found on server (id=%s): %s", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccCheckRepositoryDestroy(ctx context.Context) func(*terraform.State) error {
	return func(s *terraform.State) error {
		host := os.Getenv("VEEAM_HOST")
		username := os.Getenv("VEEAM_USERNAME")
		password := os.Getenv("VEEAM_PASSWORD")
		insecure := os.Getenv("VEEAM_INSECURE") == "true"

		port := 9419
		if p := os.Getenv("VEEAM_PORT"); p != "" {
			port, _ = strconv.Atoi(p)
		}

		c, err := client.NewVeeamClient(ctx, host, port, username, password, insecure)
		if err != nil {
			return fmt.Errorf("failed to create Veeam client: %s", err)
		}

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "veeam_repository" {
				continue
			}

			var result map[string]any
			endpoint := fmt.Sprintf("/api/v1/backupInfrastructure/repositories/%s", rs.Primary.ID)
			if err := c.GetJSON(ctx, endpoint, &result); err == nil {
				return fmt.Errorf("repository still exists on server: %s", rs.Primary.ID)
			}
		}

		return nil
	}
}

// ---------------------------------------------------------------------------
// HCL configurations used by the acceptance tests.
// ---------------------------------------------------------------------------

func testAccRepositoryConfig_basic() string {
	hostID := os.Getenv("TF_VAR_test_host_id")
	repoPath := os.Getenv("TF_VAR_test_repo_path")
	if repoPath == "" {
		repoPath = "/tmp/tf-acc-repo"
	}

	return fmt.Sprintf(`
resource "veeam_repository" "test" {
  name               = "tf-acc-repository"
  description        = "Terraform acceptance test repository"
  type               = "LinuxLocal"
  host_id            = %q
  path               = %q
  max_task_count     = 2
  task_limit_enabled = true
}
`, hostID, repoPath)
}

func testAccRepositoryConfig_updated() string {
	hostID := os.Getenv("TF_VAR_test_host_id")
	repoPath := os.Getenv("TF_VAR_test_repo_path")
	if repoPath == "" {
		repoPath = "/tmp/tf-acc-repo"
	}

	return fmt.Sprintf(`
resource "veeam_repository" "test" {
  name               = "tf-acc-repository-updated"
  description        = "Updated description"
  type               = "LinuxLocal"
  host_id            = %q
  path               = %q
  max_task_count     = 4
  task_limit_enabled = true
}
`, hostID, repoPath)
}
