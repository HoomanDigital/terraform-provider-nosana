#!/bin/bash

# HCP Terraform Public Provider Publishing Script
# This script publishes the Nosana provider to the Public Terraform Registry

set -e

# Configuration
ORG_NAME="hoomandigital"
PROVIDER_NAME="nosana"
VERSION="0.1.0"

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

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."

    if ! command -v curl >/dev/null 2>&1; then
        log_error "curl is required but not installed."
        exit 1
    fi

    if ! command -v jq >/dev/null 2>&1; then
        log_error "jq is required but not installed."
        exit 1
    fi

    if [ -z "$HCP_TOKEN" ]; then
        log_error "HCP_TOKEN environment variable is not set."
        echo "Please set your HCP Terraform API token:"
        echo "export HCP_TOKEN='your-api-token-here'"
        exit 1
    fi

    log_success "Prerequisites check passed"
}

# Create provider
create_provider() {
    log_info "Creating provider in HCP Terraform..."

    response=$(curl -s \
        --header "Authorization: Bearer $HCP_TOKEN" \
        --header "Content-Type: application/vnd.api+json" \
        --request POST \
        --data @provider.json \
        "https://app.terraform.io/api/v2/organizations/$ORG_NAME/registry-providers")

    if echo "$response" | jq -e '.data.id' >/dev/null 2>&1; then
        provider_id=$(echo "$response" | jq -r '.data.id')
        log_success "Provider created successfully with ID: $provider_id"
        echo "$provider_id" > .provider_id
    else
        log_error "Failed to create provider"
        echo "Response: $response"
        exit 1
    fi
}

# Add GPG key
add_gpg_key() {
    if [ ! -f "gpg_public.key" ]; then
        log_error "gpg_public.key not found. Please run the GPG setup first."
        exit 1
    fi

    log_info "Adding GPG key to HCP Terraform..."

    # Read the public key
    public_key=$(cat gpg_public.key)

    # Create key JSON
    cat > key.json << EOF
{
  "data": {
    "type": "gpg-keys",
    "attributes": {
      "namespace": "$ORG_NAME",
      "ascii-armor": "$(echo "$public_key" | sed 's/$/\\n/g' | tr -d '\n')"
    }
  }
}
EOF

    response=$(curl -s \
        --header "Authorization: Bearer $HCP_TOKEN" \
        --header "Content-Type: application/vnd.api+json" \
        --request POST \
        --data @key.json \
        "https://app.terraform.io/api/registry/public/v2/gpg-keys")

    if echo "$response" | jq -e '.data.attributes.key-id' >/dev/null 2>&1; then
        key_id=$(echo "$response" | jq -r '.data.attributes.key-id')
        log_success "GPG key added successfully with ID: $key_id"
        echo "$key_id" > .key_id
    else
        log_error "Failed to add GPG key"
        echo "Response: $response"
        exit 1
    fi
}

# Create version
create_version() {
    key_id=$(cat .key_id)
    log_info "Creating provider version $VERSION..."

    cat > version.json << EOF
{
  "data": {
    "type": "registry-provider-versions",
    "attributes": {
      "version": "$VERSION",
      "key-id": "$key_id",
      "protocols": ["5.0"]
    }
  }
}
EOF

    response=$(curl -s \
        --header "Authorization: Bearer $HCP_TOKEN" \
        --header "Content-Type: application/vnd.api+json" \
        --request POST \
        --data @version.json \
        "https://app.terraform.io/api/v2/organizations/$ORG_NAME/registry-providers/$ORG_NAME/$PROVIDER_NAME/versions")

    if echo "$response" | jq -e '.data.id' >/dev/null 2>&1; then
        version_id=$(echo "$response" | jq -r '.data.id')
        shasums_upload_url=$(echo "$response" | jq -r '.data.links.shasums-upload')
        shasums_sig_upload_url=$(echo "$response" | jq -r '.data.links.shasums-sig-upload')
        log_success "Version created successfully"
        echo "$version_id" > .version_id
        echo "$shasums_upload_url" > .shasums_upload_url
        echo "$shasums_sig_upload_url" > .shasums_sig_upload_url
    else
        log_error "Failed to create version"
        echo "Response: $response"
        exit 1
    fi
}

