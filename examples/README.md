# Nosana Provider Examples

This directory contains example Terraform configurations demonstrating how to use your published Nosana provider.

## Prerequisites

Before using these examples, ensure you have:

1. **Nosana Provider Published**: Complete the publishing process to HCP Terraform
2. **Solana Wallet**: A funded wallet with SOL and NOS tokens
3. **Nosana CLI**: Install with `npm install -g @nosana/cli`

## Sample Configuration

### `sample.tf`

This file demonstrates three different types of workloads you can deploy on Nosana:

#### 1. Web Service Deployment
```hcl
resource "nosana_job" "web_service" {
  job_definition = jsonencode({
    "ops": [{
      "id": "web-server",
      "args": {
        "env": {"PORT": "8080"},
        "image": "nginx:alpine",
        "expose": 8080
      },
      "type": "container/run"
    }],
    "type": "container",
    "version": "0.1"
  })
}
```

#### 2. AI/ML Workload (GPU-enabled)
```hcl
resource "nosana_job" "ai_workload" {
  job_definition = jsonencode({
    "ops": [{
      "id": "ml-training",
      "args": {
        "env": {"MODEL_NAME": "bert-base-uncased"},
        "image": "pytorch/pytorch:1.12.1-cuda11.3-cudnn8-runtime",
        "gpu": true
      },
      "type": "container/run"
    }],
    "type": "container",
    "version": "0.1"
  })
}
```

#### 3. Data Processing Job
```hcl
resource "nosana_job" "data_processing" {
  job_definition = jsonencode({
    "ops": [{
      "id": "data-processor",
      "args": {
        "env": {
          "INPUT_FILE": "s3://my-bucket/input.csv",
          "OUTPUT_FILE": "s3://my-bucket/output.json"
        },
        "image": "python:3.9-slim"
      },
      "type": "container/run"
    }],
    "type": "container",
    "version": "0.1"
  })
}
```

## Usage Instructions

1. **Initialize Terraform**:
   ```bash
   terraform init
   ```

2. **Review the plan**:
   ```bash
   terraform plan
   ```

3. **Deploy the workloads**:
   ```bash
   terraform apply
   ```

4. **Monitor job status**:
   ```bash
   terraform output
   ```

5. **Clean up when done**:
   ```bash
   terraform destroy
   ```

## Configuration Options

### Provider Configuration
```hcl
provider "nosana" {
  keypair_path   = "~/.config/solana/id.json"  # Your Solana keypair
  network        = "mainnet-beta"             # Solana network
  market_address = "nosScmHY2uR24Zh751PmGj9ww9QRNHewh9H59AfrT"  # Market address
}
```

### Job Parameters
- `job_definition`: JSON specification of the workload
- `wait_for_completion`: Whether to wait for job completion (default: false)
- `completion_timeout_seconds`: Maximum time to wait (default: 600)
- `tags`: Key-value pairs for organization

## Important Notes

1. **Costs**: Jobs are billed based on actual compute time used
2. **GPU Access**: Set `"gpu": true` in your job definition for GPU workloads
3. **Port Exposure**: Use `"expose": PORT_NUMBER` to make services accessible
4. **Time Limits**: Long-running jobs may need extended timeouts
5. **Resource Cleanup**: Always run `terraform destroy` to stop running jobs

## Troubleshooting

### Common Issues:
- **Authentication Error**: Verify your Solana keypair path and funding
- **Timeout Error**: Increase `completion_timeout_seconds` for long jobs
- **GPU Not Available**: Not all Nosana nodes have GPUs
- **Port Conflicts**: Choose unique ports for multiple services

### Getting Help:
- Check job status: `terraform output JOB_NAME_status`
- View job logs: Check the Nosana dashboard or CLI
- Debug configuration: Use `terraform validate` and `terraform plan`