---
page_title: "Nosana Provider"
subcategory: ""
description: |-
  The Nosana provider is used to interact with the Nosana decentralized compute network.
---

# Nosana Provider

The Nosana provider is used to interact with the [Nosana](https://nosana.io) decentralized compute network. Nosana allows you to deploy containerized workloads across a distributed network of compute nodes, providing cost-effective GPU and CPU resources for AI/ML workloads, web services, and general compute tasks.

Use the navigation to the left to read about the available resources.

## Example Usage

```terraform
terraform {
  required_providers {
    nosana = {
      source = "HoomanDigital/nosana"
    }
  }
}

provider "nosana" {
  network        = "mainnet"
  market_address = "7AtiXMSH6R1jjBxrcYjehCkkSF7zvYWte63gwEDBcGHq"
}

resource "nosana_job" "example" {
  job_definition = jsonencode({
    "ops": [
      {
        "id": "webserver",
        "args": {
          "image": "nginx:alpine",
          "expose": 80
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

## Authentication

The Nosana provider requires authentication with the Nosana network using a Solana wallet. You can provide credentials in one of the following ways:

### Option 1: Private Key (Recommended for CI/CD)

```terraform
provider "nosana" {
  private_key    = var.nosana_private_key
  network        = "mainnet"
  market_address = "7AtiXMSH6R1jjBxrcYjehCkkSF7zvYWte63gwEDBcGHq"
}
```

Or set the environment variable:
```bash
export NOSANA_PRIVATE_KEY="your_base58_private_key"
```

### Option 2: Keypair File

```terraform
provider "nosana" {
  keypair_path   = "~/.nosana/nosana_key.json"
  network        = "mainnet"
  market_address = "7AtiXMSH6R1jjBxrcYjehCkkSF7zvYWte63gwEDBcGHq"
}
```

Or set the environment variable:
```bash
export NOSANA_KEYPAIR_PATH="/path/to/your/keypair.json"
```

### Prerequisites

Before using the provider, ensure you have:

1. **Nosana CLI** installed: `npm install -g @nosana/cli`
2. **Funded Solana wallet** with SOL and NOS tokens
3. **Go 1.21+** (for building from source)

### Wallet Setup

If you don't have a Nosana wallet, create one using:

```bash
nosana address
```

This will create a keypair file at `~/.nosana/nosana_key.json`. Make sure to fund your wallet with:
- **SOL tokens** for transaction fees (minimum ~0.01 SOL)
- **NOS tokens** for job payments (varies by job complexity)

## Schema

### Required

- `market_address` (String) The market address for job submissions on the Nosana network.

### Optional

- `network` (String) The Nosana network to connect to. Valid values are `mainnet` and `devnet`. Defaults to `mainnet`.
- `private_key` (String, Sensitive) Solana private key in base58 format. Can also be set via the `NOSANA_PRIVATE_KEY` environment variable.
- `keypair_path` (String) Path to Nosana keypair file. If not provided and no private_key is set, uses default `~/.nosana/nosana_key.json`. Can also be set via the `NOSANA_KEYPAIR_PATH` environment variable.

## Environment Variables

The following environment variables can be used as alternatives to provider configuration:

- `NOSANA_PRIVATE_KEY` - Solana private key in base58 format
- `NOSANA_KEYPAIR_PATH` - Path to Nosana keypair file

## Networks and Market Addresses

### Mainnet
- Network: `mainnet`
- Default Market Address: `7AtiXMSH6R1jjBxrcYjehCkkSF7zvYWte63gwEDBcGHq`

### Devnet (Testing)
- Network: `devnet`
- Market Address: Contact Nosana support for devnet market addresses

## Job Types

The Nosana network supports various types of containerized workloads:

### Web Services
Deploy web applications with automatic port exposure:

```hcl
resource "nosana_job" "web_app" {
  job_definition = jsonencode({
    "ops": [
      {
        "id": "webapp",
        "args": {
          "image": "nginx:alpine",
          "expose": 80,
          "env": {
            "ENV": "production"
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

### AI/ML Workloads
Deploy GPU-enabled AI models and machine learning workloads:

```hcl
resource "nosana_job" "ai_model" {
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

### Batch Processing
Run batch jobs for data processing:

```hcl
resource "nosana_job" "batch_job" {
  job_definition = jsonencode({
    "ops": [
      {
        "id": "processor",
        "args": {
          "image": "python:3.9-slim",
          "cmd": ["python", "process_data.py"]
        },
        "type": "container/run"
      }
    ],
    "type": "container",
    "version": "0.1"
  })
  
  wait_for_completion = true
  completion_timeout_seconds = 3600
}
```
