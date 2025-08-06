terraform {
  required_providers {
    nosana = {
      source = "localhost/hoomandigital/nosana"
    }
  }
}

# Variables for provider configuration
variable "wallet_address" {
  description = "The Solana wallet address for Nosana authentication"
  type        = string
  default     = "mock-wallet-address-for-testing"
}

variable "signed_challenge" {
  description = "A signed message from the Phantom wallet for authentication"
  type        = string
  sensitive   = true
  default     = "mock-signed-challenge-for-testing"
}

variable "network" {
  description = "The Nosana network to connect to"
  type        = string
  default     = "devnet"
}

provider "nosana" {
  # SETUP INSTRUCTIONS:
  # 1. Get your wallet address from Phantom wallet (44 character string)
  # 2. Sign a challenge message with your Phantom wallet
  # 3. Set environment variables in PowerShell:
  #    $env:TF_VAR_wallet_address = "your_wallet_address_from_phantom"
  #    $env:TF_VAR_signed_challenge = "your_signed_challenge_message"
  #
  # The provider will use these variables via Terraform variable substitution
  # DO NOT put real credentials directly in this file!
  
  wallet_address    = var.wallet_address
  signed_challenge  = var.signed_challenge
  network          = var.network
}

resource "nosana_job" "ollama_server" {
  job_definition = jsonencode({
    "ops": [
      {
        "id": "oneClickLLM",
        "args": {
          "env": {
            "PORT": "8000",
            "BLOCK_SIZE": "8",
            "MODEL_NAME": "mistral",
            "SWAP_SPACE": "4",
            "MEMORY_LIMIT": "NAN",
            "QUANTIZATION": "NAN",
            "MAX_MODEL_LEN": "NAN",
            "PARAMETER_SIZE": "34B",
            "ENABLE_STREAMING": "false",
            "SERVED_MODEL_NAME": "mistral",
            "TENSOR_PARALLEL_SIZE": "1",
            "GPU_MEMORY_UTILIZATION": "0.9"
          },
          "gpu": true,
          "image": "docker.io/hoomanhq/oneclickllm:01",
          "expose": 8000
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
  completion_timeout_seconds = 600
}

output "job_id" {
  value = nosana_job.ollama_server.id
}

output "job_status" {
  value = nosana_job.ollama_server.status
}
