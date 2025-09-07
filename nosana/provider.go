// provider.go
package nosana

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/gagliardetto/solana-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mr-tron/base58"
)

// Provider returns a *schema.Provider.
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"private_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("NOSANA_PRIVATE_KEY", ""),
				Description: "Solana private key in base58 format. Can be set via NOSANA_PRIVATE_KEY environment variable.",
			},
			"keypair_path": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("NOSANA_KEYPAIR_PATH", ""),
				Description: "Path to Solana keypair file. If not provided, will use local wallet.",
			},
			"use_local_wallet": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether to use a local wallet (generated automatically if needed).",
			},
			"market_address": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Default market address for job submissions.",
			},
			"rpc_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Solana RPC URL for blockchain transactions. Use a fast RPC for better reliability.",
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

// nosanaClient represents a client for interacting with the Nosana API.
type nosanaClient struct {
	APIClient     *NosanaAPIClient
	MarketAddress string
	RpcURL        string
}

// SolanaKeypair represents a Solana keypair in JSON format
type SolanaKeypair struct {
	PublicKey  []int `json:"publicKey"`
	PrivateKey []int `json:"privateKey"`
}

// getLocalWalletPath returns the path for the local wallet file
func getLocalWalletPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home directory is not accessible
		return "./terraform-provider-nosana-wallet.json"
	}
	
	configDir := filepath.Join(homeDir, ".config", "terraform-provider-nosana")
	os.MkdirAll(configDir, 0700) // Create directory with restricted permissions
	
	return filepath.Join(configDir, "wallet.json")
}

// generateLocalWallet creates a new Solana keypair and saves it to disk
func generateLocalWallet() (string, error) {
	// Generate a new Ed25519 keypair
	_, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate keypair: %w", err)
	}

	// Convert to Solana format
	solanaPrivateKey := solana.PrivateKey(privateKey)
	solanaPublicKey := solanaPrivateKey.PublicKey()

	// Create keypair in the format expected by Solana tools
	keypair := SolanaKeypair{
		PublicKey:  make([]int, len(solanaPublicKey)),
		PrivateKey: make([]int, len(solanaPrivateKey)),
	}

	// Convert bytes to int array (Solana JSON format)
	for i, b := range solanaPublicKey {
		keypair.PublicKey[i] = int(b)
	}
	for i, b := range solanaPrivateKey {
		keypair.PrivateKey[i] = int(b)
	}

	// Save to file
	walletPath := getLocalWalletPath()
	keypairData, err := json.MarshalIndent(keypair, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal keypair: %w", err)
	}

	err = ioutil.WriteFile(walletPath, keypairData, 0600) // Restricted permissions
	if err != nil {
		return "", fmt.Errorf("failed to write wallet file: %w", err)
	}

	log.Printf("[INFO] Generated new local wallet at: %s", walletPath)
	log.Printf("[INFO] Public key: %s", solanaPublicKey.String())
	log.Printf("[INFO] IMPORTANT: Fund this wallet with SOL and NOS tokens!")

	// Return the private key in base58 format for API client
	return base58.Encode(solanaPrivateKey), nil
}

// loadLocalWallet loads an existing wallet from disk
func loadLocalWallet() (string, error) {
	walletPath := getLocalWalletPath()
	
	// Check if wallet file exists
	if _, err := os.Stat(walletPath); os.IsNotExist(err) {
		return "", fmt.Errorf("wallet file not found at %s", walletPath)
	}

	// Read and parse the wallet file
	keypairData, err := ioutil.ReadFile(walletPath)
	if err != nil {
		return "", fmt.Errorf("failed to read wallet file: %w", err)
	}

	var keypair SolanaKeypair
	err = json.Unmarshal(keypairData, &keypair)
	if err != nil {
		return "", fmt.Errorf("failed to parse wallet file: %w", err)
	}

	// Convert back to bytes
	privateKeyBytes := make([]byte, len(keypair.PrivateKey))
	for i, b := range keypair.PrivateKey {
		privateKeyBytes[i] = byte(b)
	}

	// Validate the private key
	if len(privateKeyBytes) != ed25519.PrivateKeySize {
		return "", fmt.Errorf("invalid private key length in wallet file")
	}

	// Return the private key in base58 format
	return base58.Encode(privateKeyBytes), nil
}

