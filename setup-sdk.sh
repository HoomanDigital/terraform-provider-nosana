#!/bin/bash

# Terraform Nosana Provider SDK Setup Script
# This script helps set up the Node.js SDK bridge for the Terraform provider

set -e

echo "🚀 Setting up Terraform Nosana Provider with SDK integration..."
echo

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    echo "❌ Node.js is not installed."
    echo "📦 Please install Node.js 16+ from https://nodejs.org/"
    echo "   Or use a package manager:"
    echo "   - Ubuntu/Debian: sudo apt install nodejs npm"
    echo "   - macOS: brew install node"
    echo "   - Arch Linux: sudo pacman -S nodejs npm"
    exit 1
fi

# Check Node.js version
NODE_VERSION=$(node -v | cut -d'v' -f2)
MAJOR_VERSION=$(echo $NODE_VERSION | cut -d'.' -f1)

if [ "$MAJOR_VERSION" -lt 16 ]; then
    echo "❌ Node.js version $NODE_VERSION is too old."
    echo "📦 Please upgrade to Node.js 16 or higher."
    exit 1
fi

echo "✅ Node.js version: $(node -v)"

# Check if npm is available
if ! command -v npm &> /dev/null; then
    echo "❌ npm is not installed."
    echo "📦 Please install npm along with Node.js."
    exit 1
fi

echo "✅ npm version: $(npm -v)"

# Change to scripts directory
if [ ! -d "scripts" ]; then
    echo "❌ Scripts directory not found."
    echo "📁 Please run this from the terraform-provider-nosana root directory."
    exit 1
fi

cd scripts

echo "📦 Installing Nosana SDK dependencies..."
if npm install; then
    echo "✅ Dependencies installed successfully"
else
    echo "❌ Failed to install dependencies"
    exit 1
fi

echo

# Run the setup validation script
echo "🔍 Validating SDK installation..."
if node setup.js; then
    echo "✅ SDK validation completed"
else
    echo "❌ SDK validation failed"
    exit 1
fi

echo
echo "🎉 Setup completed successfully!"
echo
echo "📋 Next steps:"
echo "1. Ensure you have a Solana wallet with SOL and NOS tokens"
echo "2. Get your private key in base58 format"
echo "3. Configure your Terraform provider:"
echo
echo "   provider \"nosana\" {"
echo "     private_key    = \"YOUR_BASE58_PRIVATE_KEY\""
echo "     network        = \"mainnet\"  # or \"devnet\""
echo "     market_address = \"7AtiXMSH6R1jjBxrcYjehCkkSF7zvYWte63gwEDBcGHq\""
echo "   }"
echo
echo "4. Run terraform plan/apply"
echo
echo "📖 For more information, see NOSANA_SDK_INTEGRATION.md"