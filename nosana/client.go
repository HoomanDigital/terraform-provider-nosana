package nosana

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
	"bytes"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	"github.com/mr-tron/base58"
)

const (
	defaultAPIURL = "https://deployment-manager.k8s.prd.nos.ci"
	defaultIPFSURL = "https://api.pinata.cloud"
	authMessage   = "DeploymentsAuthorization"
)

// DeploymentStrategy represents the deployment strategy
type DeploymentStrategy string

const (
	DeploymentStrategySimple       DeploymentStrategy = "SIMPLE"
	DeploymentStrategySimpleExtend DeploymentStrategy = "SIMPLE-EXTEND"
	DeploymentStrategyScheduled    DeploymentStrategy = "SCHEDULED"
	DeploymentStrategyInfinite     DeploymentStrategy = "INFINITE"
)

// DeploymentStatus represents the deployment status
type DeploymentStatus string

const (
	DeploymentStatusDraft              DeploymentStatus = "DRAFT"
	DeploymentStatusError              DeploymentStatus = "ERROR"
	DeploymentStatusStarting           DeploymentStatus = "STARTING"
	DeploymentStatusRunning            DeploymentStatus = "RUNNING"
	DeploymentStatusStopping           DeploymentStatus = "STOPPING"
	DeploymentStatusStopped            DeploymentStatus = "STOPPED"
	DeploymentStatusArchived           DeploymentStatus = "ARCHIVED"
	DeploymentStatusInsufficientFunds  DeploymentStatus = "INSUFFICIENT_FUNDS"
)

// DeploymentCreateBody represents the request body for creating a deployment
type DeploymentCreateBody struct {
	Name               string             `json:"name"`
	Market             string             `json:"market"`
	IpfsDefinitionHash *string            `json:"ipfs_definition_hash,omitempty"`
	JobDefinition      *interface{}       `json:"job_definition,omitempty"`
	Replicas           int                `json:"replicas"`
	Timeout            int                `json:"timeout"`
	Strategy           DeploymentStrategy `json:"strategy"`
	Schedule           *string            `json:"schedule,omitempty"`
}

// Deployment represents a Nosana deployment
type Deployment struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Status   DeploymentStatus  `json:"status"`
	Market   string            `json:"market"`
	Owner    string            `json:"owner"`
	Vault    string            `json:"vault"`
	Replicas int               `json:"replicas"`
	Timeout  int               `json:"timeout"`
	Strategy DeploymentStrategy `json:"strategy"`
	Schedule *string           `json:"schedule,omitempty"`
	Events   []DeploymentEvent `json:"events,omitempty"`
}

// DeploymentEvent represents an event in a deployment
type DeploymentEvent struct {
	Category     string `json:"category"`
	DeploymentId string `json:"deploymentId"`
	Type         string `json:"type"`
	Message      string `json:"message"`
	Tx           string `json:"tx,omitempty"`
	CreatedAt    string `json:"created_at"`
}

type NosanaAPIClient struct {
	privateKey solana.PrivateKey
	PublicKey  solana.PublicKey // Made public for testing
	baseURL    string
	ipfsURL    string
	httpClient *http.Client
	rpcClient  *rpc.Client // Solana RPC client
	rpcURL     string      // Store RPC URL for logging
}

func NewNosanaAPIClient(privateKeyBase58 string, rpcURL string) (*NosanaAPIClient, error) {
	privateKeyBytes, err := base58.Decode(privateKeyBase58)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %w", err)
	}

	if len(privateKeyBytes) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key length: expected %d bytes, got %d", ed25519.PrivateKeySize, len(privateKeyBytes))
	}

	privateKey := solana.PrivateKey(privateKeyBytes)
	publicKey := privateKey.PublicKey()

	// Create Solana RPC client if RPC URL is provided
	var rpcClient *rpc.Client
	if rpcURL != "" {
		rpcClient = rpc.New(rpcURL)
		log.Printf("[INFO] Initialized Solana RPC client with URL: %s", rpcURL)
	} else {
		log.Printf("[INFO] No RPC URL provided, blockchain operations will use Nosana backend")
	}

	return &NosanaAPIClient{
		privateKey: privateKey,
		PublicKey:  publicKey,
		baseURL:    defaultAPIURL,
		ipfsURL:    defaultIPFSURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		rpcClient:  rpcClient,
		rpcURL:     rpcURL,
	}, nil
}

// GetPrivateKeyString returns the private key as a base58 string
func (c *NosanaAPIClient) GetPrivateKeyString() string {
	return c.privateKey.String()
}

// GetRPCURL returns the RPC URL
func (c *NosanaAPIClient) GetRPCURL() string {
	return c.rpcURL
}

func (c *NosanaAPIClient) getAuthHeaders() (string, string, error) {
	publicKeyBase58 := c.PublicKey.String()
	timestamp := time.Now().UnixMilli()

	// Sign the auth message
	messageToSign := []byte(authMessage)
	signature, err := c.privateKey.Sign(messageToSign)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign message: %w", err)
	}

	// Base58 encode the signature
	base58EncodedSignature := base58.Encode(signature[:])
	
	// Construct the Authorization header
	authorizationHeader := fmt.Sprintf("DeploymentsAuthorization:%s:%d", base58EncodedSignature, timestamp)

	return publicKeyBase58, authorizationHeader, nil
}

