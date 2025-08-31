#!/bin/bash

# Bundle Node.js scripts into standalone executables
# This eliminates the need for users to install Node.js

set -e

echo "📦 Bundling Nosana SDK scripts into standalone executables..."

# Check if pkg is installed
if ! command -v pkg &> /dev/null; then
    echo "Installing pkg globally..."
    npm install -g pkg
fi

# Create output directory
mkdir -p ../dist/scripts

# Bundle each script for multiple platforms
echo "🔨 Building nosana-job-post..."
pkg nosana-job-post.js \
    --target node18-linux-x64,node18-macos-x64,node18-win-x64 \
    --out-path ../dist/scripts

echo "🔨 Building nosana-job-get..."
pkg nosana-job-get.js \
    --target node18-linux-x64,node18-macos-x64,node18-win-x64 \
    --out-path ../dist/scripts

echo "🔨 Building nosana-validate..."
pkg nosana-validate.js \
    --target node18-linux-x64,node18-macos-x64,node18-win-x64 \
    --out-path ../dist/scripts

echo "✅ Bundled executables created in dist/scripts/"
echo "📋 Files created:"
ls -la ../dist/scripts/

echo ""
echo "🎉 Users can now use the provider without installing Node.js!"