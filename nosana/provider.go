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
			"cli_path": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("NOSANA_CLI_PATH", "nosana"),
				Description: "Path to the Nosana CLI executable. Defaults to 'nosana' in PATH. Can be set via NOSANA_CLI_PATH environment variable.",
			},
			"skip_cli_validation": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Skip CLI validation during provider initialization. Useful for testing or when CLI is not available.",
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

// nosanaClient represents a client for interacting with the Nosana SDK.
type nosanaClient struct {
	PrivateKey         string
	KeypairPath        string
	Network            string
	MarketAddress      string
	CLIPath            string
	SkipCLIValidation  bool
}

// newNosanaClient creates a new Nosana SDK client.
func newNosanaClient(privateKey, keypairPath, network, marketAddress, cliPath string, skipCLIValidation bool) (*nosanaClient, error) {
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
		PrivateKey:        privateKey,
		KeypairPath:       resolvedKeypairPath,
		Network:           network,
		MarketAddress:     marketAddress,
		CLIPath:           cliPath,
		SkipCLIValidation: skipCLIValidation,
	}

	// Verify Node.js and SDK are available (unless validation is skipped)
	if !skipCLIValidation {
		if err := client.validateSDKInstallation(); err != nil {
			return nil, err
		}

		log.Printf("[INFO] Nosana SDK validation completed successfully")
	} else {
		log.Printf("[WARN] SDK validation skipped - some operations may fail if Node.js or SDK is not properly configured")
	}

	log.Printf("[INFO] Nosana SDK client initialized successfully")
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

// validateSDKInstallation tests if the SDK can access the wallet
func (c *nosanaClient) validateSDKInstallation() error {
	// Use the validation script to check SDK access
	output, err := c.runNodeJSScript("nosana-validate.js")
	if err != nil {
		return fmt.Errorf("failed to validate SDK installation: %w", err)
	}

	// Parse validation output
	validationMarker := "VALIDATION_JSON:"
	errorMarker := "VALIDATION_ERROR_JSON:"
	
	var jsonData string
	if idx := strings.Index(output, validationMarker); idx != -1 {
		jsonData = strings.TrimSpace(output[idx+len(validationMarker):])
	} else if idx := strings.Index(output, errorMarker); idx != -1 {
		jsonData = strings.TrimSpace(output[idx+len(errorMarker):])
	} else {
		return fmt.Errorf("no validation result found in SDK output: %s", output)
	}

	var validationResult struct {
		Success      bool     `json:"success"`
		WalletAddress string   `json:"wallet_address"`
		SOLBalance   float64  `json:"sol_balance"`
		NOSBalance   string   `json:"nos_balance"`
		HasSOL       bool     `json:"has_sufficient_sol"`
		HasNOS       bool     `json:"has_nos_tokens"`
		Warnings     []string `json:"warnings"`
		Error        string   `json:"error"`
	}

	if err := json.Unmarshal([]byte(jsonData), &validationResult); err != nil {
		return fmt.Errorf("failed to parse validation result: %w", err)
	}

	if !validationResult.Success {
		return fmt.Errorf("SDK validation failed: %s", validationResult.Error)
	}

	log.Printf("[INFO] Nosana wallet address: %s", validationResult.WalletAddress)
	log.Printf("[INFO] SOL balance: %.6f", validationResult.SOLBalance)
	log.Printf("[INFO] NOS balance: %s", validationResult.NOSBalance)

	// Log warnings if any
	for _, warning := range validationResult.Warnings {
		log.Printf("[WARN] %s", warning)
	}

	return nil
}

// runNodeJSScript executes a Node.js script using the Nosana SDK or bundled executable
func (c *nosanaClient) runNodeJSScript(scriptName string, args ...string) (string, error) {
	// Get the path to the scripts directory
	scriptsDir := getScriptsDir()
	
	// Try bundled executable first (for distribution), then fall back to Node.js script (for development)
	bundledExecutable := getBundledExecutablePath(scriptsDir, scriptName)
	if bundledExecutable != "" {
		return c.runBundledExecutable(bundledExecutable, args...)
	}
	
	// Fallback to Node.js script for development
	return c.runNodeJSScriptDirect(scriptsDir, scriptName, args...)
}

