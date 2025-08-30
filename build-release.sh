#!/bin/bash

# Build and package Terraform provider for public registry release
# This creates properly named and signed binaries for the public Terraform Registry

set -e

# Configuration
PROVIDER_NAME="terraform-provider-nosana"
VERSION="0.1.0"
PLATFORMS=("linux_amd64" "linux_arm64" "darwin_amd64" "darwin_arm64" "windows_amd64")

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

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

# Build binary for specific platform
build_binary() {
    local platform=$1
    local os=$(echo $platform | cut -d'_' -f1)
    local arch=$(echo $platform | cut -d'_' -f2)

    log_info "Building for $platform..."

    # Set output name
    if [ "$os" = "windows" ]; then
        local binary_name="${PROVIDER_NAME}.exe"
        local output_name="${PROVIDER_NAME}_${VERSION}_${platform}.exe"
    else
        local binary_name="${PROVIDER_NAME}"
        local output_name="${PROVIDER_NAME}_${VERSION}_${platform}"
    fi

    # Build the binary
    GOOS=$os GOARCH=$arch go build -o "dist/${output_name}" .

    if [ $? -eq 0 ]; then
        log_success "Built ${output_name}"
        echo "dist/${output_name}"
    else
        log_error "Failed to build for $platform"
        exit 1
    fi
}

# Create zip archive for binary
create_zip() {
    local binary_path=$1
    local platform=$2

    local zip_name="${PROVIDER_NAME}_${VERSION}_${platform}.zip"
    local zip_path="dist/${zip_name}"

    log_info "Creating zip: ${zip_name}"

    # Create zip containing just the binary
    cd dist
    if [ "$platform" = *"windows"* ]; then
        zip "${zip_name}" "$(basename "$binary_path")"
    else
        zip "${zip_name}" "$(basename "$binary_path")"
    fi
    cd ..

    if [ -f "dist/${zip_name}" ]; then
        log_success "Created ${zip_name}"
        echo "dist/${zip_name}"
    else
        log_error "Failed to create zip for $platform"
        exit 1
    fi
}

# Sign the zip file
sign_zip() {
    local zip_path=$1
    local sig_path="${zip_path}.asc"

    log_info "Signing ${zip_path}"

    # Sign the zip file
    gpg --detach-sign --armor --output "${sig_path}" "${zip_path}"

    if [ -f "${sig_path}" ]; then
        log_success "Signed ${zip_path}"
        echo "${sig_path}"
    else
        log_error "Failed to sign ${zip_path}"
        exit 1
    fi
}

# Create SHA256SUMS file
create_checksums() {
    log_info "Creating SHA256SUMS file..."

    cd dist
    shasum -a 256 *.zip > "${PROVIDER_NAME}_${VERSION}_SHA256SUMS"
    cd ..

    log_success "Created SHA256SUMS file"
}

# Main build process
main() {
    echo "🏗️  Building ${PROVIDER_NAME} ${VERSION} for Public Registry"
    echo "=========================================================="

    # Check if GPG key is available
    if ! gpg --list-secret-keys | grep -q "A301B99BE1C61FC0"; then
        log_error "GPG key not found. Please ensure your GPG key is available."
        exit 1
    fi

    # Create dist directory
    mkdir -p dist

    # Build, zip, and sign for each platform
    for platform in "${PLATFORMS[@]}"; do
        log_info "Processing platform: $platform"

        # Build binary
        binary_path=$(build_binary "$platform")

        # Create zip
        zip_path=$(create_zip "$binary_path" "$platform")

        # Sign zip
        sig_path=$(sign_zip "$zip_path")

        echo ""
    done

    # Create checksums
    create_checksums

    # List all created files
    echo ""
    log_success "Build complete! Created files:"
    ls -la dist/

    echo ""
    log_info "Next steps:"
    echo "1. Create a GitHub release with tag 'v${VERSION}'"
    echo "2. Upload all files from dist/ to the release"
    echo "3. The public registry will automatically discover your provider!"
    echo ""
    log_info "Release files ready in: $(pwd)/dist/"
}

# Run main function
main "$@"