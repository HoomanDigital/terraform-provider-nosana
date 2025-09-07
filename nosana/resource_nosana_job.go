// resource_nosana_job.go  
package nosana

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

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
		"run_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The run ID from direct blockchain submission.",
		},
		"transaction_hash": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The blockchain transaction hash for the job submission.",
		},
		"deployment_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The deployment ID from the Nosana API (legacy).",
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
	timeout := d.Get("timeout").(int)
	marketAddress := client.MarketAddress

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

		// EXACT SDK PATTERN - createDeployment.ts + deploymentStart.ts
		log.Printf("[INFO] 🚀 Using EXACT SDK createDeployment + start pattern")
		
		// Step 1: Create deployment (DRAFT) - exact SDK createDeployment pattern
		log.Printf("[INFO] 📝 Creating deployment via SDK pattern...")
		createReq := map[string]interface{}{
			"name":                 fmt.Sprintf("terraform-job-%d", time.Now().Unix()),
			"market":               marketAddress,
			"replicas":             d.Get("replicas").(int),
			"timeout":              timeout,
			"strategy":             "SIMPLE",
			"ipfs_definition_hash": ipfsHash,
		}

		// Create deployment exactly like SDK does
		deploymentResponse, err := client.APIClient.CreateDeploymentSDK(ctx, createReq)
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to create deployment: %w", err))
		}

		deploymentID := deploymentResponse["id"].(string)
		log.Printf("[INFO] ✅ Deployment created: %s", deploymentID)

		// Step 2: FUND VAULT with EXACT MARKET PRICE
		vaultAddress := deploymentResponse["vault"].(string)
		log.Printf("[INFO] 🏦 Funding deployment vault with exact market price: %s", vaultAddress)
		if err := client.APIClient.FundVault(ctx, vaultAddress, marketAddress); err != nil {
			log.Printf("[WARN] Vault funding failed (proceeding anyway): %v", err)
			// Don't fail deployment if vault funding fails - let's see what happens
		} else {
			log.Printf("[INFO] ✅ Vault funding completed successfully")
		}

		// Step 3: Start deployment - exact SDK deploymentStart pattern  
		log.Printf("[INFO] 🏃 Starting deployment using SDK start pattern...")
		err = client.APIClient.StartDeploymentSDK(ctx, deploymentID)
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to start deployment: %w", err))
		}

		log.Printf("[INFO] ✅ Deployment start requested")

		// Set Terraform state
		d.SetId(deploymentID)
		d.Set("job_id", deploymentID)
		d.Set("status", "STARTING")  // deploymentStart sets to STARTING
		d.Set("ipfs_hash", ipfsHash)

		// Wait up to ~70s for RUNNING
		deadline := time.Now().Add(70 * time.Second)
		for time.Now().Before(deadline) {
			dep, gerr := client.APIClient.GetDeployment(ctx, deploymentID)
			if gerr == nil {
				if dep.Status == DeploymentStatusRunning {
					log.Printf("[INFO] ✅ Deployment is RUNNING")
					d.Set("status", string(DeploymentStatusRunning))
					break
				}
				if dep.Status == DeploymentStatusError {
					log.Printf("[WARN] Deployment moved to ERROR during wait")
					d.Set("status", string(DeploymentStatusError))
					break
				}
			}
			time.Sleep(5 * time.Second)
		}

		log.Printf("[INFO] 🆔 Deployment ID: %s", deploymentID)
		log.Printf("[INFO] 💰 Vault Address: %s", vaultAddress)
		log.Printf("[INFO] 🌐 Check dashboard: https://dashboard.nosana.com/account/deployer")

	return resourceNosanaJobRead(ctx, d, m)
}

// resourceNosanaJobRead handles reading the state of a Nosana job.
func resourceNosanaJobRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[INFO] Reading Nosana job state for job ID: %s", d.Id())

	client := m.(*nosanaClient)
	
	// Get current deployment status from Nosana API
	deployment, err := client.APIClient.GetDeployment(ctx, d.Id())
	if err != nil {
		// If deployment not found, resource has been deleted
		log.Printf("[WARN] Deployment %s not found, removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}

	// Update all deployment fields from API response
	d.Set("status", string(deployment.Status))
	d.Set("job_id", deployment.ID)
	
	log.Printf("[INFO] ✅ Deployment %s status updated: %s", d.Id(), deployment.Status)
	
	// If deployment failed, get detailed error information
	if deployment.Status == DeploymentStatusError {
		log.Printf("[ERROR] 💥 Deployment %s has ERROR status! Getting detailed error info...", d.Id())
		
		// Get detailed deployment with events
		details, err := client.APIClient.GetDeploymentWithEvents(ctx, d.Id())
		if err != nil {
			log.Printf("[WARN] Could not get deployment details: %v", err)
		} else {
			// Extract and log events for debugging
			if events, ok := details["events"].([]interface{}); ok && len(events) > 0 {
				log.Printf("[ERROR] 🔍 Deployment error events:")
				for i, event := range events {
					if eventMap, ok := event.(map[string]interface{}); ok {
						eventType := eventMap["type"]
						message := eventMap["message"]
						log.Printf("[ERROR]   Event %d: %s - %s", i+1, eventType, message)
					}
				}
			}
		}
	}
	
	return nil
}

// resourceNosanaJobUpdate handles updates to a Nosana job.
func resourceNosanaJobUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	jobID := d.Id()

	// For direct blockchain jobs, most updates require a new resource
	if d.HasChange("job_content") || d.HasChange("replicas") || d.HasChange("timeout") || d.HasChange("strategy") {
		return diag.Errorf("changes to direct blockchain jobs require creating a new resource - use 'terraform taint %s' to force replacement", jobID)
	}

	return resourceNosanaJobRead(ctx, d, m)
}

// resourceNosanaJobDelete handles the deletion of a Nosana job.
func resourceNosanaJobDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	jobID := d.Id()

	// For direct blockchain jobs, deletion only removes from Terraform state
	// The job on the blockchain network continues to run independently
	log.Printf("[INFO] Removing direct blockchain job %s from Terraform state...", jobID)
	log.Printf("[INFO] ⚠️  Note: Job continues running on Nosana network - this only removes from Terraform management")
	
	d.SetId("") // Mark resource as deleted from Terraform state

	log.Printf("[INFO] Job %s removed from Terraform state successfully.", jobID)
	return nil
}
