terraform {
  required_providers {
    nosana = {
      source = "hoomandigital.com/terraform/nosana"
    }
  }
}

variable "private_key" {
  description = "Your Solana wallet private key"
  type        = string
  sensitive   = true
  default     = ""
}

variable "market_address" {
  description = "Default market address for job submissions"
  type        = string
  default     = "7AtiXMSH6R1jjBxrcYjehCkkSF7zvYWte63gwEDBcGHq"
}

variable "rpc_url" {
  description = "Solana RPC URL for blockchain transactions (use a fast RPC for better reliability)"
  type        = string
  default     = ""
}

provider "nosana" {
  private_key    = var.private_key
  market_address = var.market_address
  rpc_url        = var.rpc_url
}

resource "nosana_job" "ollama_server" {
  job_content = jsonencode({
    "meta": {
      "trigger": "terraform"
    },
    "ops": [
      {
        "args": {
          "env": {
            "BLOCK_SIZE": "8",
            "ENABLE_STREAMING": "false",
            "GPU_MEMORY_UTILIZATION": "0.9",
            "MAX_MODEL_LEN": "NAN",
            "MEMORY_LIMIT": "NAN",
            "MODEL_NAME": "mistral",
            "PARAMETER_SIZE": "34B",
            "PORT": "8000",
            "QUANTIZATION": "NAN",
            "SERVED_MODEL_NAME": "mistral",
            "SWAP_SPACE": "4",
            "TENSOR_PARALLEL_SIZE": "1"
          },
          "expose": 8000,
          "gpu": true,
          "image": "docker.io/hoomanhq/oneclickllm:stabllama"
        },
        "id": "oneClickLLM",
        "type": "container/run"
      }
    ],
    "type": "container",
    "version": "0.1"
  })

  replicas                   = 1
  strategy                   = "SIMPLE"
  timeout                    = 1200  # Increased timeout for blockchain transactions
  wait_for_completion        = false # Quick deployment 
  max_retries                = 15      # Increased retries with aggressive restart strategy
}

output "job_id" {
  value = nosana_job.ollama_server.id
}

output "job_status" {
  value = nosana_job.ollama_server.status
}
