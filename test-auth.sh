#!/usr/bin/env bash

set -e

echo "🔧 Testing Nosana API with your private key..."

cd "$(dirname "$0")"

# Set your private key
export NOSANA_PRIVATE_KEY="5YeqfFZJfJf8JRPdUqCzNjUfJuMYc7KyxkTr63T8TgcBVwPkfKYWB7yG566v9jaMoFPvDrBLnZQenAfjRVtur5ob"
export NOSANA_MARKET_ADDRESS="HanragNudL4S4zFtpLQv85dn6QbdzCm7SNEWEb9sRp17"

# Create a simple test program to check API connectivity
cat > test_auth.go << 'EOF'
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/HoomanDigital/terraform-provider-nosana/nosana"
)

func main() {
	privateKey := os.Getenv("NOSANA_PRIVATE_KEY")
	if privateKey == "" {
		log.Fatal("NOSANA_PRIVATE_KEY environment variable not set")
	}

	client, err := nosana.NewNosanaAPIClient(privateKey)
	if err != nil {
		log.Fatalf("❌ Failed to create API client: %v", err)
	}

	fmt.Printf("✅ API client created successfully\n")
	fmt.Printf("🔑 Public key: %s\n", client.PublicKey.String())

	// Test creating a deployment with the exact same data as Terraform
	ctx := context.Background()
	marketAddress := os.Getenv("NOSANA_MARKET_ADDRESS")
	
	createBody := &nosana.DeploymentCreateBody{
		Name:               "terraform-test-deployment",
		Market:             marketAddress,
		IpfsDefinitionHash: stringPtr("QmYwAPJzv5CZsnA625s3Xf2nemtYgPpHdWEz79ojWnPbdG"), // Known test IPFS hash
		Replicas:           1,
		Timeout:            300,
		Strategy:           nosana.DeploymentStrategySimple,
		Schedule:           nil,
	}

	fmt.Printf("🚀 Attempting to create deployment...\n")
	fmt.Printf("📋 Request: %+v\n", createBody)
	
	deployment, err := client.CreateDeployment(ctx, createBody)
	if err != nil {
		fmt.Printf("❌ Deployment creation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Deployment created successfully!\n")
	fmt.Printf("🆔 Deployment ID: %s\n", deployment.ID)
	fmt.Printf("📊 Status: %s\n", deployment.Status)
}

func stringPtr(s string) *string {
	return &s
}
EOF

echo "🔨 Building auth test..."
go mod tidy
go build -o test_auth test_auth.go

echo "🧪 Running authentication test..."
./test_auth

# Clean up
rm -f test_auth test_auth.go

echo "✅ Authentication test completed"
