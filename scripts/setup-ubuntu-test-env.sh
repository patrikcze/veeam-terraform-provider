#!/bin/bash

# Veeam Terraform Provider Test Environment Setup for Ubuntu 24.04
# This script helps set up the test environment for the Veeam Terraform Provider

set -e

echo "Setting up Veeam Terraform Provider test environment on Ubuntu 24.04..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running on Ubuntu 24.04
check_ubuntu_version() {
    if [[ ! -f /etc/os-release ]]; then
        print_error "Cannot determine OS version"
        exit 1
    fi
    
    source /etc/os-release
    if [[ "$ID" != "ubuntu" ]] || [[ "$VERSION_ID" != "24.04" ]]; then
        print_warning "This script is designed for Ubuntu 24.04. Current OS: $ID $VERSION_ID"
        read -p "Continue anyway? (y/N) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
}

# Update system packages
update_system() {
    print_status "Updating system packages..."
    sudo apt update
    sudo apt upgrade -y
}

# Install required dependencies
install_dependencies() {
    print_status "Installing required dependencies..."
    
    # Install basic development tools
    sudo apt install -y \
        curl \
        wget \
        git \
        build-essential \
        software-properties-common \
        apt-transport-https \
        ca-certificates \
        gnupg \
        lsb-release

    # Install Go if not already installed
    if ! command -v go &> /dev/null; then
        print_status "Installing Go..."
        wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
        sudo rm -rf /usr/local/go
        sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
        rm go1.21.5.linux-amd64.tar.gz
        
        # Add Go to PATH
        echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
        export PATH=$PATH:/usr/local/go/bin
    fi

    # Install Terraform
    if ! command -v terraform &> /dev/null; then
        print_status "Installing Terraform..."
        curl -fsSL https://apt.releases.hashicorp.com/gpg | sudo apt-key add -
        sudo apt-add-repository "deb [arch=amd64] https://apt.releases.hashicorp.com $(lsb_release -cs) main"
        sudo apt update
        sudo apt install -y terraform
    fi

    # Install Docker (for potential containerized testing)
    if ! command -v docker &> /dev/null; then
        print_status "Installing Docker..."
        curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
        echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
        sudo apt update
        sudo apt install -y docker-ce docker-ce-cli containerd.io
        sudo usermod -aG docker $USER
    fi
}

# Set up test environment
setup_test_environment() {
    print_status "Setting up test environment..."
    
    # Create .env.test file if it doesn't exist
    if [[ ! -f .env.test ]]; then
        if [[ -f .env.test.example ]]; then
            cp .env.test.example .env.test
            print_status "Created .env.test file from example"
        else
            print_warning ".env.test.example not found, creating basic .env.test"
            cat > .env.test << EOF
# Veeam Terraform Provider Test Environment Configuration
VEEAM_HOST=https://your-veeam-server:9419
VEEAM_USERNAME=your_username
VEEAM_PASSWORD=your_password
VEEAM_INSECURE=true
TEST_TIMEOUT=120m
TEST_VERBOSE=true
EOF
        fi
    fi

    # Make the .env.test file readable only by owner
    chmod 600 .env.test
    
    print_status "Environment file created. Please edit .env.test with your Veeam server details."
}

# Install test dependencies
install_test_dependencies() {
    print_status "Installing Go test dependencies..."
    
    # Install test dependencies
    go mod download
    
    # Install additional testing tools
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest
}

# Verify installation
verify_installation() {
    print_status "Verifying installation..."
    
    # Check Go version
    if go version &> /dev/null; then
        print_status "Go: $(go version)"
    else
        print_error "Go is not installed correctly"
        exit 1
    fi

    # Check Terraform version
    if terraform version &> /dev/null; then
        print_status "Terraform: $(terraform version | head -n1)"
    else
        print_error "Terraform is not installed correctly"
        exit 1
    fi

    # Check Docker version
    if docker --version &> /dev/null; then
        print_status "Docker: $(docker --version)"
    else
        print_warning "Docker is not installed"
    fi

    # Check if we can build the provider
    print_status "Testing provider build..."
    if make build &> /dev/null; then
        print_status "Provider builds successfully"
    else
        print_error "Provider build failed"
        exit 1
    fi
}

# Main execution
main() {
    print_status "Starting Veeam Terraform Provider test environment setup..."
    
    check_ubuntu_version
    update_system
    install_dependencies
    setup_test_environment
    install_test_dependencies
    verify_installation
    
    print_status "Setup complete!"
    print_status "Next steps:"
    echo "1. Edit .env.test with your Veeam server details"
    echo "2. Run 'make setup-test-env' to verify environment"
    echo "3. Run 'make testacc' to execute acceptance tests"
    echo "4. You may need to log out and back in for Docker permissions to take effect"
}

# Run main function
main "$@"
