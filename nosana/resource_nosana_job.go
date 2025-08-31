// resource_nosana_job.go  
package nosana

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// JobPostResult represents the result from the Node.js job post script
type JobPostResult struct {
	Success        bool   `json:"success"`
	JobID          string `json:"job_id"`
	TransactionID  string `json:"transaction_id"`
	IPFSHash       string `json:"ipfs_hash"`
	MarketAddress  string `json:"market_address"`
	DashboardURL   string `json:"dashboard_url"`
	MarketURL      string `json:"market_url"`
	Error          string `json:"error"`
}

// JobStatusResult represents the result from the Node.js job status script
type JobStatusResult struct {
	Success      bool        `json:"success"`
	JobID        string      `json:"job_id"`
	Status       string      `json:"status"`
	State        string      `json:"state"`
	IPFSJob      string      `json:"ipfs_job"`
	IPFSResult   string      `json:"ipfs_result"`
	Market       string      `json:"market"`
	Node         string      `json:"node"`
	Price        string      `json:"price"`
	TimeStart    string      `json:"time_start"`
	TimeEnd      string      `json:"time_end"`
	DashboardURL string      `json:"dashboard_url"`
	ResultData   interface{} `json:"result_data"`
	ResultError  string      `json:"result_error"`
	Error        string      `json:"error"`
}

// resourceNosanaJob defines the schema and CRUD operations for the nosana_job resource.
func resourceNosanaJob() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNosanaJobCreate,
		ReadContext:   resourceNosanaJobRead,
		UpdateContext: resourceNosanaJobUpdate,
		DeleteContext: resourceNosanaJobDelete,
		Schema: map[string]*schema.Schema{
			"job_definition": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "JSON-encoded job specification for the Nosana job.",
				ValidateFunc: validation.StringIsJSON, // Ensure the string is valid JSON
			},
			"wait_for_completion": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "If true, Terraform will wait for the job to complete.",
			},
			"completion_timeout_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      300, // 5 minutes
				Description:  "Maximum time (in seconds) to wait for job completion.",
				ValidateFunc: validation.IntAtLeast(1),
			},
			// Attributes that will be stored in the Terraform state after creation
			"job_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The unique ID of the Nosana job.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The current status of the Nosana job.",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext, // For MVP, basic import functionality
		},
	}
}

// --- Placeholder for Nosana API interactions ---

// NosanaJob represents the structure of a Nosana job as returned by the API.
type NosanaJob struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	// Add other relevant fields from the Nosana API response
}