// runBundledExecutable runs a pre-compiled executable (no Node.js required)
func (c *nosanaClient) runBundledExecutable(executablePath string, args ...string) (string, error) {
	// Prepare command arguments: executable privateKey network ...args
	cmdArgs := []string{c.PrivateKey, c.Network}
	cmdArgs = append(cmdArgs, args...)

	cmd := exec.Command(executablePath, cmdArgs...)
	cmd.Env = append(os.Environ())

	log.Printf("[DEBUG] Running bundled executable: %s with args: %v", filepath.Base(executablePath), args)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("bundled executable failed: %w", err)
	}

	return string(output), nil
}

// runNodeJSScriptDirect runs a Node.js script directly (development mode)
func (c *nosanaClient) runNodeJSScriptDirect(scriptsDir, scriptName string, args ...string) (string, error) {
	scriptPath := filepath.Join(scriptsDir, scriptName)

	// Prepare command arguments: node script.js privateKey network ...args
	cmdArgs := []string{scriptPath, c.PrivateKey, c.Network}
	cmdArgs = append(cmdArgs, args...)

	cmd := exec.Command("node", cmdArgs...)
	cmd.Env = append(os.Environ(),
		"NODE_ENV=production",
	)

	log.Printf("[DEBUG] Running Node.js script: %s with args: %v", scriptName, args)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("Node.js script failed: %w", err)
	}

	return string(output), nil
}

// getScriptsDir returns the path to the scripts directory
func getScriptsDir() string {
	// Try to find scripts directory relative to the current executable
	if ex, err := os.Executable(); err == nil {
		// Look for scripts directory next to the provider binary
		scriptsDir := filepath.Join(filepath.Dir(ex), "scripts")
		if _, err := os.Stat(scriptsDir); err == nil {
			return scriptsDir
		}
	}

	// Fallback to relative path (for development)
	if wd, err := os.Getwd(); err == nil {
		scriptsDir := filepath.Join(wd, "scripts")
		if _, err := os.Stat(scriptsDir); err == nil {
			return scriptsDir
		}
		
		// Try going up one directory (common in Go project structure)
		scriptsDir = filepath.Join(filepath.Dir(wd), "scripts")
		if _, err := os.Stat(scriptsDir); err == nil {
			return scriptsDir
		}
	}

	// Default fallback
	return "./scripts"
}

// getBundledExecutablePath checks for bundled executables and returns the path if found
func getBundledExecutablePath(scriptsDir, scriptName string) string {
	// Remove .js extension and get base name
	baseName := strings.TrimSuffix(scriptName, ".js")
	
	// Determine platform-specific executable name
	var executableName string
	switch {
	case strings.Contains(strings.ToLower(os.Getenv("OS")), "windows"):
		executableName = baseName + "-win.exe"
	case strings.Contains(strings.ToLower(os.Getenv("GOOS")), "darwin"):
		executableName = baseName + "-macos"
	default:
		executableName = baseName + "-linux"
	}
	
	// Check if bundled executable exists
	executablePath := filepath.Join(scriptsDir, executableName)
	if _, err := os.Stat(executablePath); err == nil {
		log.Printf("[DEBUG] Found bundled executable: %s", executablePath)
		return executablePath
	}
	
	// Try alternative naming (without platform suffix)
	executablePath = filepath.Join(scriptsDir, baseName)
	if _, err := os.Stat(executablePath); err == nil {
		log.Printf("[DEBUG] Found bundled executable: %s", executablePath)
		return executablePath
	}
	
	log.Printf("[DEBUG] No bundled executable found for %s, falling back to Node.js script", scriptName)
	return ""
}

// providerConfigure is called once at the start of a Terraform run.
func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	privateKey := d.Get("private_key").(string)
	keypairPath := d.Get("keypair_path").(string)
	network := d.Get("network").(string)
	marketAddress := d.Get("market_address").(string)
	cliPath := d.Get("cli_path").(string)
	skipCLIValidation := d.Get("skip_cli_validation").(bool)

	// Create a new Nosana client with the provided configuration.
	client, err := newNosanaClient(privateKey, keypairPath, network, marketAddress, cliPath, skipCLIValidation)
	if err != nil {
		return nil, diag.FromErr(fmt.Errorf("failed to configure Nosana provider: %w", err))
	}

	return client, nil
}
