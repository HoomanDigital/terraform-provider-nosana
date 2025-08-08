# Terraform Provider for Nosana

A Terraform provider for managing Nosana jobs on the Nosana Network.

## 🚀 Quick Start

**📖 For detailed setup instructions, see [SETUP.md](SETUP.md)**

## Prerequisites

- **Go 1.21+**, **Terraform 1.0+**, **Node.js/npm**
- **Nosana CLI**: `npm install -g @nosana/cli`
- **Funded Nosana wallet** with SOL and NOS tokens

### Quick Setup
```bash
# Clone and setup
git clone https://github.com/hoomandigital/terraform-provider-nosana.git
cd terraform-provider-nosana

# Windows: .\dev.ps1 dev
# Linux/macOS: ./dev.sh dev
```

## Development Commands

### Windows Commands

| Command | Description |
|---------|-------------|
| `.\dev.ps1 build` | Build the provider binary |
| `.\dev.ps1 clean` | Remove build artifacts |
| `.\dev.ps1 install` | Build and install provider locally |
| `.\dev.ps1 init` | Initialize Terraform |
| `.\dev.ps1 plan` | Run terraform plan |
| `.\dev.ps1 apply` | Run terraform apply |
| `.\dev.ps1 destroy` | Run terraform destroy |
| `.\dev.ps1 test` | Run Go tests |
| `.\dev.ps1 fmt` | Format Go code |
| `.\dev.ps1 vet` | Run Go vet |
| `.\dev.ps1 dev` | Full development cycle |
| `.\dev.ps1 help` | Show all commands |

### Linux/macOS Commands

| Command | Description |
|---------|-------------|
| `./dev.sh build` | Build the provider binary |
| `./dev.sh clean` | Remove build artifacts |
| `./dev.sh install` | Build and install provider locally |
| `./dev.sh init` | Initialize Terraform |
| `./dev.sh plan` | Run terraform plan |
| `./dev.sh apply` | Run terraform apply |
| `./dev.sh destroy` | Run terraform destroy |
| `./dev.sh test` | Run Go tests |
| `./dev.sh fmt` | Format Go code |
| `./dev.sh vet` | Run Go vet |
| `./dev.sh dev` | Full development cycle |
| `./dev.sh help` | Show all commands |

## 📋 What It Does

Deploy **AI/ML workloads**, **web services**, and **containerized applications** on the Nosana decentralized compute network using familiar Terraform workflows.

**Key Features:**
- 🤖 **GPU-enabled AI workloads** (LLMs, ML inference, training)
- 🌐 **Web services** with automatic port exposure  
- 💰 **Cost-effective** - Pay only for compute time used
- 🔒 **Decentralized** - No single point of failure
- 📊 **Terraform state management** - Full lifecycle support

## Example Usage

```hcl
terraform {
  required_providers {
    nosana = {
      source = "localhost/hoomandigital/nosana"
    }
  }
}

provider "nosana" {
  keypair_path   = var.keypair_path
  network        = var.network
  market_address = var.market_address
}

resource "nosana_job" "example" {
  job_definition = jsonencode({
    "ops": [
      {
        "id": "oneClickLLM",
        "args": {
          "env": {
            "MODEL_NAME": "mistral",
            "PORT": "8000"
          },
          "image": "docker.io/hoomanhq/oneclickllm:01",
          "expose": 8000
        },
        "type": "container/run"
      }
    ],
    "type": "container",
    "version": "0.1"
  })

  wait_for_completion = false
  completion_timeout_seconds = 600
}
```



## 🏗️ Project Structure

```
terraform-provider-nosana/
├── nosana/              # Provider source code
│   ├── provider.go      # Provider configuration
│   └── resource_nosana_job.go # Job resource implementation  
├── examples/            # Usage examples
│   ├── README.md        # Example documentation
│   └── main.tf          # Working configurations
├── SETUP.md             # Detailed setup guide
├── DEV_GUIDE.md         # Development guide
├── dev.ps1 / dev.sh     # Development scripts
└── main.go              # Entry point
```

## 🔧 Development

See [DEV_GUIDE.md](DEV_GUIDE.md) for development instructions.