func (c *NosanaAPIClient) makeRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	fullURL := fmt.Sprintf("%s%s", c.baseURL, path)

	// Generate auth headers
	xUserID, authorization, err := c.getAuthHeaders()
	if err != nil {
		return nil, fmt.Errorf("failed to generate auth headers: %w", err)
	}

	var reqBodyBytes []byte
	var contentType string

	if body != nil {
		reqBodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		contentType = "application/json"
	} else {
		// For requests without body, don't set Content-Type at all
		reqBodyBytes = nil
		contentType = ""
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	req.Header.Set("x-user-id", xUserID)
	req.Header.Set("Authorization", authorization)

	log.Printf("[DEBUG] Nosana API Request: %s %s", method, fullURL)
	log.Printf("[DEBUG] Headers: x-user-id=%s, Authorization=%s", xUserID, authorization)
	if body != nil {
		log.Printf("[DEBUG] Body: %s", string(reqBodyBytes))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	log.Printf("[DEBUG] Nosana API Response (Status: %d): %s", resp.StatusCode, string(respBody))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// CreateDeployment creates a new deployment
func (c *NosanaAPIClient) CreateDeployment(ctx context.Context, body *DeploymentCreateBody) (*Deployment, error) {
	respBody, err := c.makeRequest(ctx, "POST", "/api/deployment/create", body)
	if err != nil {
		return nil, err
	}

	var deployment Deployment
	if err := json.Unmarshal(respBody, &deployment); err != nil {
		return nil, fmt.Errorf("failed to unmarshal deployment response: %w", err)
	}

	return &deployment, nil
}

// GetDeployment retrieves a deployment by ID
func (c *NosanaAPIClient) GetDeployment(ctx context.Context, deploymentID string) (*Deployment, error) {
	path := fmt.Sprintf("/api/deployment/%s", deploymentID)
	respBody, err := c.makeRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var deployment Deployment
	if err := json.Unmarshal(respBody, &deployment); err != nil {
		return nil, fmt.Errorf("failed to unmarshal deployment response: %w", err)
	}

	return &deployment, nil
}

// DeleteDeployment archives a deployment
func (c *NosanaAPIClient) DeleteDeployment(ctx context.Context, deploymentID string) error {
	path := fmt.Sprintf("/api/deployment/%s/archive", deploymentID)
	_, err := c.makeRequest(ctx, "PATCH", path, nil)
	return err
}

// UpdateDeploymentReplicas updates the replica count for a deployment
func (c *NosanaAPIClient) UpdateDeploymentReplicas(ctx context.Context, deploymentID string, replicas int) error {
	path := fmt.Sprintf("/api/deployment/%s/update-replica-count", deploymentID)
	body := map[string]int{"replicas": replicas}
	_, err := c.makeRequest(ctx, "POST", path, body)
	return err
}

// UpdateDeploymentTimeout updates the timeout for a deployment
func (c *NosanaAPIClient) UpdateDeploymentTimeout(ctx context.Context, deploymentID string, timeout int) error {
	path := fmt.Sprintf("/api/deployment/%s/update-timeout", deploymentID)
	body := map[string]int{"timeout": timeout}
	_, err := c.makeRequest(ctx, "POST", path, body)
	return err
}

// StartDeployment starts a deployment (transitions from DRAFT to STARTING status)
func (c *NosanaAPIClient) StartDeployment(ctx context.Context, deploymentID string) (*Deployment, error) {
	path := fmt.Sprintf("/api/deployment/%s/start", deploymentID)
	respBody, err := c.makeRequest(ctx, "POST", path, nil)
	if err != nil {
		return nil, err
	}

	var deployment Deployment
	if err := json.Unmarshal(respBody, &deployment); err != nil {
		return nil, fmt.Errorf("failed to unmarshal start deployment response: %w", err)
	}

	return &deployment, nil
}

// GetBlockchainHealth checks the health of the Solana RPC connection
func (c *NosanaAPIClient) GetBlockchainHealth(ctx context.Context) error {
	if c.rpcClient == nil {
		return fmt.Errorf("no RPC client configured")
	}

	// Get latest blockhash to test RPC connectivity
	_, err := c.rpcClient.GetLatestBlockhash(ctx, rpc.CommitmentConfirmed)
	if err != nil {
		return fmt.Errorf("failed to connect to Solana RPC: %w", err)
	}

	log.Printf("[INFO] Successfully connected to Solana RPC: %s", c.rpcURL)
	return nil
}

// StartDeploymentWithRetry starts a deployment with retry logic for blockchain failures
func (c *NosanaAPIClient) StartDeploymentWithRetry(ctx context.Context, deploymentID string, maxRetries int) (*Deployment, error) {
	// Test blockchain connectivity if RPC client is available
	if c.rpcClient != nil {
		log.Printf("[INFO] Testing blockchain connectivity before starting deployment...")
		if err := c.GetBlockchainHealth(ctx); err != nil {
			log.Printf("[WARN] Blockchain health check failed: %v", err)
		} else {
			log.Printf("[INFO] Blockchain connectivity confirmed")
		}
	}

	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Printf("[INFO] Starting deployment %s (attempt %d/%d)", deploymentID, attempt, maxRetries)
		
		_, err := c.StartDeployment(ctx, deploymentID)
		if err != nil {
			log.Printf("[WARN] Start deployment attempt %d failed: %v", attempt, err)
			if attempt == maxRetries {
				return nil, fmt.Errorf("failed to start deployment after %d attempts: %w", maxRetries, err)
			}
			time.Sleep(10 * time.Second) // Wait before retry
			continue
		}

		// Check if deployment actually started by polling status
		log.Printf("[INFO] Checking deployment status after start attempt %d", attempt)
		time.Sleep(5 * time.Second) // Give it time to process
		
		deployment, err := c.GetDeployment(ctx, deploymentID)
		if err != nil {
			log.Printf("[WARN] Failed to check deployment status on attempt %d: %v", attempt, err)
			if attempt == maxRetries {
				return nil, fmt.Errorf("failed to verify deployment status after %d attempts: %w", maxRetries, err)
			}
			time.Sleep(10 * time.Second)
			continue
		}

		// If status is ERROR, check if it's a blockchain transaction failure
		if deployment.Status == DeploymentStatusError {
			log.Printf("[WARN] Deployment in ERROR status on attempt %d, checking for blockchain failures", attempt)
			
			// Check events for transaction timeouts
			hasBlockchainError := false
			for _, event := range deployment.Events {
				if strings.Contains(event.Message, "Transaction was not confirmed") || 
				   strings.Contains(event.Message, "transaction timeout") ||
				   event.Type == "JOB_LIST_ERROR" {
					hasBlockchainError = true
					log.Printf("[WARN] Found blockchain transaction error: %s", event.Message)
					break
				}
			}

			if hasBlockchainError && attempt < maxRetries {
				log.Printf("[WARN] 🔥 60-second blockchain timeout detected! (attempt %d/%d)", attempt, maxRetries)
				log.Printf("[INFO] 🔄 Implementing aggressive restart strategy...")
				
				// Immediate restart attempt when we detect the 60s timeout
				for _, event := range deployment.Events {
					if strings.Contains(event.Message, "Transaction was not confirmed in 60.00 seconds") {
						log.Printf("[INFO] 🎯 Found exact 60s timeout signature: %s", event.Message)
						
						// Extract transaction signature for monitoring
						if strings.Contains(event.Message, "Check signature ") {
							signature := extractSignature(event.Message)
							log.Printf("[INFO] 📝 Transaction signature: %s", signature)
						}
						
						// Multiple restart attempts with different strategies
						for restartAttempt := 1; restartAttempt <= 3; restartAttempt++ {
							log.Printf("[INFO] 🔄 Restart attempt %d/3 for deployment %s", restartAttempt, deploymentID)
							
							// Try restart endpoint
							_, restartErr := c.makeRequest(ctx, "POST", fmt.Sprintf("/api/deployment/%s/restart", deploymentID), nil)
							if restartErr != nil {
								log.Printf("[WARN] 💥 Restart attempt %d failed: %v", restartAttempt, restartErr)
								
								// If restart fails, try starting again from ERROR state
								log.Printf("[INFO] 🚀 Trying direct start from ERROR state...")
								_, startErr := c.StartDeployment(ctx, deploymentID)
								if startErr != nil {
									log.Printf("[WARN] 💥 Direct start failed: %v", startErr)
								} else {
									log.Printf("[INFO] ✅ Direct start succeeded!")
								}
							} else {
								log.Printf("[INFO] ✅ Restart initiated successfully")
								break
							}
							
							time.Sleep(10 * time.Second) // Wait between restart attempts
						}
						break
					}
				}
				
				// Wait for potential recovery
				time.Sleep(30 * time.Second)
				continue
			}
		}

		// Success or non-recoverable error
		return deployment, nil
	}

	return nil, fmt.Errorf("deployment start failed after %d attempts", maxRetries)
}

// extractSignature extracts the transaction signature from error messages
func extractSignature(message string) string {
	// Extract signature from: "Check signature xkEv4eefFdxb4jh8eYDENajJCJm5oYLZzicYfFnr835fZtpDKMSec1FwPQUXpNADKrCmL2AKSToqXhAgynMvQRt using the Solana Explorer"
	if strings.Contains(message, "Check signature ") {
		parts := strings.Split(message, "Check signature ")
		if len(parts) > 1 {
			sigParts := strings.Split(parts[1], " ")
			if len(sigParts) > 0 {
				return sigParts[0]
			}
		}
	}
	return "unknown"
}

// StopDeployment stops a running deployment
func (c *NosanaAPIClient) StopDeployment(ctx context.Context, deploymentID string) (*Deployment, error) {
	path := fmt.Sprintf("/api/deployment/%s/stop", deploymentID)
	respBody, err := c.makeRequest(ctx, "POST", path, nil)
	if err != nil {
		return nil, err
	}

	var deployment Deployment
	if err := json.Unmarshal(respBody, &deployment); err != nil {
		return nil, fmt.Errorf("failed to unmarshal stop deployment response: %w", err)
	}

	return &deployment, nil
}

// UploadToIPFS uploads job definition to IPFS using Pinata and returns the hash
func (c *NosanaAPIClient) UploadToIPFS(ctx context.Context, jobDefinition interface{}) (string, error) {
	// Marshal the job definition to JSON
	jobBytes, err := json.Marshal(jobDefinition)
	if err != nil {
		return "", fmt.Errorf("failed to marshal job definition: %w", err)
	}

	// Create a multipart form request for Pinata
	url := fmt.Sprintf("%s/pinning/pinFileToIPFS", c.ipfsURL)
	
	// Create form data with boundary
	boundary := "----formdata-terraform-nosana"
	body := &bytes.Buffer{}
	
	// Write form data for Pinata
	body.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	body.WriteString("Content-Disposition: form-data; name=\"file\"; filename=\"job.json\"\r\n")
	body.WriteString("Content-Type: application/json\r\n\r\n")
	body.Write(jobBytes)
	body.WriteString(fmt.Sprintf("\r\n--%s\r\n", boundary))
	
	// Add pinataMetadata
	body.WriteString("Content-Disposition: form-data; name=\"pinataMetadata\"\r\n\r\n")
	body.WriteString(`{"name":"nosana-job-definition","keyvalues":{"source":"terraform-provider"}}`)
	body.WriteString(fmt.Sprintf("\r\n--%s--\r\n", boundary))

	req, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return "", fmt.Errorf("failed to create IPFS request: %w", err)
	}

	req.Header.Set("Content-Type", fmt.Sprintf("multipart/form-data; boundary=%s", boundary))
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySW5mb3JtYXRpb24iOnsiaWQiOiI4ZWZjM2ZhZC1hZGU4LTRmMDktYmEyMy03YTI5YzY5MTQwNjUiLCJlbWFpbCI6ImRocnV2cHVyaS4zNUBnbWFpbC5jb20iLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwicGluX3BvbGljeSI6eyJyZWdpb25zIjpbeyJkZXNpcmVkUmVwbGljYXRpb25Db3VudCI6MSwiaWQiOiJGUkExIn0seyJkZXNpcmVkUmVwbGljYXRpb25Db3VudCI6MSwiaWQiOiJOWUMxIn1dLCJ2ZXJzaW9uIjoxfSwibWZhX2VuYWJsZWQiOmZhbHNlLCJzdGF0dXMiOiJBQ1RJVkUifSwiYXV0aGVudGljYXRpb25UeXBlIjoic2NvcGVkS2V5Iiwic2NvcGVkS2V5S2V5IjoiZjJlYmYzOGFiNjI4NmQ4NGEwMzkiLCJzY29wZWRLZXlTZWNyZXQiOiJhZjVhZDdkNmRmYzcxYjY1MWM4ZGYzOWNkMTRkYzA3ZWZiMTNkZmMwMDIxNjgwODQ3ZGZmMzFhNTlhYTQ3MWNhIiwiZXhwIjoxNzg4ODE2MDYzfQ.A4Vll3vfbuEraambYzekc9H671RgbRDWB9unRH1WTeQ")

	log.Printf("[DEBUG] IPFS Upload Request: POST %s", url)
	log.Printf("[DEBUG] Job Definition: %s", string(jobBytes))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("IPFS upload request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read IPFS response: %w", err)
	}

	log.Printf("[DEBUG] IPFS Response (Status: %d): %s", resp.StatusCode, string(respBody))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("IPFS upload failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse the Pinata response to get the hash
	var pinataResponse struct {
		IpfsHash    string `json:"IpfsHash"`
		PinSize     int    `json:"PinSize"`
		Timestamp   string `json:"Timestamp"`
		IsDuplicate bool   `json:"isDuplicate"`
	}

	if err := json.Unmarshal(respBody, &pinataResponse); err != nil {
		return "", fmt.Errorf("failed to parse Pinata response: %w", err)
	}

	if pinataResponse.IpfsHash == "" {
		return "", fmt.Errorf("Pinata response did not contain an IPFS hash")
	}

	log.Printf("[INFO] IPFS upload successful. Hash: %s (Size: %d bytes)", pinataResponse.IpfsHash, pinataResponse.PinSize)
	return pinataResponse.IpfsHash, nil
}

