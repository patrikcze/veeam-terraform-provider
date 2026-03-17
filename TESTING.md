# Testing Guide for Veeam Terraform Provider

This document describes the testing setup and approach for the Veeam Terraform Provider.

## Test Structure

The tests are organized following Go testing conventions with `_test.go` suffix files placed alongside the source code:

```
├── internal/
│   ├── client/
│   │   ├── client_test.go         # VeeamClient tests
│   │   └── rest_test.go           # REST API methods tests
│   ├── utils/
│   │   └── retry_test.go          # Retry logic tests
│   └── test_helpers.go            # Common test utilities
├── pkg/
│   ├── datasources/
│   │   ├── backup_jobs_test.go    # Backup jobs data source tests
│   │   └── repositories_test.go   # Repositories data source tests
│   └── resources/
│       ├── backup_job_test.go     # Backup job resource tests
│       ├── cloud_credential_test.go # Cloud credential resource tests
│       ├── configuration_backup_test.go # Configuration backup resource tests
│       ├── credential_test.go     # Credential resource tests
│       ├── repository_test.go      # Repository resource tests
│       └── scale_out_repository_test.go # Scale-out repository tests
│   └── datasources/
│       ├── helpers_test.go        # Shared datasource parsing/filter tests
│       └── new_datasources_smoke_test.go # Tier 2 datasource metadata smoke tests
```

## Testing Approach

### 1. Mocking Strategy

All tests use `testify/mock` to create mock implementations of the VeeamClient:

```go
type MockVeeamClient struct {
    mock.Mock
}

func (m *MockVeeamClient) GetJSON(ctx context.Context, endpoint string, result interface{}) error {
    args := m.Called(ctx, endpoint, result)
    return args.Error(0)
}
```

### 2. Test Categories

#### Unit Tests for Resources
- **Payload Tests**: Test creation of API payloads from Terraform models
- **Model Tests**: Test Terraform model structures and data validation
- **CRUD Logic Tests**: Test individual Create, Read, Update, Delete operations

#### Unit Tests for Data Sources
- **API Response Tests**: Test processing of API responses into Terraform models
- **Filtering Tests**: Test data filtering logic (by ID, name, etc.)
- **Model Structure Tests**: Test data source model structures

#### Integration Tests for Client
- **Authentication Tests**: Test token management and refresh logic
- **HTTP Method Tests**: Test GET, POST, PUT, DELETE operations
- **Error Handling Tests**: Test API error scenarios and retry logic

### 3. Test Examples

#### Resource Test Example
```go
func TestBackupJob_CreatePayload(t *testing.T) {
    // Setup mock client
    mockClient := new(MockVeeamClient)
    
    // Mock API response
    mockClient.On("PostJSON", "/backupJobs", mock.Anything, mock.Anything).Return(nil)
    
    // Create test data
    data := BackupJobModel{
        Name:    types.StringValue("test_backup"),
        Enabled: types.BoolValue(true),
    }
    
    // Test payload creation
    payload := map[string]interface{}{
        "name":    data.Name.ValueString(),
        "enabled": data.Enabled.ValueBool(),
    }
    
    // Execute and assert
    var result map[string]interface{}
    err := mockClient.PostJSON("/backupJobs", payload, &result)
    
    assert.NoError(t, err)
    mockClient.AssertExpectations(t)
}
```

## Running Tests

### Run All Tests
```bash
go test ./...
```

### Run Tests with Verbose Output
```bash
go test -v ./...
```

### Run Tests with Coverage
```bash
go test -cover ./...
```

### Run Specific Package Tests
```bash
go test ./pkg/resources
go test ./pkg/datasources
go test ./internal/client
```

### Run Unit Tests Only (No Acceptance)
```bash
go test ./internal/... ./pkg/...
```

### Run Specific Test Function
```bash
go test -run TestBackupJob_CreatePayload ./pkg/resources
```

## Test Dependencies

The tests use the following external dependencies:

- `github.com/stretchr/testify/assert` - Assertions
- `github.com/stretchr/testify/mock` - Mocking framework
- `github.com/stretchr/testify/require` - Required assertions
- `github.com/hashicorp/terraform-plugin-framework` - Terraform SDK testing utilities

## Coverage Goals

The test suite aims to achieve:

- **90%+ code coverage** for all CRUD operations
- **100% coverage** for API client methods
- **80%+ coverage** for data source read operations
- **100% coverage** for retry and error handling logic

## Best Practices

1. **Use descriptive test names** that clearly indicate what is being tested
2. **Mock external dependencies** to ensure tests are isolated and fast
3. **Test both success and failure scenarios** for robust error handling
4. **Keep tests focused** on single responsibilities
5. **Use table-driven tests** for testing multiple scenarios
6. **Document complex test scenarios** with clear comments

## Continuous Integration

All tests are automatically run on:
- Pull requests
- Merges to main branch
- Tagged releases

The CI pipeline ensures that:
- All tests pass
- Code coverage meets minimum thresholds
- No race conditions exist
- Tests run on multiple Go versions

## Adding New Tests

When adding new features:

1. Create corresponding test files with `_test.go` suffix
2. Use the existing mock patterns for consistency
3. Test all CRUD operations for resources
4. Test all read operations for data sources
5. Include both positive and negative test cases
6. Update this documentation if needed

## Troubleshooting

Common issues and solutions:

### Mock Expectations Not Met
```
Error: mock: Expected call to GetJSON(...)
```
**Solution**: Ensure all mocked methods are called or use `mock.AssertExpectations(t)`

### Import Cycle Issues
```
Error: import cycle not allowed
```
**Solution**: Move shared test utilities to `internal/test_helpers.go`

### Context Timeout Issues
```
Error: context deadline exceeded
```
**Solution**: Check for goroutine leaks or increase test timeout

For more information, see the [Go testing documentation](https://golang.org/pkg/testing/).

## Live Acceptance Coverage (Real VBR)

The following scenarios were validated against a real Veeam Backup & Replication server (v13, API 1.3-rev1):

### Resources — create/destroy verified

- `veeam_credential` (Linux)
- `veeam_repository` (WinLocal)
- `veeam_proxy` (ViProxy)
- `veeam_configuration_backup`
- `veeam_encryption_password`
- `veeam_cloud_credential` (`AzureStorage` shared key flow)

### Data sources — read verified

- `veeam_server_info`
- `veeam_license`
- `veeam_credentials`
- `veeam_managed_servers`
- `veeam_proxies`
- `veeam_repositories`
- `veeam_repository_states`
- `veeam_backup_jobs`
- `veeam_job_states`
- `veeam_backups`
- `veeam_restore_points`
- `veeam_sessions`
- `veeam_protection_groups`
- `veeam_wan_accelerators`

### Known real-world behaviors to account for

- `configBackup` updates require full-object semantics (GET + merge + PUT behavior).
- `ConfigBackupEncryptionModel` requires both `isEnabled` and `passwordId`.
- Deleting an encryption password can transiently fail with "in use by: Backup Configuration Job" even after disabling configuration backup; retrying destroy shortly after typically succeeds.
- Cloud credentials use discriminator type values exactly as defined by API (`Amazon`, `AzureStorage`, `AzureCompute`, `Google`, `GoogleService`).
