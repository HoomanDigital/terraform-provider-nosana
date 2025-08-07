# Quick Test for Out-of-the-Box Nosana Provider
# 
# Prerequisites:
# 1. Install Nosana CLI: npm install -g @nosana/cli
# 2. Set your private key: $env:NOSANA_PRIVATE_KEY = "your_phantom_wallet_private_key"
# 3. Fund your wallet with SOL and NOS tokens
#
# Usage:
# terraform init && terraform apply

terraform {
  required_providers {
    nosana = {
      source = "localhost/hoomandigital/nosana"
    }
  }
}

provider "nosana" {
  # No configuration needed! 
  # Provider automatically uses NOSANA_PRIVATE_KEY environment variable
  network        = "mainnet"
  market_address = "7AtiXMSH6R1jjBxrcYjehCkkSF7zvYWte63gwEDBcGHq"
}

resource "nosana_job" "quick_test" {
  job_definition = jsonencode({
    "type": "container",
    "version": "0.1", 
    "ops": [{
      "id": "test",
      "type": "container/run",
      "args": {
        "image": "ubuntu:latest",
        "cmd": ["echo", "🎉 Nosana + Terraform = Success!"]
      }
    }]
  })
  
  wait_for_completion = true
  completion_timeout_seconds = 120
}

output "success" {
  value = "Job ${nosana_job.quick_test.job_id} completed with status: ${nosana_job.quick_test.status}"
}
