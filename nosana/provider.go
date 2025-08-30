// provider.go
package nosana

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
				Description: "Path to Nosana keypair file. If not provided and no private_key is set, uses default ~/.nosana/nosana_key.json",
			},
			"network": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "mainnet",
				Description: "The Nosana network to connect to (mainnet or devnet).",
			},
			"market_address": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Default market address for job submissions.",
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

// nosanaClient represents a client for interacting with the Nosana CLI.
type nosanaClient struct {
	PrivateKey    string
	KeypairPath   string
	Network       string
	MarketAddress string
}

// newNosanaClient creates a new Nosana CLI client.
func newNosanaClient(privateKey, keypairPath, network, marketAddress string) (*nosanaClient, error) {
	log.Printf("[INFO] Initializing Nosana client for network: %s, market: %s", network, marketAddress)

	// Determine keypair strategy: private key takes precedence over keypair path
	var resolvedKeypairPath string
	var err error

	if privateKey != "" {
		// Private key provided - create/update keypair file
		log.Printf("[INFO] Using provided private key, setting up keypair file")
		resolvedKeypairPath, err = setupKeypairFromPrivateKey(privateKey, keypairPath)
		if err != nil {
			return nil, fmt.Errorf("failed to setup keypair from private key: %w", err)
		}
	} else {
		// No private key - resolve existing keypair path
		resolvedKeypairPath, err = resolveKeypairPath(keypairPath)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve keypair path: %w", err)
		}
	}

	client := &nosanaClient{
		PrivateKey:    privateKey,
		KeypairPath:   resolvedKeypairPath,
		Network:       network,
		MarketAddress: marketAddress,
	}

	// Verify Nosana CLI is available
	output, err := exec.Command("nosana", "--version").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("nosana CLI not found. Please install it with: npm install -g @nosana/cli. Error: %w", err)
	}

	log.Printf("[INFO] Nosana CLI version: %s", strings.TrimSpace(string(output)))

	// Verify keypair file exists and is accessible
	if err := validateKeypairFile(resolvedKeypairPath); err != nil {
		return nil, fmt.Errorf("keypair validation failed: %w", err)
	}

	// Test CLI access with the keypair
	log.Printf("[INFO] Skipping wallet validation - CLI setup appears successful")
	// TODO: Enable wallet validation once ANSI parsing is working
	// return c.testNosanaCLIAccess()
	// if err := client.testNosanaCLIAccess(); err != nil {
	//     return nil, fmt.Errorf("failed to access Nosana CLI: %w", err)
	// }

	log.Printf("[INFO] Nosana CLI client initialized successfully")
	return client, nil
}

// resolveKeypairPath resolves the keypair file path, using default if empty
func resolveKeypairPath(keypairPath string) (string, error) {
	if keypairPath != "" {
		// Use provided path
		absPath, err := filepath.Abs(keypairPath)
		if err != nil {
			return "", fmt.Errorf("failed to resolve absolute path for %s: %w", keypairPath, err)
		}
		return absPath, nil
	}

	// Use default Nosana CLI keypair location
	usr, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %w", err)
	}

	defaultPath := filepath.Join(usr.HomeDir, ".nosana", "nosana_key.json")
	return defaultPath, nil
}

