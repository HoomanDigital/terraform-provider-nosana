#!/bin/bash
# setup-nosana-provider.sh - Quick setup script for Nosana Terraform Provider

echo "🚀 Nosana Terraform Provider Setup"
echo "===================================="

# Check if required tools are installed
echo "📋 Checking prerequisites..."

# Check Node.js/npm
if ! command -v npm &> /dev/null; then
    echo "❌ npm not found. Please install Node.js first:"
    echo "   - Visit: https://nodejs.org/"
    exit 1
fi
echo "✅ npm found"

# Check Terraform
if ! command -v terraform &> /dev/null; then
    echo "❌ terraform not found. Please install Terraform first:"
    echo "   - Visit: https://terraform.io/downloads"
    exit 1
fi
echo "✅ terraform found"

# Install Nosana CLI
echo ""
echo "📦 Installing Nosana CLI..."
npm install -g @nosana/cli

if [ $? -eq 0 ]; then
    echo "✅ Nosana CLI installed successfully"
else
    echo "❌ Failed to install Nosana CLI"
    exit 1
fi

# Check Nosana CLI version
echo ""
echo "🔍 Nosana CLI version:"
nosana --version

echo ""
echo "🎉 Setup complete!"
echo ""
echo "Next steps:"
echo "1. Get your Solana private key:"
echo "   - From Phantom: Settings → Security & Privacy → Export Private Key"
echo "   - From Solflare: Settings → Export Private Key"
echo "   - From CLI: solana-keygen recover 'prompt:' (if you have the seed phrase)"
echo ""
echo "2. Set environment variables:"
echo "   export NOSANA_PRIVATE_KEY=\"your_base58_private_key_here\""
echo "   export TF_VAR_market_address=\"7AtiXMSH6R1jjBxrcYjehCkkSF7zvYWte63gwEDBcGHq\""
echo ""
echo "3. Ensure your wallet has SOL and NOS tokens:"
echo "   - SOL: For transaction fees (at least 0.01 SOL)"
echo "   - NOS: For paying job execution costs"
echo "   - You can buy NOS on exchanges like Kraken"
echo ""
echo "4. Initialize and run Terraform:"
echo "   terraform init"
echo "   terraform plan"
echo "   terraform apply"
echo ""
echo "📚 For more information, visit: https://docs.nosana.com/"
