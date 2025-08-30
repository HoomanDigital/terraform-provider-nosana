---
page_title: "nosana_job Resource - terraform-provider-nosana"
subcategory: ""
description: |-
  Manages a Nosana job on the decentralized compute network.
---

# nosana_job (Resource)

Manages a Nosana job on the decentralized compute network. Nosana jobs are containerized workloads that run on distributed compute nodes, providing cost-effective access to GPU and CPU resources.

## Example Usage

### Basic Web Service

```terraform
resource "nosana_job" "web_server" {
  job_definition = jsonencode({
    "ops": [
      {
        "id": "webserver",
        "args": {
          "image": "nginx:alpine",
          "expose": 80,
          "env": {
            "NGINX_HOST": "localhost"
          }
        },
        "type": "container/run"
      }
    ],
    "type": "container",
    "version": "0.1"
  })

  wait_for_completion = false
  completion_timeout_seconds = 300
}
```

### GPU-Enabled AI Workload

```terraform
resource "nosana_job" "ai_inference" {
  job_definition = jsonencode({
    "ops": [
      {
        "id": "llm_server",
        "args": {
          "image": "docker.io/hoomanhq/oneclickllm:latest",
          "gpu": true,
          "expose": 8000,
          "env": {
            "MODEL_NAME": "mistral",
            "PARAMETER_SIZE": "7B",
            "GPU_MEMORY_UTILIZATION": "0.9",
            "PORT": "8000"
          }
        },
        "type": "container/run"
      }
    ],
    "type": "container",
    "version": "0.1"
  })

  wait_for_completion = false
  completion_timeout_seconds = 600
}
```

### Batch Processing Job

```terraform
resource "nosana_job" "data_processor" {
  job_definition = jsonencode({
    "ops": [
      {
        "id": "processor",
        "args": {
          "image": "python:3.9-slim",
          "cmd": ["python", "-c", "import time; print('Processing...'); time.sleep(30); print('Done!')"]
        },
        "type": "container/run"
      }
    ],
    "meta": {
      "trigger": "terraform",
      "name": "batch-processing-job"
    },
    "type": "container",
    "version": "0.1"
  })

  wait_for_completion = true
  completion_timeout_seconds = 120
}
```

## Schema

### Required

- `job_definition` (String) JSON-encoded job specification defining the containerized workload to run on Nosana network.

### Optional

- `completion_timeout_seconds` (Number) Maximum time to wait for job completion when `wait_for_completion` is true. Defaults to `300`.
- `wait_for_completion` (Boolean) Whether to wait for the job to complete before returning. Defaults to `false`.

### Read-Only

- `id` (String) The unique identifier of the job assigned by the Nosana network.
- `status` (String) Current status of the job (e.g., "pending", "running", "completed", "failed").

## Job Definition Schema

The `job_definition` field accepts a JSON-encoded object with the following structure:

### Root Level

- `type` (String, Required) - Must be "container"
- `version` (String, Required) - Job specification version, typically "0.1"
- `ops` (Array, Required) - Array of operations to execute
- `meta` (Object, Optional) - Metadata about the job

### Operations (`ops`)

Each operation in the `ops` array contains:

- `id` (String, Required) - Unique identifier for the operation
- `type` (String, Required) - Operation type, typically "container/run"
- `args` (Object, Required) - Arguments for the operation

### Operation Arguments (`args`)

- `image` (String, Required) - Docker image to run
- `cmd` (Array, Optional) - Command to execute in the container
- `env` (Object, Optional) - Environment variables as key-value pairs
- `expose` (Number, Optional) - Port to expose from the container
- `gpu` (Boolean, Optional) - Whether to request GPU access
- `memory` (String, Optional) - Memory limit (e.g., "1Gi", "512Mi")
- `cpu` (String, Optional) - CPU limit (e.g., "1000m", "0.5")

### Metadata (`meta`)

- `trigger` (String, Optional) - Source of the job (e.g., "terraform")
- `name` (String, Optional) - Human-readable name for the job
- `description` (String, Optional) - Description of the job

## Job Lifecycle

1. **Creation**: When the resource is created, a job is submitted to the Nosana network
2. **Execution**: The job is queued and assigned to an available compute node
3. **Monitoring**: Job status is tracked and updated
4. **Completion**: Job completes successfully or fails
5. **Destruction**: When the resource is destroyed, the job reference is removed (the actual job cannot be cancelled once submitted)

## Job Status Values

- `pending` - Job is queued and waiting for assignment
- `running` - Job is currently executing on a compute node
- `completed` - Job has finished successfully
- `failed` - Job encountered an error and failed
- `cancelled` - Job was cancelled (rare, as most jobs cannot be cancelled)

## Notes

- **Job Persistence**: Nosana jobs are ephemeral by design. Once submitted, they cannot be modified or cancelled through this provider
- **Cost Considerations**: Jobs consume NOS tokens based on compute time and resources used
- **Network Requirements**: Ensure your wallet has sufficient SOL (for transaction fees) and NOS (for job payments)
- **Container Images**: Images must be publicly accessible Docker images
- **GPU Availability**: GPU-enabled jobs may take longer to find available nodes
- **Port Exposure**: Exposed ports are accessible via the Nosana network's public endpoints

## Import

Nosana jobs can be imported using their job ID:

```terraform
terraform import nosana_job.example job-id-12345
```

Note: Imported jobs will have limited state information as the original job definition may not be fully recoverable from the Nosana network.
