#!/usr/bin/env bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🗑️  Destroying Nosana Deployment${NC}"
echo "==============================="

# Change to example directory
cd "$(dirname "$0")/example"

echo -e "${YELLOW}This will destroy all Terraform-managed resources.${NC}"
echo ""
read -p "Continue with destruction? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    terraform destroy -auto-approve
    
    if [ $? -eq 0 ]; then
        echo ""
        echo -e "${GREEN}🎉 Resources destroyed successfully!${NC}"
    else
        echo -e "${RED}❌ Destruction failed!${NC}"
        echo "Check the error messages above for details"
        echo "You may need to manually clean up resources"
        exit 1
    fi
else
    echo -e "${YELLOW}Destruction cancelled${NC}"
    exit 0
fi
