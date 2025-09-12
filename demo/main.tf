terraform {
  required_providers {
    nosana = {
      source = "localhost/hoomandigital/nosana"
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
    "ops": [
      {
        "id": "Hello-world",
        "args": {
          "cmd": "echo hello world",
          "gpu": true,
          "image": "ubuntu"
        },
        "type": "container/run"
      }
    ],
    "meta": {
      "trigger": "dashboard"
    },
    "type": "container",
    "version": "0.1"
  })

  replicas                   = 1
  strategy                   = "SIMPLE"
  timeout                    = 600
  wait_for_completion        = false 
  max_retries                = 8
}

output "job_id" {
  value = nosana_job.ollama_server.id
}

output "job_status" {
  value = nosana_job.ollama_server.status
}
