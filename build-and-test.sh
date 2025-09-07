#!/bin/bash

# 🚀 Nosana Terraform Provider Build & Test Script
# This script automates the entire build, install, and test cycle

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROVIDER_NAME="terraform-provider-nosana_v1.0.0"
PROVIDER_NAMESPACE="localhost/hoomandigital/nosana"
PROVIDER_VERSION="1.0.0"
TERRAFORM_DIR="demo"

echo -e "${BLUE}🚀 Nosana Terraform Provider Build & Test Script${NC}"
echo -e "${BLUE}=================================================${NC}"
echo ""

# Step 1: Build the provider
echo -e "${YELLOW}🔨 Step 1: Building Terraform provider...${NC}"
cd /home/dhruv/Documents/Code/new-tf-provider
go build -o $PROVIDER_NAME

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Provider built successfully${NC}"
else
    echo -e "${RED}❌ Provider build failed${NC}"
    exit 1
fi

# Step 2: Install provider to local directory
echo -e "${YELLOW}📦 Step 2: Installing provider to local directory...${NC}"
PLUGIN_DIR="$HOME/.terraform.d/plugins/localhost/hoomandigital/nosana/$PROVIDER_VERSION/linux_amd64"
mkdir -p "$PLUGIN_DIR"
cp "$PROVIDER_NAME" "$PLUGIN_DIR/"
chmod +x "$PLUGIN_DIR/$PROVIDER_NAME"
echo -e "${GREEN}✅ Provider installed to $PLUGIN_DIR${NC}"

# Step 3: Setup Terraform configuration
echo -e "${YELLOW}⚙️  Step 3: Setting up Terraform configuration...${NC}"
cd "$TERRAFORM_DIR"

# Create/update .terraformrc for dev overrides
cat > .terraformrc << EOF
provider_installation {
  dev_overrides {
    "$PROVIDER_NAMESPACE" = "/home/dhruv/Documents/Code/new-tf-provider"
  }
  direct {
    enabled = true
  }
}
EOF

echo -e "${GREEN}✅ Created .terraformrc with dev overrides${NC}"

# Step 4: Clean up old state and lock files
echo -e "${YELLOW}🧹 Step 4: Cleaning up old Terraform state...${NC}"
rm -f terraform.tfstate* .terraform.lock.hcl
rm -rf .terraform/
# Also remove any residual lock files that might conflict
find . -name ".terraform.lock.hcl" -delete 2>/dev/null || true
echo -e "${GREEN}✅ Cleaned up old state files${NC}"

# Step 5: Handle conflicting .terraformrc files
echo -e "${YELLOW}🔧 Step 5: Managing Terraform configuration conflicts...${NC}"
if [ -f "$HOME/.terraformrc" ]; then
    echo -e "${BLUE}   📋 Backing up existing ~/.terraformrc${NC}"
    mv "$HOME/.terraformrc" "$HOME/.terraformrc.backup"
fi
echo -e "${GREEN}✅ Cleared conflicting Terraform configurations${NC}"

# Step 6: Set environment variables for debug mode
echo -e "${YELLOW}🔍 Step 6: Setting up debug environment...${NC}"
export TERRAFORM_CLI_CONFIG_FILE="$(pwd)/.terraformrc"
export TF_LOG=DEBUG
export TF_LOG_PROVIDER=DEBUG

echo -e "${GREEN}✅ Debug environment configured${NC}"
echo -e "${BLUE}   TERRAFORM_CLI_CONFIG_FILE: $TERRAFORM_CLI_CONFIG_FILE${NC}"
echo -e "${BLUE}   TF_LOG: $TF_LOG${NC}"
echo -e "${BLUE}   TF_LOG_PROVIDER: $TF_LOG_PROVIDER${NC}"

# Step 7: Initialize Terraform (skip if plan only and init not needed)
echo -e "${YELLOW}🔄 Step 7: Initializing Terraform (if needed)...${NC}"
# Try a quick init to handle any dependency issues
if terraform init -input=false &>/dev/null; then
    echo -e "${GREEN}✅ Terraform initialized successfully${NC}"
else
    echo -e "${BLUE}   📋 Terraform init not needed or skipped${NC}"
fi

# Step 8: Test the provider
echo ""
echo -e "${YELLOW}🧪 Step 8: Testing the provider...${NC}"
echo -e "${BLUE}=================================================${NC}"

# Check if user wants to run plan or apply
ACTION=${1:-plan}

case $ACTION in
    "plan")
        echo -e "${BLUE}📋 Running: terraform plan${NC}"
        terraform plan
        ;;
    "apply")
        echo -e "${BLUE}🚀 Running: terraform apply (auto-approve)${NC}"
        terraform apply -auto-approve
        ;;
    "apply-interactive")
        echo -e "${BLUE}🚀 Running: terraform apply (interactive)${NC}"
        terraform apply
        ;;
    "destroy")
        echo -e "${BLUE}💥 Running: terraform destroy${NC}"
        terraform destroy -auto-approve
        ;;
    *)
        echo -e "${RED}❌ Unknown action: $ACTION${NC}"
        echo -e "${YELLOW}Usage: $0 [plan|apply|apply-interactive|destroy]${NC}"
        echo -e "${YELLOW}Default: plan${NC}"
        exit 1
        ;;
esac

if [ $? -eq 0 ]; then
    echo ""
    echo -e "${GREEN}🎉 SUCCESS! Provider test completed successfully${NC}"
    echo -e "${BLUE}🌐 Check your deployment at: https://dashboard.nosana.com/account/deployer${NC}"
else
    echo ""
    echo -e "${RED}❌ FAILED! Provider test encountered an error${NC}"
    exit 1
fi

echo ""
echo -e "${BLUE}=================================================${NC}"

# Step 9: Restore original configuration
echo -e "${YELLOW}🔄 Step 9: Restoring original configuration...${NC}"
if [ -f "$HOME/.terraformrc.backup" ]; then
    mv "$HOME/.terraformrc.backup" "$HOME/.terraformrc"
    echo -e "${GREEN}✅ Restored original ~/.terraformrc${NC}"
else
    echo -e "${BLUE}   📋 No original ~/.terraformrc to restore${NC}"
fi

echo -e "${BLUE}🏁 Build & Test Complete!${NC}"
echo ""
echo -e "${YELLOW}💡 Tips:${NC}"
echo -e "${YELLOW}   • Run with 'apply' to deploy: ./build-and-test.sh apply${NC}"
echo -e "${YELLOW}   • Run with 'destroy' to clean up: ./build-and-test.sh destroy${NC}"
echo -e "${YELLOW}   • Debug logs are enabled for detailed output${NC}"
echo -e "${YELLOW}   • Check terraform.tfstate for deployment details${NC}"