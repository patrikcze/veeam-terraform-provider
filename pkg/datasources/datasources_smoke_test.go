package datasources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/stretchr/testify/assert"
)

func TestDataSources_MetadataTypeNames(t *testing.T) {
	tests := []struct {
		name     string
		factory  func() datasource.DataSource
		expected string
	}{
		{"server_info", NewServerInfoDataSource, "veeam_server_info"},
		{"sessions", NewSessionsDataSource, "veeam_sessions"},
		{"backups", NewBackupsDataSource, "veeam_backups"},
		{"restore_points", NewRestorePointsDataSource, "veeam_restore_points"},
		{"proxies", NewProxiesDataSource, "veeam_proxies"},
		{"managed_servers", NewManagedServersDataSource, "veeam_managed_servers"},
		{"protection_groups", NewProtectionGroupsDataSource, "veeam_protection_groups"},
		{"wan_accelerators", NewWanAcceleratorsDataSource, "veeam_wan_accelerators"},
		{"repository_states", NewRepositoryStatesDataSource, "veeam_repository_states"},
		{"license", NewLicenseDataSource, "veeam_license"},
		{"job_states", NewJobStatesDataSource, "veeam_job_states"},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			ds := testCase.factory()
			assert.NotNil(t, ds)

			var resp datasource.MetadataResponse
			ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "veeam"}, &resp)
			assert.Equal(t, testCase.expected, resp.TypeName)
		})
	}
}