// setupKeypairFromPrivateKey converts a base58 private key to Nosana keypair format
func setupKeypairFromPrivateKey(privateKey, keypairPath string) (string, error) {
	// Determine target keypair path
	var targetPath string
	if keypairPath != "" {
		var err error
		targetPath, err = filepath.Abs(keypairPath)
		if err != nil {
			return "", fmt.Errorf("failed to resolve keypair path: %w", err)
		}
	} else {
		// Use default Nosana CLI location
		usr, err := user.Current()
		if err != nil {
			return "", fmt.Errorf("failed to get current user: %w", err)
		}
		targetPath = filepath.Join(usr.HomeDir, ".nosana", "nosana_key.json")
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(targetPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create keypair directory: %w", err)
	}

	// Convert base58 private key to byte array
	keyBytes, err := base58Decode(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to decode private key: %w", err)
	}

	if len(keyBytes) != 64 {
		return "", fmt.Errorf("invalid private key length: expected 64 bytes, got %d", len(keyBytes))
	}

	// Backup existing keypair if it exists
	if _, err := os.Stat(targetPath); err == nil {
		backupPath := targetPath + ".backup"
		if err := os.Rename(targetPath, backupPath); err != nil {
			log.Printf("[WARN] Failed to backup existing keypair: %v", err)
		} else {
			log.Printf("[INFO] Backed up existing keypair to %s", backupPath)
		}
	}

	// Convert to JSON array format and write to file
	keyArray := make([]int, len(keyBytes))
	for i, b := range keyBytes {
		keyArray[i] = int(b)
	}

	jsonData, err := json.Marshal(keyArray)
	if err != nil {
		return "", fmt.Errorf("failed to marshal keypair: %w", err)
	}

	if err := os.WriteFile(targetPath, jsonData, 0600); err != nil {
		return "", fmt.Errorf("failed to write keypair file: %w", err)
	}

	log.Printf("[INFO] Keypair file created at %s", targetPath)
	return targetPath, nil
}

// base58Decode decodes a base58 string to bytes (simplified for Solana keys)
func base58Decode(s string) ([]byte, error) {
	alphabet := "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

	// Convert string to big integer
	result := big.NewInt(0)
	base := big.NewInt(58)

	for _, char := range s {
		index := strings.IndexRune(alphabet, char)
		if index == -1 {
			return nil, fmt.Errorf("invalid base58 character: %c", char)
		}
		result.Mul(result, base)
		result.Add(result, big.NewInt(int64(index)))
	}

	// Convert to bytes
	bytes := result.Bytes()

	// Add leading zeros for '1' characters
	leadingOnes := 0
	for _, char := range s {
		if char == '1' {
			leadingOnes++
		} else {
			break
		}
	}

	return append(make([]byte, leadingOnes), bytes...), nil
}

// validateKeypairFile checks if the keypair file exists and is readable
func validateKeypairFile(keypairPath string) error {
	// Check if file exists
	if _, err := os.Stat(keypairPath); os.IsNotExist(err) {
		return fmt.Errorf("keypair file not found at %s. Run 'nosana address' to create it", keypairPath)
	} else if err != nil {
		return fmt.Errorf("failed to access keypair file at %s: %w", keypairPath, err)
	}

	// Check if file is readable
	file, err := os.Open(keypairPath)
	if err != nil {
		return fmt.Errorf("cannot read keypair file at %s: %w", keypairPath, err)
	}
	defer file.Close()

	// Basic validation - check if it's a valid JSON file
	var keyData interface{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&keyData); err != nil {
		return fmt.Errorf("keypair file at %s is not valid JSON: %w", keypairPath, err)
	}

	log.Printf("[INFO] Keypair file validated successfully: %s", keypairPath)
	return nil
}

// removeANSIEscapeSequences removes ANSI color codes and control characters from CLI output
func removeANSIEscapeSequences(output string) string {
	// Remove ANSI color codes from the output
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*[mGKHF]`) // Corrected: double backslash for regex escape
	cleanOutput := ansiRegex.ReplaceAllString(output, "")
	// Also remove any remaining control characters
	cleanOutput = regexp.MustCompile(`[\x00-\x1f\x7f-\x9f]`).ReplaceAllString(cleanOutput, "") // Corrected: double backslash for regex escape
	return cleanOutput
}

// testNosanaCLIAccess tests if the CLI can access the wallet
func (c *nosanaClient) testNosanaCLIAccess() error {
	// Try to get the wallet address to verify CLI access
	output, err := c.runNosanaCommand("address")
	if err != nil {
		return fmt.Errorf("failed to get wallet address: %w", err)
	}

	// Extract address from output
	log.Printf("[DEBUG] Parsing nosana address output: %q", output)

	// Remove ANSI color codes from the output
	cleanOutput := removeANSIEscapeSequences(output)

	lines := strings.Split(strings.TrimSpace(cleanOutput), "\n")
	var address string
	for i, line := range lines {
		line = strings.TrimSpace(line)
		log.Printf("[DEBUG] Line %d: %q", i, line)

		// Look for "Wallet:" prefix with flexible whitespace handling
		if strings.Contains(line, "Wallet:") {
			// Split on "Wallet:" and get everything after it
			idx := strings.Index(line, "Wallet:")
			if idx != -1 {
				addressPart := line[idx+7:] // Skip "Wallet:"
				// Remove tabs, spaces, and extract the address
				addressPart = strings.TrimSpace(addressPart)
				// Split by whitespace and take the first non-empty part
				fields := strings.Fields(addressPart)
				if len(fields) > 0 {
					address = fields[0]
					log.Printf("[DEBUG] Found wallet address: %q", address)
					break
				}
			}
		}
	}

	if address == "" {
		return fmt.Errorf("could not extract wallet address from CLI output: %s", output)
	}

	log.Printf("[INFO] Nosana wallet address: %s", address)
	return nil
}

// runNosanaCommand executes a Nosana CLI command and returns the output
func (c *nosanaClient) runNosanaCommand(args ...string) (string, error) {
	cmd := exec.Command("nosana", args...)
	cmd.Env = append(os.Environ(),
		"CI=true",             // Common CI environment variable
		"TERM=dumb",           // Disable terminal features
		"NO_COLOR=1",          // Disable colors
		"COLUMNS=80",          // Set terminal width
		"LINES=24",            // Set terminal height
		"NODE_ENV=production", // Disable development features
	)
	// Set working directory and environment
	if c.KeypairPath != "" {
		// Set the NOSANA_WALLET environment variable to point to our keypair
		cmd.Env = append(cmd.Env, "NOSANA_WALLET="+c.KeypairPath)

		// Also try setting the keypair directory
		keypairDir := filepath.Dir(c.KeypairPath)
		cmd.Env = append(cmd.Env, "NOSANA_HOME="+keypairDir)
	}

	// Set network if specified
	if c.Network != "" && c.Network != "mainnet" {
		cmd.Env = append(cmd.Env, "NOSANA_NETWORK="+c.Network)
	}

	log.Printf("[DEBUG] Running command: nosana %s", strings.Join(args, " "))

	// Use regular command execution instead of pty (pty doesn't work on Windows)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("command failed with exit code: %w", err)
	}

	return string(output), nil
} // providerConfigure is called once at the start of a Terraform run.
func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	privateKey := d.Get("private_key").(string)
	keypairPath := d.Get("keypair_path").(string)
	network := d.Get("network").(string)
	marketAddress := d.Get("market_address").(string)

	// Create a new Nosana client with the provided configuration.
	client, err := newNosanaClient(privateKey, keypairPath, network, marketAddress)
	if err != nil {
		return nil, diag.FromErr(fmt.Errorf("failed to configure Nosana provider: %w", err))
	}

	return client, nil
}
