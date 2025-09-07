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
