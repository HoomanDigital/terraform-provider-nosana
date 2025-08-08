# Nosana Terraform Provider Examples

This directory contains example configurations demonstrating how to use the Nosana Terraform Provider.

## Basic Example

The `main.tf` file shows a basic example of deploying a containerized AI workload (Mistral LLM) on the Nosana Network.

### Prerequisites

1. Install the Nosana CLI: `npm install -g @nosana/cli`
2. Configure your Nosana wallet: `nosana address`
3. Ensure you have SOL and NOS tokens in your wallet

### Usage

1. Initialize Terraform: `terraform init`
2. Plan the deployment: `terraform plan`
3. Apply the configuration: `terraform apply`

### Configuration

The example deploys:
- **Model**: Mistral 34B LLM
- **Container**: `docker.io/hoomanhq/oneclickllm:stabllama`
- **GPU**: Enabled with 0.9 memory utilization
- **Port**: 8000 for API access

### Outputs

- `job_id`: The unique identifier for the deployed job
- `job_status`: Current status of the job (PENDING, RUNNING, COMPLETED, etc.)