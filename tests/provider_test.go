package tests

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/patrikcze/terraform-provider-veeam/internal"
	"github.com/patrikcze/terraform-provider-veeam/internal/client"
)

// testAccProvider is a global variable that stores the provider instance for testing
var testAccProvider *internal.Provider

// testAccProtoV6ProviderFactories are used to instantiate a provider during acceptance testing.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"veeam": providerserver.NewProtocol6WithError(internal.New("test")()),
}

// testAccPreCheck verifies that the necessary test environment variables are set.
func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("VEEAM_HOST"); v == "" {
		t.Fatal("VEEAM_HOST must be set for acceptance tests")
	}
	if v := os.Getenv("VEEAM_USERNAME"); v == "" {
		t.Fatal("VEEAM_USERNAME must be set for acceptance tests")
	}
	if v := os.Getenv("VEEAM_PASSWORD"); v == "" {
		t.Fatal("VEEAM_PASSWORD must be set for acceptance tests")
	}

	// Test the connection to the Veeam server
	if err := testAccProviderConnection(); err != nil {
		t.Fatalf("Failed to connect to Veeam server: %s", err)
	}
}

// testAccProviderConnection tests the connection to the Veeam server
func testAccProviderConnection() error {
	host := os.Getenv("VEEAM_HOST")
	username := os.Getenv("VEEAM_USERNAME")
	password := os.Getenv("VEEAM_PASSWORD")

	client, err := client.NewVeeamClient(host, username, password)
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	// Test a simple API call to verify connection
	// This would typically be a health check or simple endpoint
	// For now, we'll just verify the client was created successfully
	if client == nil {
		return fmt.Errorf("API client is nil")
	}

	return nil
}

// testAccProviderConfigure configures the provider for acceptance testing
func testAccProviderConfigure() error {
	// For now, just create a basic provider instance
	// The actual configuration will be handled by the test framework
	testAccProvider = &internal.Provider{}
	return nil
}

// TestMain runs before all tests to set up the test environment
func TestMain(m *testing.M) {
	// Only run setup for acceptance tests
	if os.Getenv("TF_ACC") != "" {
		// Set up test environment
		if err := testAccProviderConfigure(); err != nil {
			fmt.Printf("Failed to configure test provider: %s\n", err)
			os.Exit(1)
		}
	}

	// Run tests
	exitCode := m.Run()

	// Clean up if needed
	os.Exit(exitCode)
}