// createNosanaJobAPI submits a job to Nosana using the SDK.
func (c *nosanaClient) createNosanaJobAPI(jobDefinition, marketAddress string) (*NosanaJob, error) {
	log.Printf("[INFO] Nosana SDK: Creating job")

	// Determine market address - use job-specific if provided, otherwise provider default
	market := marketAddress
	if market == "" {
		market = c.MarketAddress
	}

	// Validate job definition is valid JSON
	var jobData interface{}
	if err := json.Unmarshal([]byte(jobDefinition), &jobData); err != nil {
		return nil, fmt.Errorf("invalid JSON in job definition: %w", err)
	}

	// Call Node.js script to post job using SDK
	// Args: market_address, job_definition_json
	output, err := c.runNodeJSScript("nosana-job-post.js", market, jobDefinition)
	if err != nil {
		return nil, fmt.Errorf("failed to submit job via SDK: %w", err)
	}

	// Parse the script output to extract job result
	result, err := parseJobPostOutput(output)
	if err != nil {
		return nil, fmt.Errorf("failed to parse job post result: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("job submission failed: %s", result.Error)
	}

	log.Printf("[INFO] Nosana SDK: Job created with ID: %s", result.JobID)
	return &NosanaJob{ID: result.JobID, Status: "PENDING"}, nil
}

// createTempJobFile creates a temporary JSON file with the job definition
func createTempJobFile(jobDefinition string) (string, error) {
	// Validate that it's valid JSON
	var jobData interface{}
	if err := json.Unmarshal([]byte(jobDefinition), &jobData); err != nil {
		return "", fmt.Errorf("invalid JSON in job definition: %w", err)
	}

	// Create temporary file
	tempFile, err := os.CreateTemp("", "nosana-job-*.json")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tempFile.Close()

	// Write job definition to file
	if _, err := tempFile.WriteString(jobDefinition); err != nil {
		os.Remove(tempFile.Name())
		return "", fmt.Errorf("failed to write job definition to temp file: %w", err)
	}

	log.Printf("[DEBUG] Created temporary job file: %s", tempFile.Name())
	return tempFile.Name(), nil
}

// parseJobPostOutput parses the output from the Node.js job post script
func parseJobPostOutput(output string) (*JobPostResult, error) {
	// Look for JSON result marker
	resultMarker := "JOB_RESULT_JSON:"
	errorMarker := "JOB_ERROR_JSON:"
	
	var jsonData string
	if idx := strings.Index(output, resultMarker); idx != -1 {
		jsonData = strings.TrimSpace(output[idx+len(resultMarker):])
	} else if idx := strings.Index(output, errorMarker); idx != -1 {
		jsonData = strings.TrimSpace(output[idx+len(errorMarker):])
	} else {
		return nil, fmt.Errorf("no JSON result found in script output: %s", output)
	}

	var result JobPostResult
	if err := json.Unmarshal([]byte(jsonData), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON result: %w", err)
	}

	return &result, nil
}

// parseJobStatusOutput parses the output from the Node.js job status script
func parseJobStatusOutput(output string) (*JobStatusResult, error) {
	// Look for JSON status marker
	statusMarker := "JOB_STATUS_JSON:"
	errorMarker := "JOB_ERROR_JSON:"
	
	var jsonData string
	if idx := strings.Index(output, statusMarker); idx != -1 {
		jsonData = strings.TrimSpace(output[idx+len(statusMarker):])
	} else if idx := strings.Index(output, errorMarker); idx != -1 {
		jsonData = strings.TrimSpace(output[idx+len(errorMarker):])
	} else {
		return nil, fmt.Errorf("no JSON status found in script output: %s", output)
	}

	var result JobStatusResult
	if err := json.Unmarshal([]byte(jsonData), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON status: %w", err)
	}

	return &result, nil
}

// extractJobIDFromOutput parses the Nosana CLI output to extract the job ID
func extractJobIDFromOutput(output string) string {
	log.Printf("[DEBUG] Parsing CLI output for job ID: %s", output)

	// Try to parse as JSON first (when using --format json)
	// Look for JSON content in the output (might be mixed with other text)
	jsonStart := strings.Index(output, "{")
	jsonEnd := strings.LastIndex(output, "}")
	var jsonResponse map[string]interface{}

	if jsonStart != -1 && jsonEnd != -1 && jsonEnd > jsonStart {
		jsonContent := output[jsonStart : jsonEnd+1]
		if err := json.Unmarshal([]byte(jsonContent), &jsonResponse); err == nil {
			// Look for job ID in JSON response
			if jobID, ok := jsonResponse["id"].(string); ok && jobID != "" {
				log.Printf("[DEBUG] Extracted job ID from JSON response: %s", jobID)
				return jobID
			}
			// Alternative field names in JSON response
			if jobID, ok := jsonResponse["job"].(string); ok && jobID != "" {
				log.Printf("[DEBUG] Extracted job ID from JSON 'job' field: %s", jobID)
				return jobID
			}
			if jobID, ok := jsonResponse["jobId"].(string); ok && jobID != "" {
				log.Printf("[DEBUG] Extracted job ID from JSON 'jobId' field: %s", jobID)
				return jobID
			}
			if txID, ok := jsonResponse["transaction"].(string); ok && txID != "" {
				log.Printf("[DEBUG] Using transaction ID as job ID from JSON: %s", txID)
				return txID
			}
			// Look for transaction_id in job_posting nested structure
			if jobPosting, ok := jsonResponse["job_posting"].(map[string]interface{}); ok {
				if txID, ok := jobPosting["transaction_id"].(string); ok && txID != "" {
					log.Printf("[DEBUG] Using transaction_id as job ID from job_posting: %s", txID)
					return txID
				}
			}
		}
	}

	// Fallback to text parsing for non-JSON output
	// Look for job ID in patterns like:
	// "Job: https://dashboard.nosana.com/jobs/FQTP2F5hNP2rNGUtQm4Annrx462PgxPcSA6ND6ToPTxH"
	// Or "Job posted: <job_id>"
	jobURLRegex := regexp.MustCompile(`Job:\s+https://dashboard\.nosana\.com/jobs/([A-Za-z0-9]+)`)
	matches := jobURLRegex.FindStringSubmatch(output)
	if len(matches) > 1 {
		log.Printf("[DEBUG] Extracted job ID from URL: %s", matches[1])
		return matches[1]
	}

	// Look for "Job posted: <job_id>" pattern
	jobPostedRegex := regexp.MustCompile(`Job posted:\s+([A-Za-z0-9]+)`)
	matches = jobPostedRegex.FindStringSubmatch(output)
	if len(matches) > 1 {
		log.Printf("[DEBUG] Extracted job ID from 'Job posted' message: %s", matches[1])
		return matches[1]
	}

	// Look for transaction hash pattern:
	// "job posted with tx 2r75ajjHdr5mPZV85NjFxtY28tKYK3UvNtdD7W7TfYCKvCXGgEdgJsia3jWdWaz5VES5sZWipEabnjwQkoE1dcwf!"
	txRegex := regexp.MustCompile(`job posted with tx ([A-Za-z0-9]+)`)
	matches = txRegex.FindStringSubmatch(output)
	if len(matches) > 1 {
		log.Printf("[DEBUG] Extracted transaction hash as job ID: %s", matches[1])
		return matches[1]
	}

	// Look for any base58-like string that could be a job ID (32-44 characters)
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		words := strings.Fields(line)
		for _, word := range words {
			word = strings.TrimRight(word, "!.,")
			if len(word) >= 32 && len(word) <= 50 && isBase58Like(word) {
				log.Printf("[DEBUG] Found potential job ID: %s", word)
				return word
			}
		}
	}

	log.Printf("[WARN] Could not extract job ID from CLI output")
	return ""
}

