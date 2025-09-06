#!/usr/bin/env bash

set -e

echo "🔧 Testing Nosana API connectivity..."

cd "$(dirname "$0")"

# Create a simple test program to check API connectivity
cat > test_api.go << 'EOF'
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/HoomanDigital/terraform-provider-nosana/nosana"
)

func main() {
	// Create a client without using the full provider
	privateKey := os.Getenv("NOSANA_PRIVATE_KEY")
	if privateKey == "" {
		fmt.Println("❌ NOSANA_PRIVATE_KEY environment variable not set")
		fmt.Println("Please export your Solana private key in base58 format")
		os.Exit(1)
	}

	client, err := nosana.NewNosanaAPIClient(privateKey)
	if err != nil {
		log.Fatalf("❌ Failed to create API client: %v", err)
	}

	fmt.Println("✅ API client created successfully")
	fmt.Printf("🔑 Public key: %s\n", client.PublicKey.String())

	// Test a simple API call (this might fail due to auth but will show if we can reach the API)
	ctx := context.Background()
	_, err = client.GetDeployment(ctx, "test-id")
	if err != nil {
		fmt.Printf("⚠️  Expected error (testing connectivity): %v\n", err)
		if err.Error() == "API request failed with status 404: Not Found" {
			fmt.Println("✅ API is reachable (404 is expected for test ID)")
		} else {
			fmt.Printf("⚠️  Unexpected API response: %v\n", err)
		}
	} else {
		fmt.Println("✅ API call successful")
	}
}
EOF

echo "🔨 Building API test..."
go mod tidy
go build -o test_api test_api.go

echo "🧪 Running API connectivity test..."
echo "Note: You need to set NOSANA_PRIVATE_KEY environment variable"
echo "Example: export NOSANA_PRIVATE_KEY=your_base58_encoded_private_key"

if [ -n "$NOSANA_PRIVATE_KEY" ]; then
    ./test_api
else
    echo "⚠️  NOSANA_PRIVATE_KEY not set, skipping API test"
fi

# Clean up
rm -f test_api test_api.go

echo "✅ API connectivity test completed"
