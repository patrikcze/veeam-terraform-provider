package tests

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/patrikcze/terraform-provider-veeam/internal/client"
)

func TestAccRepository_Basic(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Skipping acceptance test - set TF_ACC=1 to run")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists("veeam_repository.test"),
					resource.TestCheckResourceAttr("veeam_repository.test", "name", "test-repository"),
					resource.TestCheckResourceAttr("veeam_repository.test", "description", "Test repository for acceptance tests"),
					resource.TestCheckResourceAttr("veeam_repository.test", "path", "/backup/test"),
					resource.TestCheckResourceAttr("veeam_repository.test", "type", "local"),
					resource.TestCheckResourceAttrSet("veeam_repository.test", "id"),
				),
			},
		},
	})
}

func TestAccRepository_WithCapacity(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Skipping acceptance test - set TF_ACC=1 to run")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_withCapacity(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists("veeam_repository.test"),
					resource.TestCheckResourceAttr("veeam_repository.test", "name", "test-repository-capacity"),
					resource.TestCheckResourceAttr("veeam_repository.test", "path", "/backup/test-capacity"),
					resource.TestCheckResourceAttr("veeam_repository.test", "type", "local"),
					resource.TestCheckResourceAttr("veeam_repository.test", "capacity", "1073741824"), // 1GB
					resource.TestCheckResourceAttrSet("veeam_repository.test", "id"),
				),
			},
		},
	})
}

func TestAccRepository_Update(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Skipping acceptance test - set TF_ACC=1 to run")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists("veeam_repository.test"),
					resource.TestCheckResourceAttr("veeam_repository.test", "name", "test-repository"),
					resource.TestCheckResourceAttr("veeam_repository.test", "description", "Test repository for acceptance tests"),
				),
			},
			{
				Config: testAccRepositoryConfig_updated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists("veeam_repository.test"),
					resource.TestCheckResourceAttr("veeam_repository.test", "name", "test-repository-updated"),
					resource.TestCheckResourceAttr("veeam_repository.test", "description", "Updated description"),
				),
			},
		},
	})
}

func TestAccRepository_Import(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Skipping acceptance test - set TF_ACC=1 to run")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists("veeam_repository.test"),
				),
			},
			{
				ResourceName:      "veeam_repository.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckRepositoryExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Repository ID is set")
		}

		// Get client from environment for testing
		host := os.Getenv("VEEAM_HOST")
		username := os.Getenv("VEEAM_USERNAME")
		password := os.Getenv("VEEAM_PASSWORD")

		client, err := client.NewVeeamClient(host, username, password, false)
		if err != nil {
			return fmt.Errorf("Failed to create client: %s", err)
		}

		var result map[string]interface{}
		err = client.GetJSON(fmt.Sprintf("/repositories/%s", rs.Primary.ID), &result)
		if err != nil {
			return fmt.Errorf("Repository not found: %s", err)
		}

		return nil
	}
}

func testAccCheckRepositoryDestroy(s *terraform.State) error {
	// Get client from environment for testing
	host := os.Getenv("VEEAM_HOST")
	username := os.Getenv("VEEAM_USERNAME")
	password := os.Getenv("VEEAM_PASSWORD")

	client, err := client.NewVeeamClient(host, username, password, false)
	if err != nil {
		return fmt.Errorf("Failed to create client: %s", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "veeam_repository" {
			continue
		}

		var result map[string]interface{}
		err := client.GetJSON(fmt.Sprintf("/repositories/%s", rs.Primary.ID), &result)
		if err == nil {
			return fmt.Errorf("Repository still exists: %s", rs.Primary.ID)
		}
	}

	return nil
}

func testAccRepositoryConfig_basic() string {
	return `
resource "veeam_repository" "test" {
  name        = "test-repository"
  description = "Test repository for acceptance tests"
  path        = "/backup/test"
  type        = "local"
}
`
}

func testAccRepositoryConfig_withCapacity() string {
	return `
resource "veeam_repository" "test" {
  name        = "test-repository-capacity"
  description = "Test repository with capacity"
  path        = "/backup/test-capacity"
  type        = "local"
  capacity    = 1073741824
}
`
}

func testAccRepositoryConfig_updated() string {
	return `
resource "veeam_repository" "test" {
  name        = "test-repository-updated"
  description = "Updated description"
  path        = "/backup/test-updated"
  type        = "local"
}
`
}
