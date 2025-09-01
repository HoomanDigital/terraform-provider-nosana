terraform {
  required_providers {
    nosana = {
      source  = "registry.terraform.io/HoomanDigital/nosana"
      # version = "~> 0.2.0"  # Commented out - using local dev build
    }
  }
}

# Variables for provider configuration
variable "private_key" {
  description = "Solana private key in base58 format"
  type        = string
  sensitive   = true
}

variable "network" {
  description = "The Nosana network to connect to"
  type        = string
  default     = "mainnet"
}

variable "market_address" {
  description = "Default market address for job submissions"
  type        = string
  default     = "7AtiXMSH6R1jjBxrcYjehCkkSF7zvYWte63gwEDBcGHq"  
}

provider "nosana" {
  # Testing the published provider with SDK integration
  private_key    = var.private_key
  network        = var.network
  market_address = var.market_address
}

# Simple "Hello World" job to test the published provider
resource "nosana_job" "hello_world_test" {
  job_definition = jsonencode({
    "ops": [
      {
        "type": "container/run",
        "id": "hello-world",
        "args": {
          "image": "ubuntu:latest",
          "cmd": ["echo", "Hello from Nosana SDK integration! 🚀"]
        }
      }
    ],
    "meta": {
      "trigger": "live-demo-test"
    },
    "type": "container",
    "version": "0.1"
  })

  wait_for_completion = false
  completion_timeout_seconds = 60
}

output "job_id" {
  value = nosana_job.hello_world_test.id
  description = "The ID of the submitted Nosana job"
}

output "job_status" {
  value = nosana_job.hello_world_test.status
  description = "The status of the submitted Nosana job"
}

output "dashboard_url" {
  value = "https://dashboard.nosana.com/jobs/${nosana_job.hello_world_test.id}"
  description = "Direct link to view the job on Nosana dashboard"
}