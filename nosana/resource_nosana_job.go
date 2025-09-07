// resource_nosana_job.go  
package nosana

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// resourceNosanaJob defines the schema and CRUD operations for the nosana_job resource.
func resourceNosanaJob() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNosanaJobCreate,
		ReadContext:   resourceNosanaJobRead,
		UpdateContext: resourceNosanaJobUpdate,
		DeleteContext: resourceNosanaJobDelete,
		Schema: map[string]*schema.Schema{
			"job_content": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "JSON-encoded job definition content.",
				ValidateFunc: validation.StringIsJSON,
			},
			"replicas": {
				Type:         schema.TypeInt,
				Required:     true,
				Description:  "Number of replicas for the deployment.",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"timeout": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      600,
				Description:  "Timeout in seconds for the deployment.",
				ValidateFunc: validation.IntAtLeast(60),
			},
			"strategy": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Deployment strategy (e.g., SIMPLE, INFINITE, SCHEDULED).",
				ValidateFunc: validation.StringInSlice([]string{"SIMPLE", "SIMPLE-EXTEND", "SCHEDULED", "INFINITE"}, false),
			},
			"schedule": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Cron expression for SCHEDULED strategy.",
				// TODO: Add cron expression validation
			},
			"auto_start": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "If true, automatically start the deployment after creation. If false, leaves it in DRAFT status.",
			},
			"wait_for_completion": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "If true, Terraform will wait for the deployment to reach a stable state.",
			},
			"completion_timeout_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      300, // 5 minutes
				Description:  "Maximum time (in seconds) to wait for deployment completion.",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"max_retries": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     8, // Increased to handle 60s timeout better
				Description: "Maximum number of retry attempts for deployment start failures",
			},
			// Attributes that will be stored in the Terraform state after creation
			"job_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The unique ID of the Nosana deployment.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The current status of the Nosana job (QUEUED, RUNNING, COMPLETED, etc.).",
			},
		"ipfs_hash": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The IPFS hash of the job definition.",
		},
		"deployment_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The deployment ID from the Nosana API.",
		},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext, // For MVP, basic import functionality
		},
	}
}

