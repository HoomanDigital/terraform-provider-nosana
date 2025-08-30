# Sample Terraform Configuration using Nosana Provider
# This demonstrates how to use your published Nosana provider

terraform {
  required_providers {
    nosana = {
      source  = "app.terraform.io/codebrewery/nosana"
      version = "~> 0.1"
    }
  }
}

# Configure the Nosana provider
provider "nosana" {
  # Your Solana wallet keypair path
  keypair_path = var.keypair_path

  # Network configuration (mainnet-beta, devnet, etc.)
  network = var.network

  # Nosana market address
  market_address = var.market_address
}

# Variables for configuration
variable "keypair_path" {
  description = "Path to your Solana keypair JSON file"
  type        = string
  default     = "~/.config/solana/id.json"
}

variable "network" {
  description = "Solana network to use"
  type        = string
  default     = "mainnet-beta"
}

variable "market_address" {
  description = "Nosana market program address"
  type        = string
  default     = "nosScmHY2uR24Zh751PmGj9ww9QRNHewh9H59AfrT"
}

# Example 1: Deploy a simple web service
resource "nosana_job" "web_service" {
  job_definition = jsonencode({
    "ops": [
      {
        "id": "web-server",
        "args": {
          "env": {
            "PORT": "8080"
          },
          "image": "nginx:alpine",
          "expose": 8080
        },
        "type": "container/run"
      }
    ],
    "type": "container",
    "version": "0.1"
  })

  wait_for_completion = false
  completion_timeout_seconds = 600

  tags = {
    "environment" = "demo"
    "service"     = "web"
  }
}

# Example 2: Deploy an AI/ML workload (GPU-enabled)
resource "nosana_job" "ai_workload" {
  job_definition = jsonencode({
    "ops": [
      {
        "id": "ml-training",
        "args": {
          "env": {
            "MODEL_NAME": "bert-base-uncased",
            "EPOCHS": "10",
            "BATCH_SIZE": "8"
          },
          "image": "pytorch/pytorch:1.12.1-cuda11.3-cudnn8-runtime",
          "gpu": true,
          "expose": 8888
        },
        "type": "container/run"
      }
    ],
    "type": "container",
    "version": "0.1"
  })

  wait_for_completion = true
  completion_timeout_seconds = 3600

  tags = {
    "environment" = "ml-training"
    "gpu"         = "required"
  }
}

# Example 3: Run a data processing job
resource "nosana_job" "data_processing" {
  job_definition = jsonencode({
    "ops": [
      {
        "id": "data-processor",
        "args": {
          "env": {
            "INPUT_FILE": "s3://my-bucket/input.csv",
            "OUTPUT_FILE": "s3://my-bucket/output.json"
          },
          "image": "python:3.9-slim",
          "command": ["python", "process_data.py"]
        },
        "type": "container/run"
      }
    ],
    "type": "container",
    "version": "0.1"
  })

  wait_for_completion = true
  completion_timeout_seconds = 1800

  tags = {
    "environment" = "data-processing"
    "input"       = "s3"
  }
}

# Outputs
output "web_service_job_id" {
  description = "Job ID of the web service deployment"
  value       = nosana_job.web_service.id
}

output "web_service_status" {
  description = "Current status of the web service job"
  value       = nosana_job.web_service.status
}

output "web_service_result" {
  description = "Result/output of the web service job"
  value       = nosana_job.web_service.result
}

output "ai_workload_job_id" {
  description = "Job ID of the AI/ML workload"
  value       = nosana_job.ai_workload.id
}

output "ai_workload_status" {
  description = "Current status of the AI workload"
  value       = nosana_job.ai_workload.status
}

output "data_processing_job_id" {
  description = "Job ID of the data processing job"
  value       = nosana_job.data_processing.id
}