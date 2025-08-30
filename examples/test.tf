# Simple Test Configuration for Nosana Provider
# Use this to verify your provider is working correctly

terraform {
  required_providers {
    nosana = {
      source  = "registry.terraform.io/hoomandigital/nosana"
      version = "~> 0.1"
    }
  }
}

provider "nosana" {
  keypair_path   = var.keypair_path
  network        = var.network
  market_address = var.market_address
}

variable "keypair_path" {
  description = "Path to your Solana keypair"
  type        = string
  default     = "~/.config/solana/id.json"
}

variable "network" {
  description = "Solana network"
  type        = string
  default     = "devnet"  # Use devnet for testing
}

variable "market_address" {
  description = "Nosana market address"
  type        = string
  default     = "nosScmHY2uR24Zh751PmGj9ww9QRNHewh9H59AfrT"
}

# Simple test job - just run a basic command
resource "nosana_job" "test_job" {
  job_definition = jsonencode({
    "ops": [
      {
        "id": "test",
        "args": {
          "env": {
            "MESSAGE": "Hello from Nosana!"
          },
          "image": "alpine:latest",
          "command": ["echo", "$MESSAGE"]
        },
        "type": "container/run"
      }
    ],
    "type": "container",
    "version": "0.1"
  })

  wait_for_completion = true
  completion_timeout_seconds = 300

  tags = {
    "test" = "true"
  }
}

# Output the test results
output "test_job_id" {
  description = "Test job ID"
  value       = nosana_job.test_job.id
}

output "test_job_status" {
  description = "Test job status"
  value       = nosana_job.test_job.status
}

output "test_job_result" {
  description = "Test job result"
  value       = nosana_job.test_job.result
}