// getPrivateKey determines the private key to use based on configuration
func getPrivateKey(d *schema.ResourceData) (string, error) {
	// Priority: 1. Explicit private_key, 2. keypair_path, 3. local wallet
	
	// Check for explicit private key
	if privateKey := d.Get("private_key").(string); privateKey != "" {
		log.Printf("[INFO] Using explicit private key")
		return privateKey, nil
	}

	// Check for keypair path
	if keypairPath := d.Get("keypair_path").(string); keypairPath != "" {
		log.Printf("[INFO] Loading keypair from: %s", keypairPath)
		return loadKeypairFromFile(keypairPath)
	}

	// Use local wallet
	useLocalWallet := d.Get("use_local_wallet").(bool)
	if !useLocalWallet {
		return "", fmt.Errorf("no authentication method configured. Set private_key, keypair_path, or enable use_local_wallet")
	}

	log.Printf("[INFO] Using local wallet")
	
	// Try to load existing wallet first
	privateKey, err := loadLocalWallet()
	if err != nil {
		log.Printf("[INFO] No existing wallet found, generating new one...")
		// Generate new wallet if none exists
		return generateLocalWallet()
	}

	log.Printf("[INFO] Loaded existing local wallet")
	return privateKey, nil
}

// loadKeypairFromFile loads a Solana keypair from a JSON file
func loadKeypairFromFile(filePath string) (string, error) {
	keypairData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read keypair file: %w", err)
	}

	var keypair SolanaKeypair
	err = json.Unmarshal(keypairData, &keypair)
	if err != nil {
		return "", fmt.Errorf("failed to parse keypair file: %w", err)
	}

	// Convert back to bytes
	privateKeyBytes := make([]byte, len(keypair.PrivateKey))
	for i, b := range keypair.PrivateKey {
		privateKeyBytes[i] = byte(b)
	}

	// Validate the private key
	if len(privateKeyBytes) != ed25519.PrivateKeySize {
		return "", fmt.Errorf("invalid private key length in keypair file")
	}

	// Return the private key in base58 format
	return base58.Encode(privateKeyBytes), nil
}

// newNosanaClient creates a new Nosana API client.
func newNosanaClient(privateKey, marketAddress, rpcURL string) (*nosanaClient, error) {
	log.Printf("[INFO] Initializing Nosana client for market: %s", marketAddress)
	if rpcURL != "" {
		log.Printf("[INFO] Using RPC URL: %s", rpcURL)
	} else {
		log.Printf("[INFO] No RPC URL configured, using Nosana backend for blockchain operations")
	}

	apiClient, err := NewNosanaAPIClient(privateKey, rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create Nosana API client: %w", err)
	}

	client := &nosanaClient{
		APIClient:     apiClient,
		MarketAddress: marketAddress,
		RpcURL:        rpcURL,
	}
	log.Printf("[INFO] Nosana API client initialized successfully")
	return client, nil
}

// providerConfigure is called once at the start of a Terraform run.
func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	marketAddress := d.Get("market_address").(string)
	rpcURL := d.Get("rpc_url").(string)

	// Get private key using the new local wallet system
	privateKey, err := getPrivateKey(d)
	if err != nil {
		return nil, diag.FromErr(fmt.Errorf("failed to get private key: %w", err))
	}

	// Create a new Nosana client with the provided configuration.
	client, err := newNosanaClient(privateKey, marketAddress, rpcURL)
	if err != nil {
		return nil, diag.FromErr(fmt.Errorf("failed to configure Nosana provider: %w", err))
	}

	return client, nil
}
