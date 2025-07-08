# Veeam Terraform Provider - Acceptance Tests

This directory contains acceptance tests for the Veeam Terraform Provider. These tests require a real Veeam Backup & Replication server to run against.

## Prerequisites

### Ubuntu 24.04 Setup

1. **Install Veeam Backup & Replication Server**
   - Follow the official Veeam documentation for Ubuntu 24.04 installation
   - Ensure the API is enabled and accessible

2. **Configure Environment**
   - Run the setup script: `./scripts/setup-ubuntu-test-env.sh`
   - Or manually install Go, Terraform, and other dependencies

3. **Set Environment Variables**
   - Copy `.env.test.example` to `.env.test`
   - Edit `.env.test` with your Veeam server details:
     ```bash
     VEEAM_HOST=https://your-veeam-server.example.com:9419
     VEEAM_USERNAME=your_username
     VEEAM_PASSWORD=your_password
     VEEAM_INSECURE=true
     ```

## Running Tests

### All Acceptance Tests
```bash
make testacc
```

### Individual Resource Tests
```bash
make testacc-credential    # Test credential resources
make testacc-repository    # Test repository resources  
make testacc-backup-job    # Test backup job resources
```

### Manual Test Execution
```bash
# Set environment variables
export VEEAM_HOST=https://your-server:9419
export VEEAM_USERNAME=your_username
export VEEAM_PASSWORD=your_password
export VEEAM_INSECURE=true

# Run acceptance tests
TF_ACC=1 go test -v ./tests -timeout 120m
```

## Test Structure

### Test Files
- `credential_test.go` - Tests for credential resource
- `repository_test.go` - Tests for repository resource
- `backup_job_test.go` - Tests for backup job resource
- `provider_test.go` - Common test configuration and helpers

### Test Types
Each resource has the following test types:
- **Basic** - Basic resource creation and verification
- **Update** - Resource update operations
- **Import** - Resource import functionality
- **Destroy** - Resource cleanup verification

## Environment Variables

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `TF_ACC` | Enable acceptance tests | Yes | - |
| `VEEAM_HOST` | Veeam server URL | Yes | - |
| `VEEAM_USERNAME` | Authentication username | Yes | - |
| `VEEAM_PASSWORD` | Authentication password | Yes | - |
| `VEEAM_INSECURE` | Skip TLS verification | No | `false` |
| `TEST_TIMEOUT` | Test timeout duration | No | `120m` |

## Test Configuration

### Terraform Configuration
Each test includes Terraform configuration snippets that define the resources being tested. These configurations are automatically applied and destroyed during test execution.

### Example Credential Test Configuration
```hcl
resource "veeam_credential" "test" {
  name        = "test-credential"
  description = "Test credential for acceptance tests"
  username    = "testuser"
  password    = "testpass123"
  type        = "linux"
}
```

## Troubleshooting

### Common Issues

1. **Connection Failed**
   - Verify `VEEAM_HOST` URL is correct
   - Check if Veeam API is accessible
   - Ensure credentials are valid

2. **SSL Certificate Issues**
   - Set `VEEAM_INSECURE=true` for self-signed certificates
   - Or configure proper SSL certificates

3. **Test Timeouts**
   - Increase timeout: `TEST_TIMEOUT=180m`
   - Check network connectivity to Veeam server

4. **Resource Already Exists**
   - Clean up any existing test resources manually
   - Ensure unique resource names across test runs

### Debug Mode
Enable verbose logging:
```bash
TF_LOG=DEBUG TF_ACC=1 go test -v ./tests
```

### Manual Cleanup
If tests fail to clean up resources:
```bash
# List existing resources
curl -k -H "Authorization: Bearer $TOKEN" $VEEAM_HOST/api/v1/credentials

# Delete specific resource
curl -k -X DELETE -H "Authorization: Bearer $TOKEN" $VEEAM_HOST/api/v1/credentials/$ID
```

## CI/CD Integration

### GitHub Actions Example
```yaml
name: Acceptance Tests
on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  testacc:
    runs-on: ubuntu-24.04
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    - name: Run acceptance tests
      env:
        TF_ACC: 1
        VEEAM_HOST: ${{ secrets.VEEAM_HOST }}
        VEEAM_USERNAME: ${{ secrets.VEEAM_USERNAME }}
        VEEAM_PASSWORD: ${{ secrets.VEEAM_PASSWORD }}
        VEEAM_INSECURE: true
      run: make testacc
```

## Best Practices

1. **Resource Naming**
   - Use unique prefixes for test resources
   - Include timestamps or random suffixes

2. **Test Isolation**
   - Each test should be independent
   - Clean up resources in destroy functions

3. **Error Handling**
   - Always check for errors in test functions
   - Provide meaningful error messages

4. **Test Coverage**
   - Test all CRUD operations
   - Test edge cases and error conditions
   - Verify resource attributes and states

## Contributing

When adding new acceptance tests:

1. Follow the existing test patterns
2. Include all test types (Basic, Update, Import, Destroy)
3. Use meaningful test names and descriptions
4. Update this README with any new requirements
5. Ensure tests pass in CI/CD environment

## Support

For issues with acceptance tests:
1. Check the troubleshooting section
2. Review test logs for error details
3. Verify Veeam server configuration
4. Open an issue with test output and environment details
