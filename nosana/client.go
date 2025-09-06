package nosana

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
	"bytes"

	"github.com/gagliardetto/solana-go"
	"github.com/mr-tron/base58"
)

const (
	defaultAPIURL = "https://deployment-manager.k8s.prd.nos.ci"
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
}

type NosanaAPIClient struct {
	privateKey solana.PrivateKey
	PublicKey  solana.PublicKey // Made public for testing
	baseURL    string
	httpClient *http.Client
}

func NewNosanaAPIClient(privateKeyBase58 string) (*NosanaAPIClient, error) {
	privateKeyBytes, err := base58.Decode(privateKeyBase58)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %w", err)
	}

	if len(privateKeyBytes) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key length: expected %d bytes, got %d", ed25519.PrivateKeySize, len(privateKeyBytes))
	}

	privateKey := solana.PrivateKey(privateKeyBytes)
	publicKey := privateKey.PublicKey()

	return &NosanaAPIClient{
		privateKey: privateKey,
		PublicKey:  publicKey,
		baseURL:    defaultAPIURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
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
	if body != nil {
		reqBodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
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
