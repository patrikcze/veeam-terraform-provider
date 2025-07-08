# Veeam Terraform Provider Examples

This directory contains comprehensive examples demonstrating various usage patterns of the Veeam Terraform Provider.

## Structure

The examples are organized into the following directories:

### [Basic](basic/)
Demonstrates fundamental usage patterns including:
- Basic provider configuration
- Simple resource creation (repositories, backup jobs, credentials)
- Basic variable usage
- Essential outputs

**Best for**: Getting started with the provider, learning basic concepts

### [Advanced](advanced/)
Shows more complex scenarios including:
- Multiple resource types working together
- Complex credential management (Windows, Linux, vCenter)
- Resource dependencies and relationships
- Advanced variable usage
- Comprehensive outputs and data processing

**Best for**: Production environments, complex backup strategies

### [Data Sources](data-sources/)
Focuses on querying existing Veeam resources:
- Using data sources to query backup jobs and repositories
- Filtering and processing data
- Conditional resource creation based on existing resources
- Complex data analysis with locals
- Comprehensive reporting outputs

**Best for**: Integration with existing Veeam environments, monitoring and reporting

## Common Prerequisites

All examples require:
- Terraform >= 1.0
- Access to a Veeam Backup & Replication server (version 11.0 or later)
- Valid credentials for the Veeam server
- Network connectivity to the Veeam REST API (typically port 9419)

## Security Best Practices

When using these examples:

1. **Use Environment Variables**: Store sensitive information in environment variables rather than hardcoding in files
2. **Terraform Variables**: Use `terraform.tfvars` files for configuration (add to `.gitignore`)
3. **State Management**: Use remote state backends for production deployments
4. **Credential Management**: Consider using external credential management systems

## Environment Variables

You can set the following environment variables to avoid hardcoding sensitive information:

```bash
export VEEAM_HOST="veeam.example.com"
export VEEAM_USERNAME="admin"
export VEEAM_PASSWORD="your-password"
export VEEAM_INSECURE="false"
```

## Quick Start

1. Choose an example directory
2. Copy the example to your working directory
3. Create a `terraform.tfvars` file with your configuration
4. Run the standard Terraform workflow:

```bash
terraform init
terraform plan
terraform apply
```

## Common Issues and Solutions

### TLS Certificate Verification
If you encounter TLS certificate errors:
- Set `insecure = true` in your provider configuration (development only)
- Use proper SSL certificates in production

### API Access
Ensure your Veeam server has the REST API enabled and accessible:
- Check firewall rules (port 9419)
- Verify user permissions
- Test API access with curl or similar tools

### Resource Naming
Veeam has specific naming requirements:
- Resource names must be unique within the Veeam environment
- Some characters may not be allowed in names
- Check Veeam documentation for naming conventions

## Contributing

If you have additional examples or improvements:

1. Follow the existing structure and documentation patterns
2. Include a comprehensive README in your example directory
3. Test your examples thoroughly
4. Submit a pull request with clear descriptions

## Support

For issues related to these examples:
- Check the main project documentation
- Review the Veeam API documentation
- Open an issue in the main project repository

## License

These examples are provided under the same license as the main project (MPL-2.0).
