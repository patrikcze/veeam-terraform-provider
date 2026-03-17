# Unit Tests Summary - Veeam Terraform Provider

## Scope
This summary covers executable unit tests in `internal/` and `pkg/` only.
Acceptance tests in `tests/` are intentionally excluded when no Veeam server is available.

## Current Unit-Test Inventory

### New Tier 1 Resource Tests
- `pkg/resources/cloud_credential_test.go`
  - `TestCloudCredential_BuildSpec`
  - `TestCloudCredential_SyncFromAPI`
- `pkg/resources/configuration_backup_test.go`
  - `TestConfigurationBackup_PutConfig`
  - `TestConfigurationBackup_TriggerBackup`
- `pkg/resources/credential_test.go`
  - `TestCredential_BuildSpec_Linux`
  - `TestIsCredentialInUseError`
- `pkg/resources/managed_server_test.go`
  - `TestManagedServerBuildSpecLinuxHost`
  - `TestShouldResolveLinuxFingerprint`
  - `TestIsManagedServerNotFound`
- `pkg/resources/scale_out_repository_test.go`
  - `TestScaleOutRepository_SyncFromAPI`
  - `TestScaleOutRepository_ModelValues`

### New Tier 2 / Shared Data Source Tests
- `pkg/datasources/helpers_test.go`
  - `TestFetchList_FromArray`
  - `TestFetchList_FromWrappedData`
  - `TestNormalizeDataSourceID`
- `pkg/datasources/new_datasources_smoke_test.go`
  - `TestNewTier2DataSources_MetadataTypeNames`

### New Model Round-Trip Tests
- `internal/models/models_test.go`
  - `TestCloudCredentialSpec_RoundTrip`
  - `TestConfigurationBackupModel_RoundTrip`
  - `TestManagedServerModel_RoundTrip`

## Execution Result (Unit Tests Only)
Command equivalent: `go test ./internal/... ./pkg/...`

Result from CI tool run in this workspace:
- **Passed:** 121
- **Failed:** 0

## Notes
- These are real `_test.go` files executed by the test runner.
- No acceptance tests were run in this pass because a live Veeam server was unavailable.
