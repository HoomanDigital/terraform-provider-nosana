terraform {
  required_providers {
    nosana = {
      source = "localhost/hoomandigital/nosana"
    }
  }
}

# Variables for provider configuration
variable "keypair_path" {
  description = "Path to Nosana keypair file"
  type        = string
  default     = ""  # Uses default ~/.nosana/nosana_key.json
}

variable "network" {
  description = "The Nosana network to connect to"
  type        = string
  default     = "mainnet"
}

variable "market_address" {
  description = "Default market address for job submissions"
  type        = string
  default     = "7AtiXMSH6R1jjBxrcYjehCkkSF7zvYWte63gwEDBcGHq"  # Example market from docs
}

provider "nosana" {
  # 🚀 OUT-OF-THE-BOX CONFIGURATION
  # Just set environment variable: $env:NOSANA_PRIVATE_KEY = "your_phantom_wallet_private_key"
  # The provider automatically handles keypair setup!
  
  network        = var.network
  market_address = var.market_address
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
          "image": "docker.io/hoomanhq/oneclickllm:stabllama",
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
  completion_timeout_seconds = 60
}

output "job_id" {
  value = nosana_job.ollama_server.id
}

output "job_status" {
  value = nosana_job.ollama_server.status
}