// resourceNosanaJobCreate handles the creation of a Nosana job.
func resourceNosanaJobCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*nosanaClient)
	jobContentStr := d.Get("job_content").(string)
	replicas := d.Get("replicas").(int)
	timeout := d.Get("timeout").(int)
	strategy := DeploymentStrategy(d.Get("strategy").(string))
	// Schedule is optional
	scheduleVal, ok := d.GetOk("schedule")
	var schedule *string
	if ok {
		sched := scheduleVal.(string)
		schedule = &sched
	}

	waitForCompletion := d.Get("wait_for_completion").(bool)
	completionTimeoutSeconds := d.Get("completion_timeout_seconds").(int)
	marketAddress := client.MarketAddress

	// Generate a unique name for the deployment using uuid
	deploymentName := "terraform-deployment-" + uuid.New().String()

	// Parse the job content JSON
	var jobContent interface{}
	if err := json.Unmarshal([]byte(jobContentStr), &jobContent); err != nil {
		return diag.FromErr(fmt.Errorf("failed to parse job_content JSON: %w", err))
	}

	// Upload job definition to IPFS
	log.Printf("[INFO] Uploading job definition to IPFS...")
	ipfsHash, err := client.APIClient.UploadToIPFS(ctx, jobContent)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to upload job definition to IPFS: %w", err))
	}

	log.Printf("[INFO] Job definition uploaded to IPFS with hash: %s", ipfsHash)

	// Create deployment with IPFS hash
	createBody := &DeploymentCreateBody{
		Name:               deploymentName,
		Market:             marketAddress,
		IpfsDefinitionHash: &ipfsHash,
		Replicas:           replicas,
		Timeout:            timeout,
		Strategy:           strategy,
		Schedule:           schedule,
	}

	// Create the Nosana Deployment via API
	log.Printf("[INFO] Creating Nosana deployment with IPFS hash: %s", ipfsHash)
	newDeployment, err := client.APIClient.CreateDeployment(ctx, createBody)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to create Nosana deployment: %w", err))
	}

	log.Printf("[INFO] Nosana deployment %s created with status: %s", newDeployment.ID, newDeployment.Status)

	// Now START the deployment to transition from DRAFT → STARTING → RUNNING
	maxRetries := d.Get("max_retries").(int)
	log.Printf("[INFO] 🚀 Starting deployment %s (DRAFT → STARTING → RUNNING)...", newDeployment.ID)
	
	startedDeployment, err := client.APIClient.StartDeploymentWithRetry(ctx, newDeployment.ID, maxRetries)
	if err != nil {
		// If start fails, still set deployment info so user can debug
		d.SetId(newDeployment.ID)
		d.Set("job_id", newDeployment.ID)
		d.Set("status", "DRAFT")
		d.Set("ipfs_hash", ipfsHash)
		d.Set("deployment_id", newDeployment.ID)
		
		return diag.FromErr(fmt.Errorf("deployment created but failed to start: %w", err))
	}
	
	log.Printf("[INFO] ✅ Deployment %s started successfully! Status: %s", startedDeployment.ID, startedDeployment.Status)
	
	// Set Terraform state with the started deployment
	d.SetId(newDeployment.ID)
	d.Set("job_id", newDeployment.ID)
	d.Set("status", string(startedDeployment.Status))
	d.Set("ipfs_hash", ipfsHash)
	d.Set("deployment_id", newDeployment.ID)
	
	log.Printf("[INFO] 🎉 DEPLOYMENT STARTED ON NOSANA NETWORK!")
	log.Printf("[INFO] 🆔 Deployment ID: %s", newDeployment.ID)
	log.Printf("[INFO] 📊 Status: %s", startedDeployment.Status)
	log.Printf("[INFO] 🌐 Check your deployment at dashboard.nosana.com/account/deployer")

	// 2. Optional: Wait for deployment completion (status becomes RUNNING or COMPLETED)
	if waitForCompletion {
		log.Printf("[INFO] Waiting for Nosana deployment %s to reach a stable state (timeout: %d seconds)...", newDeployment.ID, completionTimeoutSeconds)
		timeoutChan := time.After(time.Duration(completionTimeoutSeconds) * time.Second)
		tick := time.NewTicker(5 * time.Second) // Poll every 5 seconds
		defer tick.Stop()

		retryCount := 0
		maxAutoRetries := 3 // Allow up to 3 automatic restarts

		for {
			select {
			case <-timeoutChan:
				return diag.Errorf("timeout waiting for Nosana deployment %s to complete after %d seconds", newDeployment.ID, completionTimeoutSeconds)
			case <-tick.C:
				currentDeployment, err := client.APIClient.GetDeployment(ctx, newDeployment.ID)
				if err != nil {
					log.Printf("[WARN] Error polling deployment status for %s: %v", newDeployment.ID, err)
					continue
				}

				d.Set("status", currentDeployment.Status) // Update status in state

				// Check if deployment failed with blockchain timeout and auto-restart if possible
				if currentDeployment.Status == DeploymentStatusError && retryCount < maxAutoRetries {
					log.Printf("[WARN] Deployment %s failed with ERROR status, checking for blockchain timeout...", newDeployment.ID)
					
					// Check events for blockchain timeout
					hasBlockchainTimeout := false
					for _, event := range currentDeployment.Events {
						if strings.Contains(event.Message, "Transaction was not confirmed") || 
						   strings.Contains(event.Message, "transaction timeout") ||
						   event.Type == "JOB_LIST_ERROR" {
							hasBlockchainTimeout = true
							log.Printf("[WARN] Found blockchain timeout: %s", event.Message)
							break
						}
					}

					if hasBlockchainTimeout {
						retryCount++
						log.Printf("[INFO] Attempting automatic restart of deployment %s (attempt %d/%d)...", newDeployment.ID, retryCount, maxAutoRetries)
						
						// Wait before restart attempt
						time.Sleep(45 * time.Second)
						
						// Try to restart the deployment
						_, err := client.APIClient.StartDeploymentWithRetry(ctx, newDeployment.ID, 3)
						if err != nil {
							log.Printf("[WARN] Auto-restart attempt %d failed: %v", retryCount, err)
							if retryCount >= maxAutoRetries {
								return diag.Errorf("deployment %s failed after %d restart attempts due to blockchain timeouts", newDeployment.ID, maxAutoRetries)
							}
						} else {
							log.Printf("[INFO] Auto-restart attempt %d successful, continuing to monitor...", retryCount)
						}
						continue
					}
				}

				if currentDeployment.Status == DeploymentStatusRunning || currentDeployment.Status == DeploymentStatusStopped || currentDeployment.Status == DeploymentStatusArchived || currentDeployment.Status == DeploymentStatusError || currentDeployment.Status == DeploymentStatusInsufficientFunds {
					log.Printf("[INFO] Nosana deployment %s reached stable state: %s.", newDeployment.ID, currentDeployment.Status)
					return resourceNosanaJobRead(ctx, d, m) // Read to ensure state is up-to-date
				}
				log.Printf("[INFO] Nosana deployment %s current status: %s", newDeployment.ID, currentDeployment.Status)
			}
		}
	}

	return resourceNosanaJobRead(ctx, d, m)
}

