# Nosana Terraform Provider Setup Guide

This guide walks you through setting up and using the Nosana Terraform Provider to deploy containerized workloads on the Nosana decentralized compute network.

## Prerequisites

Before you begin, ensure you have the following installed:

- **Go 1.21+** - [Download Go](https://golang.org/dl/)
- **Terraform 1.0+** - [Download Terraform](https://www.terraform.io/downloads.html)
- **Node.js/npm** - [Download Node.js](https://nodejs.org/) (for Nosana CLI)
- **Git** - For cloning the repository

## Quick Setup

### 1. Clone the Repository

```bash
git clone https://github.com/hoomandigital/terraform-provider-nosana.git
cd terraform-provider-nosana
```

### 2. Install Nosana CLI

```bash
npm install -g @nosana/cli
```

### 3. Setup Your Wallet

Initialize your Nosana wallet:
```bash
nosana address
```

This creates a wallet file at `~/.nosana/nosana_key.json`. If you don't have one, it will guide you through creating a new wallet.

**Important**: Ensure your wallet has:
- **SOL tokens** for transaction fees (minimum ~0.01 SOL)
- **NOS tokens** for job payments (varies by job complexity)

You can fund your wallet through:
- [Nosana Dashboard](https://dashboard.nosana.com/)
- DEX platforms like Jupiter or Raydium
- Centralized exchanges that support NOS

### 4. Build and Install the Provider

**For Windows (PowerShell):**
```powershell
.\dev.ps1 dev
```

**For Linux/macOS (Bash):**
```bash
chmod +x dev.sh
./dev.sh dev
```

This will:
- Build the provider binary
- Install it locally for Terraform
- Initialize a test environment

## Using the Provider in Your Own Configurations

### 1. Create a New Terraform Configuration

Create a new directory for your project:
```bash
mkdir my-nosana-project
cd my-nosana-project
```

### 2. Create Your Configuration File

Create a `main.tf` file with your configuration:

```hcl
terraform {
  required_providers {
    nosana = {
      source = "localhost/hoomandigital/nosana"
    }
  }
}

provider "nosana" {
  network        = "mainnet"  # or "devnet" for testing
  market_address = "7AtiXMSH6R1jjBxrcYjehCkkSF7zvYWte63gwEDBcGHq"
}

resource "nosana_job" "my_job" {
  job_definition = jsonencode({
    "ops": [
      {
        "id": "myContainer",
        "args": {
          "env": {
            "MY_VAR": "value"
          },
          "image": "nginx:latest",
          "expose": 80
        },
        "type": "container/run"
      }
    ],
    "meta": {
      "trigger": "terraform"
    },
    "type": "container",
    "version": "0.1"
  })

  wait_for_completion = false
  completion_timeout_seconds = 300
}

output "job_id" {
  value = nosana_job.my_job.id
}

output "job_status" {
  value = nosana_job.my_job.status
}
```

### 3. Deploy Your Configuration

```bash
# Initialize Terraform
terraform init

# Plan your deployment
terraform plan

# Apply your configuration
terraform apply
```

### 4. Monitor Your Job

After deployment, you can:
- View the job ID and status in Terraform outputs
- Visit the [Nosana Dashboard](https://dashboard.nosana.com/) to monitor your job
- Use the Nosana CLI: `nosana job get <job-id>`

## Configuration Options

### Provider Configuration

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| `network` | No | `mainnet` | Network to use (`mainnet` or `devnet`) |
| `market_address` | Yes | - | Market address for job submissions |
| `keypair_path` | No | `~/.nosana/nosana_key.json` | Path to your wallet keypair |
| `private_key` | No | - | Base58 private key (alternative to keypair_path) |

### Resource Configuration

The `nosana_job` resource accepts:

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| `job_definition` | Yes | - | JSON-encoded job specification |
| `wait_for_completion` | No | `false` | Wait for job completion |
| `completion_timeout_seconds` | No | `300` | Timeout for job completion |

### Environment Variables

You can use environment variables instead of provider configuration:

```bash
export NOSANA_PRIVATE_KEY="your_base58_private_key"
export NOSANA_KEYPAIR_PATH="/path/to/your/keypair.json"
export TF_VAR_network="mainnet"
export TF_VAR_market_address="7AtiXMSH6R1jjBxrcYjehCkkSF7zvYWte63gwEDBcGHq"
```

## Example Job Configurations

### Web Server
```hcl
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
}
```

### AI/ML Workload (GPU)
```hcl
resource "nosana_job" "ai_model" {
  job_definition = jsonencode({
    "ops": [
      {
        "id": "ai_inference",
        "args": {
          "image": "docker.io/hoomanhq/oneclickllm:stabllama",
          "gpu": true,
          "expose": 8000,
          "env": {
            "MODEL_NAME": "mistral",
            "GPU_MEMORY_UTILIZATION": "0.9"
          }
        },
        "type": "container/run"
      }
    ],
    "type": "container",
    "version": "0.1"
  })
}
```

### Background Job
```hcl
resource "nosana_job" "data_processing" {
  job_definition = jsonencode({
    "ops": [
      {
        "id": "processor",
        "args": {
          "image": "python:3.9-slim",
          "cmd": ["python", "-c", "print('Processing data...')"]
        },
        "type": "container/run"
      }
    ],
    "type": "container",
    "version": "0.1"
  })
  
  wait_for_completion = true
  completion_timeout_seconds = 600
}
```

## Development Commands

### Windows (PowerShell)
- `.\dev.ps1 build` - Build the provider
- `.\dev.ps1 clean` - Clean build artifacts  
- `.\dev.ps1 install` - Install provider locally
- `.\dev.ps1 test` - Run tests
- `.\dev.ps1 plan` - Test with terraform plan
- `.\dev.ps1 apply` - Test with terraform apply

### Linux/macOS (Bash)
- `./dev.sh build` - Build the provider
- `./dev.sh clean` - Clean build artifacts
- `./dev.sh install` - Install provider locally  
- `./dev.sh test` - Run tests
- `./dev.sh plan` - Test with terraform plan
- `./dev.sh apply` - Test with terraform apply

## Troubleshooting

### Common Issues

**Issue**: `nosana: command not found`
**Solution**: Install the Nosana CLI: `npm install -g @nosana/cli`

**Issue**: `failed to configure Nosana provider`
**Solution**: Ensure your wallet is set up: `nosana address`

**Issue**: Insufficient funds
**Solution**: Add SOL and NOS tokens to your wallet

**Issue**: `could not extract job ID`
**Solution**: Check if your job definition is valid JSON and your wallet has funds

### Getting Help

- Check the [examples](./examples/) directory for working configurations
- Visit the [Nosana Documentation](https://docs.nosana.io/)
- Join the [Nosana Discord](https://discord.gg/nosana) community

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.