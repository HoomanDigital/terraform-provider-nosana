terraform {
  required_providers {
    nosana = {
      source  = "HoomanDigital/nosana"
      version = "~>0.1"
    }
  }
}

# Variables for provider configuration
variable "market_address" {
  description = "Default market address for job submissions"
  type        = string
  default     = "HanragNudL4S4zFtpLQv85dn6QbdzCm7SNEWEb9sRp17"  
}

provider "nosana" {
  # Use your specific private key
  private_key    = "5YeqfFZJfJf8JRPdUqCzNjUfJuMYc7KyxkTr63T8TgcBVwPkfKYWB7yG566v9jaMoFPvDrBLnZQenAfjRVtur5ob"
  market_address = var.market_address
}

resource "nosana_job" "ollama_server" {
  job_content = jsonencode({
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
            "GPU_MEMORY_UTILIZATION": "0.9",
            "AIOHTTP_CLIENT_TIMEOUT": "3600"
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
  replicas = 1
  timeout  = 300
  strategy = "SIMPLE"

  wait_for_completion        = false
  completion_timeout_seconds = 60
}

output "job_id" {
  value = nosana_job.ollama_server.id
}

output "job_status" {
  value = nosana_job.ollama_server.status
}
