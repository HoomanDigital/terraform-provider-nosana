terraform {
  required_providers {
    nosana = {
      source = "localhost/hoomandigital/nosana"
    }
  }
}

# 🚀 Out-of-the-box Nosana Provider Configuration
# Simply set environment variables and start using Nosana!

provider "nosana" {
  # Option 1: Use environment variables (recommended for production)
  # Set these in your terminal before running terraform:
  # 
  # PowerShell/Windows:
  #   $env:NOSANA_PRIVATE_KEY = "your_base58_private_key_here"
  #   $env:TF_VAR_market_address = "7AtiXMSH6R1jjBxrcYjehCkkSF7zvYWte63gwEDBcGHq"
  #
  # Bash/Linux/Mac:
  #   export NOSANA_PRIVATE_KEY="your_base58_private_key_here"
  #   export TF_VAR_market_address="7AtiXMSH6R1jjBxrcYjehCkkSF7zvYWte63gwEDBcGHq"
  
  # The provider will automatically:
  # 1. Read your private key from NOSANA_PRIVATE_KEY environment variable
  # 2. Convert it to the proper Nosana CLI format
  # 3. Set up the keypair file in ~/.nosana/nosana_key.json
  # 4. Validate your wallet has sufficient SOL/NOS tokens
  
  # Required: Market address for job submissions
  market_address = var.market_address
  
  # Optional: Network (defaults to mainnet)
  network = "mainnet"
  
  # Optional: If you prefer to use a specific keypair file instead of private key
  # keypair_path = "/path/to/your/keypair.json"
}

# Variables for flexibility
variable "market_address" {
  description = "Nosana market address for job submissions"
  type        = string
  default     = "7AtiXMSH6R1jjBxrcYjehCkkSF7zvYWte63gwEDBcGHq"  # RTX 3060 market
}

# Example: Deploy a simple Hello World container
resource "nosana_job" "hello_world" {
  job_definition = jsonencode({
    "ops": [
      {
        "id": "hello",
        "args": {
          "image": "ubuntu:latest",
          "gpu": false
        },
        "type": "container/run"
      }
    ],
    "meta": {
      "trigger": "terraform-hello-world"
    },
    "type": "container",
    "version": "0.1"
  })

  wait_for_completion = false
  completion_timeout_seconds = 300  # 5 minutes
}

# Example: Deploy an AI model with GPU support
resource "nosana_job" "ai_inference" {
  job_definition = jsonencode({
    "ops": [
      {
        "id": "llama",
        "args": {
          "env": {
            "MODEL_NAME": "llama2",
            "PORT": "8000"
          },
          "gpu": true,
          "image": "huggingface/transformers-pytorch-gpu:latest",
          "expose": 8000
        },
        "type": "container/run"
      }
    ],
    "meta": {
      "trigger": "terraform-ai-inference"
    },
    "type": "container",
    "version": "0.1"
  })

  wait_for_completion = false
  completion_timeout_seconds = 600  # 10 minutes
}

# Outputs
output "hello_world_job_id" {
  description = "Job ID for the hello world container"
  value       = nosana_job.hello_world.job_id
}

output "hello_world_status" {
  description = "Status of the hello world job"
  value       = nosana_job.hello_world.status
}

output "ai_inference_job_id" {
  description = "Job ID for the AI inference container"
  value       = nosana_job.ai_inference.job_id
}

output "ai_inference_status" {
  description = "Status of the AI inference job"
  value       = nosana_job.ai_inference.status
}