// GetDeploymentWithEvents gets detailed deployment info including events for debugging
func (c *NosanaAPIClient) GetDeploymentWithEvents(ctx context.Context, deploymentID string) (map[string]interface{}, error) {
	// Get authentication headers
	userID, authHeader, err := c.getAuthHeaders()
	if err != nil {
		return nil, fmt.Errorf("failed to get auth headers: %w", err)
	}

	// Get detailed deployment info via API
	url := fmt.Sprintf("https://deployment-manager.k8s.prd.nos.ci/api/deployment/%s", deploymentID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("x-user-id", userID)
	req.Header.Set("Authorization", authHeader)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get deployment with events failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response, nil
}

// CreateDeploymentSDK creates deployment exactly like SDK createDeployment pattern
func (c *NosanaAPIClient) CreateDeploymentSDK(ctx context.Context, deploymentReq map[string]interface{}) (map[string]interface{}, error) {
	log.Printf("[INFO] 🔗 SDK createDeployment pattern - POST /api/deployment")
	
	// Get authentication headers
	userID, authHeader, err := c.getAuthHeaders()
	if err != nil {
		return nil, fmt.Errorf("failed to get auth headers: %w", err)
	}

	// Convert to JSON exactly like SDK
	reqBody, err := json.Marshal(deploymentReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal deployment request: %w", err)
	}

	log.Printf("[DEBUG] SDK createDeployment request: %s", string(reqBody))

	// POST /api/deployment/create exactly like SDK
	req, err := http.NewRequestWithContext(ctx, "POST", "https://deployment-manager.k8s.prd.nos.ci/api/deployment/create", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-user-id", userID)
	req.Header.Set("Authorization", authHeader)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	log.Printf("[DEBUG] SDK createDeployment response: %s", string(body))

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("SDK createDeployment failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to decode createDeployment response: %w", err)
	}

	return response, nil
}

// StartDeploymentSDK starts deployment exactly like SDK deploymentStart pattern
func (c *NosanaAPIClient) StartDeploymentSDK(ctx context.Context, deploymentID string) error {
	log.Printf("[INFO] 🚀 SDK deploymentStart pattern - POST /api/deployment/{deployment}/start")
	
	// Get authentication headers
	userID, authHeader, err := c.getAuthHeaders()
	if err != nil {
		return fmt.Errorf("failed to get auth headers: %w", err)
	}

	// POST /api/deployment/{deployment}/start exactly like SDK deploymentStart.ts
	url := fmt.Sprintf("https://deployment-manager.k8s.prd.nos.ci/api/deployment/%s/start", deploymentID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// No Content-Type header for empty body (like SDK deploymentStart.ts)
	req.Header.Set("x-user-id", userID)
	req.Header.Set("Authorization", authHeader)

	log.Printf("[DEBUG] SDK deploymentStart request: %s", url)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	log.Printf("[DEBUG] SDK deploymentStart response: %s", string(body))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("SDK deploymentStart failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// FundVault implements the exact SDK vault funding pattern
func (c *NosanaAPIClient) FundVault(ctx context.Context, vaultAddress, marketAddress string) error {
	log.Printf("[INFO] 🏦 Starting vault funding for %s (SDK pattern)", vaultAddress)
	
	// Step 1: Check current vault balance (like SDK vaultGetBalance)
	log.Printf("[INFO] 🔍 Checking vault balance via API...")
	balance, err := c.GetVaultBalance(ctx, vaultAddress)
	if err != nil {
		return fmt.Errorf("failed to get vault balance: %w", err)
	}
	
	log.Printf("[INFO] 💰 Current vault balance: SOL=%.6f, NOS=%.6f", balance["SOL"], balance["NOS"])
	
	// Step 2: Determine topup amounts (adaptive to available wallet balance)
	solNeeded := 0.003   // increased to cover SOL transfer + ATA rent (~0.002 SOL) + fees
	nosNeeded := 0.0     // will be set based on market job price and SOL availability
	
	// Query market to get actual job price (like SDK does)
	if c.rpcURL != "" {
		rpcClient := rpc.New(c.rpcURL)
		marketPubkey := solana.MustPublicKeyFromBase58(marketAddress)
		marketInfo, merr := rpcClient.GetAccountInfo(ctx, marketPubkey)
		if merr == nil && marketInfo != nil && len(marketInfo.Value.Data.GetBinary()) >= 48 {
			// Parse market account data to extract actual job price
			// Market struct: authority(32) + jobExpiration(8) + jobPrice(8) + ...
			// jobPrice is at offset 40-47 (u64 little-endian)
			data := marketInfo.Value.Data.GetBinary()
			jobPriceBytes := data[40:48]
			
			// Parse u64 little-endian (NOS price in microlamports)
			var jobPriceMicroNOS uint64
			for i := 0; i < 8; i++ {
				jobPriceMicroNOS |= uint64(jobPriceBytes[i]) << (8 * i)
			}
			
			// Convert microlamports to NOS (6 decimals)
			marketJobPrice := float64(jobPriceMicroNOS) / 1e6
			if marketJobPrice == 0 {
				nosNeeded = 0.0 // free market - no NOS needed
				log.Printf("[INFO] 🏪 Market %s: FREE (job price = 0 NOS)", marketAddress)
			} else {
				// Check if wallet has enough SOL for both SOL transfer + NOS ATA creation (~0.003 SOL total)
				rpcClient2 := rpc.New(c.rpcURL)
				balRes, berr := rpcClient2.GetBalance(ctx, c.PublicKey, rpc.CommitmentConfirmed)
				if berr == nil {
					walletSOL := float64(balRes.Value) / 1e9
					minSOLForNOSTransfer := 0.003 // ~0.001 SOL transfer + 0.002 SOL ATA rent
					if walletSOL >= minSOLForNOSTransfer {
						nosNeeded = marketJobPrice
						log.Printf("[INFO] 🏪 Market %s: job price = %.6f NOS (will fund with exact amount)", marketAddress, nosNeeded)
					} else {
						log.Printf("[INFO] 🏪 Market %s: job price = %.6f NOS (insufficient SOL %.6f < %.6f for ATA creation)", marketAddress, marketJobPrice, walletSOL, minSOLForNOSTransfer)
						nosNeeded = 0.0  // skip NOS until wallet has more SOL
					}
				} else {
					log.Printf("[INFO] 🏪 Market %s: job price = %.6f NOS (skipping due to balance check error)", marketAddress, marketJobPrice)
					nosNeeded = 0.0
				}
			}
		} else {
			log.Printf("[WARN] Could not query market %s or insufficient data, using default", marketAddress)
			nosNeeded = 0.0  // skip NOS if market query fails
		}
	}

	// Adapt SOL top-up based on actual wallet balance (keep tiny buffer for fees)
	if c.rpcURL != "" {
		rpcClient := rpc.New(c.rpcURL)
		balRes, berr := rpcClient.GetBalance(ctx, c.PublicKey, rpc.CommitmentConfirmed)
		if berr == nil {
			walletSOL := float64(balRes.Value) / 1e9
			// With 0.001 SOL available, use most of it but keep buffer for fees
			maxAffordable := walletSOL - 0.0001  // keep 0.0001 SOL for transaction fees
			if maxAffordable < 0 {
				maxAffordable = 0
			}
			if maxAffordable < solNeeded {
				solNeeded = maxAffordable
			}
			log.Printf("[INFO] 💳 Wallet %.6f SOL + 39+ NOS, will top-up %.6f SOL (NOS skipped)", walletSOL, solNeeded)
		} else {
			log.Printf("[WARN] Could not read wallet balance, using default SOL topup %.6f: %v", solNeeded, berr)
		}
	}

	if solNeeded <= 0 {
		return fmt.Errorf("insufficient SOL for transaction fees; need at least 0.0001 SOL")
	}

	if balance["SOL"] < solNeeded || balance["NOS"] < nosNeeded {
		log.Printf("[INFO] 💸 Vault needs funding - adding SOL=%.6f, NOS=%.6f", solNeeded, nosNeeded)
		
		// Use Nosana's backend to fund the vault (simulating TokenManager)
		err = c.TopupVault(ctx, vaultAddress, solNeeded, nosNeeded)
		if err != nil {
			return fmt.Errorf("failed to topup vault: %w", err)
		}
		
		log.Printf("[INFO] ✅ Vault funding completed!")
		
		// Step 3: Update vault balance in backend (like SDK updateBalance)
		err = c.UpdateVaultBalance(ctx, vaultAddress)
		if err != nil {
			log.Printf("[WARN] Failed to update vault balance in backend: %v", err)
		}
	} else {
		log.Printf("[INFO] ✅ Vault already has sufficient funds")
	}
	
	return nil
}

// GetVaultBalance gets vault balance via Nosana API
func (c *NosanaAPIClient) GetVaultBalance(ctx context.Context, vaultAddress string) (map[string]float64, error) {
	// Get authentication headers
	userID, authHeader, err := c.getAuthHeaders()
	if err != nil {
		return nil, fmt.Errorf("failed to get auth headers: %w", err)
	}

	// Check if there's a vault balance endpoint
	url := fmt.Sprintf("https://deployment-manager.k8s.prd.nos.ci/api/vault/%s", vaultAddress)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("x-user-id", userID)
	req.Header.Set("Authorization", authHeader)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	
	if resp.StatusCode == 404 {
		// Vault doesn't exist yet or balance endpoint not available - assume empty
		log.Printf("[INFO] Vault balance endpoint not available, assuming empty vault")
		return map[string]float64{"SOL": 0, "NOS": 0}, nil
	}
	
	if resp.StatusCode != http.StatusOK {
		log.Printf("[WARN] Vault balance check failed with status %d: %s", resp.StatusCode, string(body))
		// Assume empty vault if we can't check balance
		return map[string]float64{"SOL": 0, "NOS": 0}, nil
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		// If we can't parse, assume empty vault
		return map[string]float64{"SOL": 0, "NOS": 0}, nil
	}

	// Extract balance values
	solBalance := 0.0
	nosBalance := 0.0
	
	if sol, ok := response["sol_balance"].(float64); ok {
		solBalance = sol
	}
	if nos, ok := response["nos_balance"].(float64); ok {
		nosBalance = nos
	}

	return map[string]float64{"SOL": solBalance, "NOS": nosBalance}, nil
}

// TopupVault implements EXACT SDK TokenManager pattern for vault funding
func (c *NosanaAPIClient) TopupVault(ctx context.Context, vaultAddress string, solAmount, nosAmount float64) error {
	log.Printf("[INFO] 🎯 EXACT SDK TokenManager: Funding vault %s with SOL=%.6f + NOS=%.6f", vaultAddress, solAmount, nosAmount)
	
	// Step 1: Setup connection and addresses (exact SDK pattern)
	client := rpc.New(c.rpcURL)
	
	vaultPubkey, err := solana.PublicKeyFromBase58(vaultAddress)
	if err != nil {
		return fmt.Errorf("invalid vault address: %w", err)
	}
	
	walletKey := c.privateKey
	walletPubkey := walletKey.PublicKey()
	
	log.Printf("[INFO] 🏦 TokenManager: source=%s → destination=%s", walletPubkey, vaultPubkey)
	
	// Step 2: Create instruction slice (exact SDK TokenManager constructor)
	instructions := []solana.Instruction{}
	
	// Step 3: Add SOL transfer instruction (SDK addSOL pattern)
	if solAmount > 0 {
		err = c.addSOLInstruction(ctx, client, &instructions, walletPubkey, vaultPubkey, solAmount)
		if err != nil {
			return fmt.Errorf("failed to add SOL instruction: %w", err)
		}
		log.Printf("[INFO] ✅ Added SOL transfer instruction: %.6f SOL", solAmount)
	}
	
	// Step 4: Add NOS transfer instruction (SDK addNOS pattern)
	if nosAmount > 0 {
		err = c.addNOSInstruction(ctx, client, &instructions, walletPubkey, vaultPubkey, nosAmount)
		if err != nil {
			log.Printf("[WARN] Failed to add NOS instruction (SOL-only transfer): %v", err)
		} else {
			log.Printf("[INFO] ✅ Added NOS transfer instruction: %.6f NOS", nosAmount)
		}
	}
	
	// Step 5: Execute batched transaction (SDK transfer pattern)
	if len(instructions) > 0 {
		err = c.executeTokenManagerTransaction(ctx, client, instructions, walletKey)
		if err != nil {
			return fmt.Errorf("TokenManager transaction failed: %w", err)
		}
		log.Printf("[INFO] 🎉 EXACT SDK TokenManager completed successfully!")
	} else {
		log.Printf("[INFO] ⚠️ No transfer instructions added - skipping transaction")
	}
	
	return nil
}

// addSOLInstruction implements exact SDK createTransferSOLInstruction pattern
func (c *NosanaAPIClient) addSOLInstruction(ctx context.Context, client *rpc.Client, instructions *[]solana.Instruction, source, destination solana.PublicKey, amount float64) error {
	log.Printf("[INFO] 🪙 SDK addSOL: Adding SOL transfer instruction for %.6f SOL", amount)
	
	// Convert SOL to lamports (exact SDK pattern)
	lamports := uint64(amount * 1e9)
	
	// Balance check (exact SDK pattern)
	balanceResult, err := client.GetBalance(ctx, source, rpc.CommitmentConfirmed)
	if err != nil {
		return fmt.Errorf("failed to get source balance: %w", err)
	}
	
	balance := balanceResult.Value
	log.Printf("[INFO] 💰 Wallet balance: %.6f SOL, required: %.6f SOL", float64(balance)/1e9, amount)
	
	if balance < lamports {
		return fmt.Errorf("insufficient SOL balance: have %.6f, need %.6f", float64(balance)/1e9, amount)
	}
	
	// Create SystemProgram.transfer instruction (exact SDK pattern)
	transferInst := system.NewTransferInstruction(
		lamports,
		source,
		destination,
	).Build()
	
	// Add to instruction slice (exact SDK pattern)
	*instructions = append(*instructions, transferInst)
	
	log.Printf("[INFO] ✅ SOL instruction added to TokenManager transaction")
	return nil
}

// addNOSInstruction implements exact SDK createTransferNOSInstruction pattern  
func (c *NosanaAPIClient) addNOSInstruction(ctx context.Context, client *rpc.Client, instructions *[]solana.Instruction, source, destination solana.PublicKey, amount float64) error {
	log.Printf("[INFO] 🟡 SDK addNOS: Adding NOS transfer instruction for %.6f NOS", amount)
	
	// NOS token address on mainnet (from SDK config)
	nosTokenMint := solana.MustPublicKeyFromBase58("nosXBVoaCTtYdLvKY6Csb4AC8JCdQKKAaWYtx2ZMoo7")
	
	// Convert NOS to token units (NOS has 6 decimals - exact SDK pattern)
	tokenAmount := uint64(amount * 1e6)
	
	// Get source token account with balance check (exact SDK pattern)
	sourceATA, sourceBalance, err := c.getNOSTokenAccountWithBalance(ctx, client, source, nosTokenMint)
	if err != nil {
		return fmt.Errorf("failed to get source NOS account: %w", err)
	}
	
	if sourceBalance < tokenAmount {
		return fmt.Errorf("insufficient NOS balance: have %.6f, need %.6f", float64(sourceBalance)/1e6, amount)
	}
	
	log.Printf("[INFO] 💰 NOS balance: %.6f NOS, required: %.6f NOS", float64(sourceBalance)/1e6, amount)
	
	// Get destination token account (create if needed - exact SDK pattern)
	destATA, needsDestinationAccount, err := c.getNOSTokenAccountForDestination(ctx, client, destination, nosTokenMint)
	if err != nil {
		return fmt.Errorf("failed to get destination NOS account: %w", err)
	}
	
	// Add create ATA instruction if needed (exact SDK pattern)
	if needsDestinationAccount {
		createATAInst, err := c.createAssociatedTokenAccountInstruction(source, destATA, destination, nosTokenMint)
		if err != nil {
			return fmt.Errorf("failed to create ATA instruction: %w", err)
		}
		*instructions = append(*instructions, createATAInst)
		log.Printf("[INFO] ✅ Added create destination ATA instruction")
	}
	
	// Add token transfer instruction (exact SDK pattern)
	transferInst, err := c.createSPLTokenTransferInstruction(sourceATA, destATA, source, tokenAmount)
	if err != nil {
		return fmt.Errorf("failed to create token transfer instruction: %w", err)
	}
	
	*instructions = append(*instructions, transferInst)
	
	log.Printf("[INFO] ✅ NOS instruction added to TokenManager transaction")
	return nil
}

// executeTokenManagerTransaction implements exact SDK sendAndConfirmTransaction pattern
func (c *NosanaAPIClient) executeTokenManagerTransaction(ctx context.Context, client *rpc.Client, instructions []solana.Instruction, signer solana.PrivateKey) error {
	log.Printf("[INFO] 🚀 SDK transfer: Executing TokenManager transaction with %d instructions", len(instructions))
	
	// Get latest blockhash (exact SDK pattern)
	latestBlockhash, err := client.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return fmt.Errorf("failed to get latest blockhash: %w", err)
	}
	
	// Create proper transaction with blockhash and fee payer (exact SDK pattern)
	signerPubkey := signer.PublicKey()
	finalTx, err := solana.NewTransaction(
		instructions,
		latestBlockhash.Value.Blockhash,
		solana.TransactionPayer(signerPubkey),
	)
	if err != nil {
		return fmt.Errorf("failed to create final transaction: %w", err)
	}
	
	// Sign transaction (exact SDK pattern)
	_, err = finalTx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(signerPubkey) {
			return &signer
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to sign transaction: %w", err)
	}
	
	// Send and confirm transaction (exact SDK pattern)
	opts := rpc.TransactionOpts{
		SkipPreflight:       false,
		PreflightCommitment: rpc.CommitmentConfirmed,
	}
	
	signature, err := client.SendTransactionWithOpts(ctx, finalTx, opts)
	if err != nil {
		return fmt.Errorf("failed to send transaction: %w", err)
	}
	
	log.Printf("[INFO] ✅ TokenManager transaction sent: %s", signature)
	
	// Wait for confirmation (exact SDK pattern)
	err = c.waitForConfirmation(ctx, client, signature)
	if err != nil {
		return fmt.Errorf("transaction confirmation failed: %w", err)
	}
	
	log.Printf("[INFO] ✅ TokenManager transaction confirmed on blockchain!")
	return nil
}

// Helper functions implementing exact SDK patterns

// getNOSTokenAccountWithBalance implements exact SDK getNosTokenAddressForAccount pattern
func (c *NosanaAPIClient) getNOSTokenAccountWithBalance(ctx context.Context, client *rpc.Client, wallet, mint solana.PublicKey) (solana.PublicKey, uint64, error) {
	// Calculate associated token account address (exact SDK pattern)
	ata, _, err := solana.FindAssociatedTokenAddress(wallet, mint)
	if err != nil {
		return solana.PublicKey{}, 0, fmt.Errorf("failed to find associated token address: %w", err)
	}
	
	// Get account info and balance (exact SDK pattern)
	accountInfo, err := client.GetAccountInfo(ctx, ata)
	if err != nil {
		return solana.PublicKey{}, 0, fmt.Errorf("NOS token account does not exist on source")
	}
	
	// Parse token account data to get balance (simplified)
	// In production, you'd parse the full SPL token account data
	if len(accountInfo.Value.Data.GetBinary()) < 64 {
		return solana.PublicKey{}, 0, fmt.Errorf("invalid token account data")
	}
	
	// Extract balance from token account data (bytes 64-72)
	balanceBytes := accountInfo.Value.Data.GetBinary()[64:72]
	balance := uint64(0)
	for i, b := range balanceBytes {
		balance |= uint64(b) << (8 * i)
	}
	
	return ata, balance, nil
}

// getNOSTokenAccountForDestination implements exact SDK pattern for destination accounts
func (c *NosanaAPIClient) getNOSTokenAccountForDestination(ctx context.Context, client *rpc.Client, wallet, mint solana.PublicKey) (solana.PublicKey, bool, error) {
	// Calculate associated token account address (exact SDK pattern)
	ata, _, err := solana.FindAssociatedTokenAddress(wallet, mint)
	if err != nil {
		return solana.PublicKey{}, false, fmt.Errorf("failed to find associated token address: %w", err)
	}
	
	// Check if account exists (exact SDK pattern)
	_, err = client.GetAccountInfo(ctx, ata)
	if err != nil {
		// Account doesn't exist - need to create it (exact SDK pattern)
		return ata, true, nil
	}
	
	// Account exists
	return ata, false, nil
}

// createAssociatedTokenAccountInstruction implements exact SDK createAssociatedTokenAccountInstruction
func (c *NosanaAPIClient) createAssociatedTokenAccountInstruction(payer, ata, owner, mint solana.PublicKey) (solana.Instruction, error) {
	// Create the associated token account (program derives ata internally)
	return associatedtokenaccount.NewCreateInstruction(
		payer,
		owner,
		mint,
	).Build(), nil
}

// createSPLTokenTransferInstruction implements exact SDK createTransferInstruction pattern
func (c *NosanaAPIClient) createSPLTokenTransferInstruction(source, destination, owner solana.PublicKey, amount uint64) (solana.Instruction, error) {
	// NOS token mint address
	nosTokenMint := solana.MustPublicKeyFromBase58("nosXBVoaCTtYdLvKY6Csb4AC8JCdQKKAaWYtx2ZMoo7")
	
	// Proper SPL token transfer (NOS has 6 decimals)
	return token.NewTransferCheckedInstruction(
		amount,
		6,
		source,
		nosTokenMint,
		destination,
		owner,
		nil,
	).Build(), nil
}

func (c *NosanaAPIClient) waitForConfirmation(ctx context.Context, client *rpc.Client, signature solana.Signature) error {
	log.Printf("[INFO] ⏳ Waiting for transaction confirmation: %s", signature)
	
	// Poll for confirmation with timeout
	timeout := time.After(60 * time.Second)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-timeout:
			return fmt.Errorf("transaction confirmation timeout after 60 seconds")
		case <-ticker.C:
			// Check transaction status
			statuses, err := client.GetSignatureStatuses(ctx, true, signature)
			if err != nil {
				log.Printf("[WARN] Failed to get signature status: %v", err)
				continue
			}
			
			if len(statuses.Value) > 0 && statuses.Value[0] != nil {
				status := statuses.Value[0]
				if status.Err != nil {
					return fmt.Errorf("transaction failed: %v", status.Err)
				}
				
				if status.ConfirmationStatus == rpc.ConfirmationStatusConfirmed || 
				   status.ConfirmationStatus == rpc.ConfirmationStatusFinalized {
					log.Printf("[INFO] ✅ Transaction confirmed with status: %s", status.ConfirmationStatus)
					return nil
				}
			}
		}
	}
}

// UpdateVaultBalance updates vault balance in Nosana backend (like SDK)
func (c *NosanaAPIClient) UpdateVaultBalance(ctx context.Context, vaultAddress string) error {
	// Get authentication headers
	userID, authHeader, err := c.getAuthHeaders()
	if err != nil {
		return fmt.Errorf("failed to get auth headers: %w", err)
	}

	// PATCH /api/vault/{vault}/update-balance (exact SDK pattern)
	url := fmt.Sprintf("https://deployment-manager.k8s.prd.nos.ci/api/vault/%s/update-balance", vaultAddress)
	
	// Send empty body to trigger balance update
	reqBody, _ := json.Marshal(map[string]interface{}{})
	
	req, err := http.NewRequestWithContext(ctx, "PATCH", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-user-id", userID)
	req.Header.Set("Authorization", authHeader)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	log.Printf("[DEBUG] Update vault balance response: %s", string(body))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("update vault balance failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
