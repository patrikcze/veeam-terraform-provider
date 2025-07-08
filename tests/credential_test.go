package tests

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/patrikcze/terraform-provider-veeam/internal/client"
)

func TestAccCredential_Basic(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Skipping acceptance test - set TF_ACC=1 to run")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCredentialDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCredentialConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCredentialExists("veeam_credential.test"),
					resource.TestCheckResourceAttr("veeam_credential.test", "name", "test-credential"),
					resource.TestCheckResourceAttr("veeam_credential.test", "description", "Test credential for acceptance tests"),
					resource.TestCheckResourceAttr("veeam_credential.test", "username", "testuser"),
					resource.TestCheckResourceAttr("veeam_credential.test", "type", "linux"),
					resource.TestCheckResourceAttrSet("veeam_credential.test", "id"),
				),
			},
		},
	})
}

func TestAccCredential_WithDomain(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Skipping acceptance test - set TF_ACC=1 to run")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCredentialDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCredentialConfig_withDomain(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCredentialExists("veeam_credential.test"),
					resource.TestCheckResourceAttr("veeam_credential.test", "name", "test-credential-domain"),
					resource.TestCheckResourceAttr("veeam_credential.test", "username", "testuser"),
					resource.TestCheckResourceAttr("veeam_credential.test", "type", "windows"),
					resource.TestCheckResourceAttr("veeam_credential.test", "domain", "TESTDOMAIN"),
					resource.TestCheckResourceAttrSet("veeam_credential.test", "id"),
				),
			},
		},
	})
}

func TestAccCredential_Update(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Skipping acceptance test - set TF_ACC=1 to run")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCredentialDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCredentialConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCredentialExists("veeam_credential.test"),
					resource.TestCheckResourceAttr("veeam_credential.test", "name", "test-credential"),
					resource.TestCheckResourceAttr("veeam_credential.test", "description", "Test credential for acceptance tests"),
				),
			},
			{
				Config: testAccCredentialConfig_updated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCredentialExists("veeam_credential.test"),
					resource.TestCheckResourceAttr("veeam_credential.test", "name", "test-credential-updated"),
					resource.TestCheckResourceAttr("veeam_credential.test", "description", "Updated description"),
				),
			},
		},
	})
}

func TestAccCredential_Import(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Skipping acceptance test - set TF_ACC=1 to run")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCredentialDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCredentialConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCredentialExists("veeam_credential.test"),
				),
			},
			{
				ResourceName:      "veeam_credential.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{"password"}, // Password is not returned by API
			},
		},
	})
}

func testAccCheckCredentialExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Credential ID is set")
		}

		// Get client from environment for testing
		host := os.Getenv("VEEAM_HOST")
		username := os.Getenv("VEEAM_USERNAME")
		password := os.Getenv("VEEAM_PASSWORD")
		
		client, err := client.NewVeeamClient(host, username, password)
		if err != nil {
			return fmt.Errorf("Failed to create client: %s", err)
		}
		
		var result map[string]interface{}
		err = client.GetJSON(fmt.Sprintf("/credentials/%s", rs.Primary.ID), &result)
		if err != nil {
			return fmt.Errorf("Credential not found: %s", err)
		}

		return nil
	}
}

func testAccCheckCredentialDestroy(s *terraform.State) error {
	// Get client from environment for testing
	host := os.Getenv("VEEAM_HOST")
	username := os.Getenv("VEEAM_USERNAME")
	password := os.Getenv("VEEAM_PASSWORD")
	
	client, err := client.NewVeeamClient(host, username, password)
	if err != nil {
		return fmt.Errorf("Failed to create client: %s", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "veeam_credential" {
			continue
		}

		var result map[string]interface{}
		err := client.GetJSON(fmt.Sprintf("/credentials/%s", rs.Primary.ID), &result)
		if err == nil {
			return fmt.Errorf("Credential still exists: %s", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCredentialConfig_basic() string {
	return `
resource "veeam_credential" "test" {
  name        = "test-credential"
  description = "Test credential for acceptance tests"
  username    = "testuser"
  password    = "testpass123"
  type        = "linux"
}
`
}

func testAccCredentialConfig_withDomain() string {
	return `
resource "veeam_credential" "test" {
  name        = "test-credential-domain"
  description = "Test credential with domain"
  username    = "testuser"
  password    = "testpass123"
  type        = "windows"
  domain      = "TESTDOMAIN"
}
`
}

func testAccCredentialConfig_updated() string {
	return `
resource "veeam_credential" "test" {
  name        = "test-credential-updated"
  description = "Updated description"
  username    = "testuser"
  password    = "testpass123"
  type        = "linux"
}
`
}