// isBase58Like checks if a string looks like base58 encoding
func isBase58Like(s string) bool {
	base58Regex := regexp.MustCompile(`^[123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz]+$`)
	return base58Regex.MatchString(s)
}

// getNosanaJobStatusAPI gets job status using the SDK.
func (c *nosanaClient) getNosanaJobStatusAPI(jobID string) (*NosanaJob, error) {
	log.Printf("[INFO] Nosana SDK: Getting status for job ID: %s", jobID)

	// Call Node.js script to get job status using SDK
	output, err := c.runNodeJSScript("nosana-job-get.js", jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get job status via SDK: %w", err)
	}

	// Parse the script output to extract job status
	result, err := parseJobStatusOutput(output)
	if err != nil {
		return nil, fmt.Errorf("failed to parse job status result: %w", err)
	}

	if !result.Success {
		// Check if it's a "job not found" error
		if strings.Contains(result.Error, "not found") || strings.Contains(result.Error, "Job not found") {
			return nil, fmt.Errorf("job not found")
		}
		return nil, fmt.Errorf("job status check failed: %s", result.Error)
	}

	return &NosanaJob{ID: result.JobID, Status: result.Status}, nil
}

// deleteNosanaJobAPI simulates calling the Nosana API to delete a job.
func (c *nosanaClient) deleteNosanaJobAPI(jobID string) error {
	log.Printf("[INFO] Nosana API: Deleting job ID: %s", jobID)
	// In a real scenario, this would make an HTTP DELETE request to Nosana's job deletion endpoint.
	// Example:
	// req, _ := http.NewRequest("DELETE", c.BaseURL + "/jobs/" + jobID, nil)
	// resp, err := http.DefaultClient.Do(req)
	// ... check response status
	log.Printf("[INFO] Nosana API: Job ID %s deleted successfully.", jobID)
	return nil
}

// --- End Placeholder ---

