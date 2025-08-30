#!/bin/bash

# Setup script for HCP Terraform Private Provider Publishing
# This script helps set up GPG keys and prepares the release files

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Setup GPG key
setup_gpg() {
    log_info "Setting up GPG key for signing..."

    if [ -f "gpg_publich.key" ] && [ -f "gpg_public.key" ]; then
        log_warning "GPG keys already exist. Skipping generation."
        return
    fi

    echo "You'll need to generate a GPG key for signing releases."
    echo "Run the following commands:"
    echo ""
    echo "1. Generate GPG key:"
    echo "   gpg --full-generate-key"
    echo ""
    echo "2. When prompted:"
    echo "   - Choose: (1) RSA and RSA (default)"
    echo "   - Keysize: 4096"
    echo "   - Expiration: 0 (never expires)"
    echo "   - Name: Your Name"
    echo "   - Email: your-email@hoomandigital.com"
    echo "   - Comment: HCP Terraform Provider Signing"
    echo ""
    echo "3. List your keys:"
    echo "   gpg --list-secret-keys --keyid-format LONG"
    echo ""
    echo "4. Export keys (replace YOUR_KEY_ID with the actual key ID):"
    echo "   gpg --armor --export-secret-keys YOUR_KEY_ID > gpg_publich.key"
    echo "   gpg --armor --export YOUR_KEY_ID > gpg_public.key"
    echo ""
    echo "5. Create passphrase file:"
    echo "   echo 'your-gpg-passphrase' > gpg_passphrase.txt"
    echo ""
    read -p "Press Enter after you've generated and exported your GPG keys..."

    if [ ! -f "gpg_publich.key" ] || [ ! -f "gpg_public.key" ]; then
        log_error "GPG key files not found. Please generate them first."
        exit 1
    fi

    log_success "GPG keys are ready"
}

# Build provider binaries
build_binaries() {
    log_info "Building provider binaries..."

    if [ ! -d "dist" ]; then
        mkdir dist
    fi

    # Build for multiple platforms
    platforms=("linux/amd64" "linux/arm64" "darwin/amd64" "darwin/arm64" "windows/amd64")

    for platform in "${platforms[@]}"; do
        os=$(echo $platform | cut -d'/' -f1)
        arch=$(echo $platform | cut -d'/' -f2)

        log_info "Building for $os/$arch..."

        if [ "$os" = "windows" ]; then
            output="dist/terraform-provider-nosana_0.1.0_${os}_${arch}.exe"
        else
            output="dist/terraform-provider-nosana_0.1.0_${os}_${arch}"
        fi

        GOOS=$os GOARCH=$arch go build -o "$output" .

        if [ $? -eq 0 ]; then
            log_success "Built $output"
        else
            log_error "Failed to build for $os/$arch"
        fi
    done
}

# Generate SHA256SUMS and sign
generate_checksums() {
    log_info "Generating SHA256SUMS and signatures..."

    cd dist

    # Generate SHA256SUMS
    shasum -a 256 * > ../terraform-provider-nosana_0.1.0_SHA256SUMS
    mv ../terraform-provider-nosana_0.1.0_SHA256SUMS .

    # Sign the SHA256SUMS file
    gpg --detach-sign --armor terraform-provider-nosana_0.1.0_SHA256SUMS

    cd ..
    log_success "Checksums and signatures generated"
}

# Setup HCP Terraform token
setup_token() {
    log_info "Setting up HCP Terraform API token..."

    if [ -z "$HCP_TOKEN" ]; then
        echo ""
        echo "You need an HCP Terraform API token with the following permissions:"
        echo "- Manage Private Registry (or be in the owners team)"
        echo ""
        echo "To create a token:"
        echo "1. Go to: https://app.terraform.io/app/settings/tokens"
        echo "2. Click 'Create a token'"
        echo "3. Give it a name like 'Nosana Provider Publishing'"
        echo "4. Copy the token"
        echo ""
        read -p "Enter your HCP Terraform API token: " -s token
        echo ""
        export HCP_TOKEN="$token"
        echo "export HCP_TOKEN='$token'" >> ~/.bashrc
        log_success "Token saved to environment"
    else
        log_success "HCP_TOKEN already set"
    fi
}

# Main function
main() {
    echo "🔧 Setting up HCP Terraform Private Provider Publishing"
    echo "======================================================="

    setup_gpg
    build_binaries
    generate_checksums
    setup_token

    log_success "Setup complete!"
    echo ""
    echo "🎯 Next steps:"
    echo "1. Verify you're in the owners team OR have 'Manage Private Registry' permissions"
    echo "2. Run: ./publish-publich.sh"
    echo ""
    echo "📖 For users to install:"
    echo "   source = \"app.terraform.io/hoomandigital/nosana\""
}

# Run main function
main "$@"