// resourceNosanaJobRead handles reading the state of a Nosana job.
func resourceNosanaJobRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*nosanaClient)
	deploymentID := d.Id() // Get the deployment ID from the Terraform state

	// 1. Get deployment status from Nosana API
	deployment, err := client.APIClient.GetDeployment(ctx, deploymentID)
	if err != nil {
		// Check if it's a 404 Not Found error from the API
		if strings.Contains(err.Error(), "API request failed with status 404") {
			log.Printf("[WARN] Nosana deployment %s not found (404), removing from state.", deploymentID)
			d.SetId("") // Mark resource for deletion from state
			return nil
		}
		return diag.FromErr(fmt.Errorf("failed to read Nosana deployment %s: %w", deploymentID, err))
	}

	// 2. Update Terraform state with current deployment details
	d.Set("job_id", deployment.ID)
	d.Set("status", deployment.Status)

	// Since job_content is user-provided and not returned by API, we keep the last known value.
	// If ipfs_definition_hash were truly managed by IPFS, we would verify it here.

	d.Set("replicas", deployment.Replicas)
	d.Set("timeout", deployment.Timeout)
	d.Set("strategy", deployment.Strategy)
	if deployment.Schedule != nil {
		d.Set("schedule", *deployment.Schedule)
	}

	log.Printf("[INFO] Read Nosana deployment %s. Status: %s", deployment.ID, deployment.Status)
	return nil
}

// resourceNosanaJobUpdate handles updates to a Nosana job.
func resourceNosanaJobUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*nosanaClient)
	deploymentID := d.Id()

	if d.HasChange("job_content") {
		// If job_content changes, force a replacement. This is because a change to job_content
		// implies a new IPFS hash, which the Nosana API treats as a new deployment.
		// Terraform handles this by planning a destroy and then a create.
		return diag.Errorf("changes to `job_content` require a new resource to be created, please use `terraform taint` or plan a replacement")
	}

	if d.HasChange("replicas") {
		old, new := d.GetChange("replicas")
		log.Printf("[INFO] Updating replicas for deployment %s from %d to %d", deploymentID, old, new)
		err := client.APIClient.UpdateDeploymentReplicas(ctx, deploymentID, new.(int))
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to update replicas for deployment %s: %w", deploymentID, err))
		}
	}

	if d.HasChange("timeout") {
		old, new := d.GetChange("timeout")
		log.Printf("[INFO] Updating timeout for deployment %s from %d to %d", deploymentID, old, new)
		err := client.APIClient.UpdateDeploymentTimeout(ctx, deploymentID, new.(int))
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to update timeout for deployment %s: %w", deploymentID, err))
		}
	}

	// If strategy or schedule change, it would typically require a new deployment.
	// For simplicity in this MVP, we assume these also trigger a replacement.

	return resourceNosanaJobRead(ctx, d, m)
}

// resourceNosanaJobDelete handles the deletion (archiving) of a Nosana job.
func resourceNosanaJobDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*nosanaClient)
	deploymentID := d.Id()

	log.Printf("[INFO] Archiving Nosana deployment %s...", deploymentID)
	err := client.APIClient.DeleteDeployment(ctx, deploymentID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to archive Nosana deployment %s: %w", deploymentID, err))
	}

	d.SetId("") // Mark resource as deleted from Terraform state

	log.Printf("[INFO] Nosana deployment %s archived successfully.", deploymentID)
	return nil
}
