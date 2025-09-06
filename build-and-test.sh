#!/usr/bin/env bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 Building and Testing Terraform Provider for Nosana${NC}"
echo "=================================================="

# Change to project directory
cd "$(dirname "$0")"

# Set environment variables for testing
export NOSANA_MARKET_ADDRESS="HanragNudL4S4zFtpLQv85dn6QbdzCm7SNEWEb9sRp17"
export TF_LOG=DEBUG

echo -e "${YELLOW}📦 Installing Go dependencies...${NC}"
go mod tidy

echo -e "${YELLOW}🔨 Building the provider...${NC}"
go build -o terraform-provider-nosana .

# Copy to the development override location
echo -e "${YELLOW}� Installing provider binary to dev override location...${NC}"
mkdir -p /home/dhruv/go/bin
cp terraform-provider-nosana /home/dhruv/go/bin/terraform-provider-nosana

echo -e "${GREEN}✅ Provider built and installed successfully!${NC}"

# Run unit tests
echo -e "${YELLOW}🧪 Running unit tests...${NC}"
go test ./nosana -v

echo -e "${YELLOW}🎯 Testing example configuration...${NC}"

# Change to example directory
cd example

# Clean up any existing state
rm -rf .terraform .terraform.lock.hcl terraform.tfstate terraform.tfstate.backup

# Validate the configuration
echo -e "${YELLOW}✅ Validating configuration...${NC}"
terraform validate

# Plan the deployment (this will test provider authentication and API connectivity)
echo -e "${YELLOW}📋 Planning deployment...${NC}"
terraform plan

echo -e "${GREEN}🎉 All tests completed successfully!${NC}"
echo ""
echo -e "${BLUE}Next steps:${NC}"
echo "1. Review the plan output above"
echo "2. Ensure your local wallet has sufficient SOL and NOS tokens"
echo "3. If everything looks good, run: terraform apply"
echo "4. To clean up resources, run: terraform destroy"
echo ""
echo -e "${YELLOW}Wallet Information:${NC}"
echo "- Local wallet will be created automatically at: ~/.config/terraform-provider-nosana/wallet.json"
echo "- You can check the public key by examining the wallet file"
echo "- Make sure to fund this wallet with SOL and NOS tokens for deployments"
echo ""
echo -e "${YELLOW}Environment Variables:${NC}"
echo "- NOSANA_MARKET_ADDRESS: ${NOSANA_MARKET_ADDRESS}"
echo "- TF_LOG: ${TF_LOG} (for debugging)"