# Upload signatures
upload_signatures() {
    log_info "Uploading SHA256SUMS and signature files..."

    shasums_url=$(cat .shasums_upload_url)
    shasums_sig_url=$(cat .shasums_sig_upload_url)

    # Upload SHA256SUMS
    curl -s -T "dist/terraform-provider-nosana_${VERSION}_SHA256SUMS" "$shasums_url"
    log_success "Uploaded SHA256SUMS file"

    # Upload SHA256SUMS.sig
    curl -s -T "dist/terraform-provider-nosana_${VERSION}_SHA256SUMS.sig" "$shasums_sig_url"
    log_success "Uploaded SHA256SUMS.sig file"
}

# Create platform
create_platform() {
    platform=$1
    os=$(echo $platform | cut -d'_' -f1)
    arch=$(echo $platform | cut -d'_' -f2)

    log_info "Creating platform for $platform..."

    # Calculate SHA256 of the binary
    if [ "$os" = "windows" ]; then
        binary_file="dist/terraform-provider-nosana_${VERSION}_${platform}.exe"
        filename="terraform-provider-nosana_${VERSION}_${platform}.exe"
    else
        binary_file="dist/terraform-provider-nosana_${VERSION}_${platform}"
        filename="terraform-provider-nosana_${VERSION}_${platform}"
    fi

    if [ ! -f "$binary_file" ]; then
        log_warning "Binary not found: $binary_file - skipping this platform"
        return
    fi

    shasum=$(shasum -a 256 "$binary_file" | cut -d' ' -f1)

    cat > platform.json << EOF
{
  "data": {
    "type": "registry-provider-version-platforms",
    "attributes": {
      "os": "$os",
      "arch": "$arch",
      "shasum": "$shasum",
      "filename": "$filename"
    }
  }
}
EOF

    response=$(curl -s \
        --header "Authorization: Bearer $HCP_TOKEN" \
        --header "Content-Type: application/vnd.api+json" \
        --request POST \
        --data @platform.json \
        "https://app.terraform.io/api/v2/organizations/$ORG_NAME/registry-providers/$ORG_NAME/$PROVIDER_NAME/versions/$VERSION/platforms")

    if echo "$response" | jq -e '.data.links.provider-binary-upload' >/dev/null 2>&1; then
        binary_upload_url=$(echo "$response" | jq -r '.data.links.provider-binary-upload')
        log_success "Platform created for $platform"

        # Upload binary
        curl -s -T "$binary_file" "$binary_upload_url"
        log_success "Uploaded binary for $platform"
    else
        log_error "Failed to create platform for $platform"
        echo "Response: $response"
    fi
}

# Main function
main() {
    echo "🚀 Publishing Nosana Provider to HCP Terraform Private Registry"
    echo "================================================================="

    check_prerequisites

    # Check if we need to create provider (only once)
    if [ ! -f ".provider_id" ]; then
        create_provider
    else
        log_info "Provider already exists, skipping creation"
    fi

    # Check if we need to add GPG key (only once)
    if [ ! -f ".key_id" ]; then
        add_gpg_key
    else
        log_info "GPG key already exists, skipping addition"
    fi

    # Create version (for each new version)
    if [ ! -f ".version_id" ]; then
        create_version
        upload_signatures
    else
        log_info "Version already exists, skipping creation"
    fi

    # Create platforms (for each platform)
    platforms=("linux_amd64" "linux_arm64" "darwin_amd64" "darwin_arm64" "windows_amd64")
    for platform in "${platforms[@]}"; do
        create_platform "$platform"
    done

    log_success "Provider publishing complete!"
    echo ""
    echo "📦 Users can now install with:"
    echo "   source = \"app.terraform.io/$ORG_NAME/$PROVIDER_NAME\""
    echo ""
    echo "🔗 Check your provider at:"
    echo "   https://app.terraform.io/app/$ORG_NAME/registry/private"
}

# Run main function
main "$@"