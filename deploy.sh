#!/usr/bin/env bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 Deploying Nosana Job via Terraform${NC}"
echo "=================================="

# Change to example directory
cd "$(dirname "$0")/example"

# Check if wallet exists and display public key
WALLET_PATH="$HOME/.config/terraform-provider-nosana/wallet.json"
if [ -f "$WALLET_PATH" ]; then
    echo -e "${GREEN}✅ Local wallet found${NC}"
    # Extract public key from wallet file
    PUBLIC_KEY=$(jq -r '.publicKey | map(tostring) | join("")' "$WALLET_PATH" 2>/dev/null || echo "Could not read public key")
    if [ "$PUBLIC_KEY" != "Could not read public key" ]; then
        # Convert the array of numbers back to base58
        echo -e "${YELLOW}📱 Wallet public key available${NC}"
        echo "Please ensure this wallet is funded with SOL and NOS tokens"
    fi
else
    echo -e "${YELLOW}⚠️  No local wallet found - will be created automatically${NC}"
fi

echo ""
echo -e "${YELLOW}🎯 Applying Terraform configuration...${NC}"

# Apply with auto-approve after confirmation
echo -e "${BLUE}This will create a real deployment on Nosana network.${NC}"
echo -e "${BLUE}Make sure your wallet has sufficient funds!${NC}"
echo ""
read -p "Continue with deployment? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    terraform apply -auto-approve
    
    if [ $? -eq 0 ]; then
        echo ""
        echo -e "${GREEN}🎉 Deployment completed successfully!${NC}"
        echo ""
        echo -e "${BLUE}Deployment Details:${NC}"
        terraform output
        echo ""
        echo -e "${YELLOW}Next steps:${NC}"
        echo "1. Monitor the deployment status in Nosana dashboard"
        echo "2. Check logs and metrics for your deployment"
        echo "3. When done, run: ./destroy.sh to clean up resources"
    else
        echo -e "${RED}❌ Deployment failed!${NC}"
        echo "Check the error messages above for details"
        exit 1
    fi
else
    echo -e "${YELLOW}Deployment cancelled${NC}"
    exit 0
fi
