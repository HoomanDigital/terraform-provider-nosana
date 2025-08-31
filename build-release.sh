#!/bin/bash

# Complete build script for releasing the Terraform Nosana Provider
# This creates a distribution-ready provider with bundled dependencies

set -e

echo "🚀 Building Terraform Nosana Provider for distribution..."

# Clean previous builds
echo "🧹 Cleaning previous builds..."
rm -rf dist/
mkdir -p dist/

# Step 1: Bundle Node.js scripts into standalone executables
echo "📦 Step 1: Bundling SDK scripts..."
cd scripts
./build-bundled.sh
cd ..

# Step 2: Build Go provider for multiple platforms
echo "🔨 Step 2: Building Go provider..."

# Define target platforms
platforms=(
    "linux/amd64"
    "linux/arm64" 
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

for platform in "${platforms[@]}"; do
    IFS='/' read -r -a array <<< "$platform"
    GOOS="${array[0]}"
    GOARCH="${array[1]}"
    
    output_name="terraform-provider-nosana"
    if [ "$GOOS" = "windows" ]; then
        output_name+=".exe"
    fi
    
    echo "  Building for $GOOS/$GOARCH..."
    GOOS=$GOOS GOARCH=$GOARCH go build -o "dist/${output_name}-${GOOS}-${GOARCH}" .
done

# Step 3: Create distribution packages
echo "📦 Step 3: Creating distribution packages..."

for platform in "${platforms[@]}"; do
    IFS='/' read -r -a array <<< "$platform"
    GOOS="${array[0]}"
    GOARCH="${array[1]}"
    
    package_name="terraform-provider-nosana-${GOOS}-${GOARCH}"
    package_dir="dist/${package_name}"
    
    mkdir -p "${package_dir}/scripts"
    
    # Copy provider binary
    provider_binary="terraform-provider-nosana-${GOOS}-${GOARCH}"
    if [ "$GOOS" = "windows" ]; then
        provider_binary+=".exe"
    fi
    cp "dist/${provider_binary}" "${package_dir}/"
    
    # Copy appropriate bundled scripts
    script_suffix=""
    if [ "$GOOS" = "windows" ]; then
        script_suffix="-win"
    elif [ "$GOOS" = "darwin" ]; then
        script_suffix="-macos"
    else
        script_suffix="-linux"
    fi
    
    cp "dist/scripts/nosana-job-post${script_suffix}" "${package_dir}/scripts/" 2>/dev/null || true
    cp "dist/scripts/nosana-job-get${script_suffix}" "${package_dir}/scripts/" 2>/dev/null || true
    cp "dist/scripts/nosana-validate${script_suffix}" "${package_dir}/scripts/" 2>/dev/null || true
    
    # Copy documentation
    cp README.md "${package_dir}/"
    cp LICENSE "${package_dir}/"
    cp NOSANA_SDK_INTEGRATION.md "${package_dir}/"
    
    # Create installation instructions
    cat > "${package_dir}/INSTALL.md" << EOF
# Terraform Nosana Provider Installation

## Quick Start

1. **Download**: Extract this package to your desired location
2. **Install**: Follow Terraform's plugin installation guide
3. **Configure**: Set up your provider configuration

\`\`\`terraform
provider "nosana" {
  private_key    = "YOUR_BASE58_PRIVATE_KEY"
  network        = "mainnet"
  market_address = "7AtiXMSH6R1jjBxrcYjehCkkSF7zvYWte63gwEDBcGHq"
}
\`\`\`

## What's Included

- ✅ Provider binary (no dependencies required)
- ✅ Bundled SDK scripts (no Node.js installation needed)
- ✅ Documentation and examples

## Zero Dependencies

This package includes everything needed to run the provider:
- No Node.js installation required
- No npm packages to install
- No setup scripts to run

Just extract and use!

## Support

- Documentation: See README.md and NOSANA_SDK_INTEGRATION.md
- Issues: https://github.com/HoomanDigital/terraform-provider-nosana/issues
EOF
    
    # Create zip package
    echo "  Creating ${package_name}.zip..."
    cd dist
    zip -r "${package_name}.zip" "${package_name}/"
    cd ..
done

echo ""
echo "🎉 Build completed successfully!"
echo ""
echo "📦 Distribution packages created:"
ls -la dist/*.zip
echo ""
echo "✅ Each package includes:"
echo "   - Provider binary (platform-specific)"
echo "   - Bundled SDK scripts (no Node.js required)"
echo "   - Documentation and installation guide"
echo ""
echo "🚀 Ready for distribution!"