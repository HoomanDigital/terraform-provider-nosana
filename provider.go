// provider.go
package main

import (
	"context"
	// "encoding/json"
	"fmt"
	"log"
	// "time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider returns a *schema.Provider.
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"wallet_address": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The Solana wallet address for Nosana authentication.",
			},
			"signed_challenge": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "A signed message from the Phantom wallet for authentication.",
				Sensitive:   true, // Mark as sensitive to prevent logging
			},
			"network": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "mainnet",
				Description: "The Nosana network to connect to (e.g., 'mainnet', 'devnet').",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"nosana_job": resourceNosanaJob(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			// No data sources defined for MVP
		},
		ConfigureContextFunc: providerConfigure,
	}
}

// nosanaClient represents a simplified client for interacting with the Nosana API.
// In a real scenario, this would handle API requests, authentication, and error handling.
type nosanaClient struct {
	WalletAddress string
	SignedChallenge string
	Network string
	// Add fields for API base URL, HTTP client, authentication token, etc.
	authToken string
}

// newNosanaClient creates a new Nosana API client.
// This function would perform the actual authentication with the Nosana API.
func newNosanaClient(walletAddress, signedChallenge, network string) (*nosanaClient, error) {
	log.Printf("[INFO] Initializing Nosana client for wallet: %s, network: %s", walletAddress, network)

	client := &nosanaClient{
		WalletAddress: walletAddress,
		SignedChallenge: signedChallenge,
		Network: network,
	}

	// --- Placeholder for actual Nosana API authentication ---
	// In a real implementation, you would make an HTTP request to Nosana's auth endpoint
	// using walletAddress and signedChallenge, and receive an auth token.
	// Example:
	// resp, err := http.Post("https://api.nosana.com/auth", "application/json", body)
	// if err != nil { return nil, fmt.Errorf("auth request failed: %w", err) }
	// defer resp.Body.Close()
	// if resp.StatusCode != http.StatusOK { return nil, fmt.Errorf("auth failed with status: %d", resp.StatusCode) }
	// Parse response to get client.authToken
	client.authToken = "mock-nosana-auth-token-123" // Mock token for illustration
	log.Printf("[INFO] Nosana client authenticated. Mock token: %s", client.authToken)
	// --- End placeholder ---

	return client, nil
}

// providerConfigure is called once at the start of a Terraform run.
func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	walletAddress := d.Get("wallet_address").(string)
	signedChallenge := d.Get("signed_challenge").(string)
	network := d.Get("network").(string)

	// Create a new Nosana client with the provided configuration.
	client, err := newNosanaClient(walletAddress, signedChallenge, network)
	if err != nil {
		return nil, diag.FromErr(fmt.Errorf("failed to configure Nosana provider: %w", err))
	}

	return client, nil
}