// resourceNosanaJobCreate handles the creation of a Nosana job.
func resourceNosanaJobCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*nosanaClient)
	jobDefinition := d.Get("job_definition").(string)
	waitForCompletion := d.Get("wait_for_completion").(bool)
	completionTimeoutSeconds := d.Get("completion_timeout_seconds").(int)
	marketAddress := client.MarketAddress

	// 1. Create the Nosana Job via CLI
	job, err := client.createNosanaJobAPI(jobDefinition, marketAddress)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to create Nosana job: %w", err))
	}

	d.SetId(job.ID) // Set the Terraform resource ID to the Nosana job ID
	d.Set("job_id", job.ID)
	d.Set("status", job.Status)

	log.Printf("[INFO] Nosana job %s created. Initial status: %s", job.ID, job.Status)

	// 2. Optional: Wait for job completion
	if waitForCompletion {
		log.Printf("[INFO] Waiting for Nosana job %s to complete (timeout: %d seconds)...", job.ID, completionTimeoutSeconds)
		timeout := time.After(time.Duration(completionTimeoutSeconds) * time.Second)
		tick := time.NewTicker(5 * time.Second) // Poll every 5 seconds
		defer tick.Stop()

		for {
			select {
			case <-timeout:
				return diag.Errorf("timeout waiting for Nosana job %s to complete after %d seconds", job.ID, completionTimeoutSeconds)
			case <-tick.C:
				currentJob, err := client.getNosanaJobStatusAPI(job.ID)
				if err != nil {
					log.Printf("[WARN] Error polling job status for %s: %v", job.ID, err)
					// Continue polling, but log the error
					continue
				}

				d.Set("status", currentJob.Status) // Update status in state

				if currentJob.Status == "COMPLETED" {
					log.Printf("[INFO] Nosana job %s completed successfully.", job.ID)
					return resourceNosanaJobRead(ctx, d, m) // Read to ensure state is up-to-date
				} else if currentJob.Status == "FAILED" || currentJob.Status == "CANCELLED" {
					return diag.Errorf("Nosana job %s failed or was cancelled with status: %s", job.ID, currentJob.Status)
				}
				log.Printf("[INFO] Nosana job %s current status: %s", job.ID, currentJob.Status)
			}
		}
	}

	return resourceNosanaJobRead(ctx, d, m)
}

// resourceNosanaJobRead handles reading the state of a Nosana job.
func resourceNosanaJobRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*nosanaClient)
	jobID := d.Id() // Get the job ID from the Terraform state

	// 1. Get job status from Nosana API
	job, err := client.getNosanaJobStatusAPI(jobID)
	if err != nil {
		// If the job is not found, it means it has been deleted outside Terraform.
		// Invalidate the resource from state.
		if err.Error() == "job not found" { // Customize this check based on actual API error
			log.Printf("[WARN] Nosana job %s not found, removing from state.", jobID)
			d.SetId("") // Mark resource for deletion from state
			return nil
		}
		return diag.FromErr(fmt.Errorf("failed to read Nosana job %s: %w", jobID, err))
	}

	// 2. Update Terraform state with current job details
	d.Set("job_id", job.ID)
	d.Set("status", job.Status)
	// You might also want to re-set job_definition if the API returns it,
	// to detect drift, but for simple jobs, it's often assumed immutable.

	log.Printf("[INFO] Read Nosana job %s. Status: %s", job.ID, job.Status)
	return nil
}

// resourceNosanaJobUpdate handles updates to a Nosana job.
// For many job-based systems, jobs are immutable. An "update" might mean
// deleting the old job and creating a new one. For this MVP, we'll assume
// that changes to `job_definition` will trigger a recreation, and other
// fields like `wait_for_completion` are handled by Terraform's diffing.
func resourceNosanaJobUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// If job_definition changes, Terraform will automatically plan a destroy and then create.
	// For other fields like wait_for_completion, Terraform handles it locally.
	// If Nosana API supports in-place updates, implement them here.
	log.Printf("[INFO] Nosana job %s update called. No direct in-place update implemented for this MVP.", d.Id())
	return resourceNosanaJobRead(ctx, d, m) // Just re-read to ensure state is consistent
}

// resourceNosanaJobDelete handles the deletion of a Nosana job.
func resourceNosanaJobDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*nosanaClient)
	jobID := d.Id()

	// 1. Delete the Nosana Job via API
	err := client.deleteNosanaJobAPI(jobID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to delete Nosana job %s: %w", jobID, err))
	}

	d.SetId("") // Mark resource as deleted from Terraform state

	log.Printf("[INFO] Nosana job %s deleted successfully.", jobID)
	return nil
}
