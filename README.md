# Terraform Provider for Nosana

A Terraform provider for managing Nosana jobs on the Nosana Network.

## Prerequisites

- **Go 1.21+** - [Download Go](https://golang.org/dl/)
- **Terraform 1.0+** - [Download Terraform](https://www.terraform.io/downloads.html)
- **Node.js/npm** - [Download Node.js](https://nodejs.org/) (for Nosana CLI)
- **Nosana CLI** - Install with: `npm install -g @nosana/cli`
- **Git** - For version control

## Quick Start

### Windows (PowerShell)

```powershell
# Clone and build
git clone <repository-url>
cd TerraformProvider-Nosana

# Development cycle (build, install, initialize)
.\dev.ps1 dev

# Test the provider
.\dev.ps1 plan
.\dev.ps1 apply
.\dev.ps1 destroy
```

### Linux/macOS (Bash)

```bash
# Clone and build
git clone <repository-url>
cd TerraformProvider-Nosana

# Make script executable
chmod +x dev.sh

# Development cycle (build, install, initialize)
./dev.sh dev

# Test the provider
./dev.sh plan
./dev.sh apply
./dev.sh destroy
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

## Configuration

### Setup Nosana CLI
First, install and configure the Nosana CLI:

```bash
# Install the Nosana CLI
npm install -g @nosana/cli

# Initialize your wallet (creates ~/.nosana/nosana_key.json)
nosana address

# Fund your wallet with SOL and NOS tokens
# Visit https://dashboard.nosana.com/ for more info
```

### Using with Default Settings
The provider works with the default Nosana CLI configuration:

**Windows:**
```powershell
.\dev.ps1 apply
```

**Linux/macOS:**
```bash
./dev.sh apply
```

### Using Custom Configuration
Set environment variables to customize the configuration:

**Windows:**
```powershell
$env:TF_VAR_keypair_path = "C:\path\to\your\keypair.json"  # Optional
$env:TF_VAR_network = "mainnet"  # or "devnet"
$env:TF_VAR_market_address = "your_preferred_market_address"

.\dev.ps1 apply
```

**Linux/macOS:**
```bash
export TF_VAR_keypair_path="/path/to/your/keypair.json"  # Optional
export TF_VAR_network="mainnet"  # or "devnet"
export TF_VAR_market_address="your_preferred_market_address"

./dev.sh apply
```

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

## Project Structure

```
TerraformProvider-Nosana/
├── main.go                    # Entry point
├── provider.go               # Provider configuration
├── resource_nosana_job.go    # Job resource implementation
├── test-local.tf            # Test configuration
├── dev.ps1                  # Windows development script
├── dev.sh                   # Linux development script
├── go.mod                   # Go module definition
└── go.sum                   # Go module checksums
```

## Manual Build (Alternative)

If you prefer not to use the dev scripts:

**Windows:**
```powershell
# Build
go build -o terraform-provider-nosana.exe .

# Install
$pluginPath = "$env:APPDATA\terraform.d\plugins\localhost\hoomandigital\nosana\1.0.0\windows_amd64"
New-Item -ItemType Directory -Path $pluginPath -Force
Copy-Item terraform-provider-nosana.exe $pluginPath -Force

# Test
terraform init
terraform plan
terraform apply
```

**Linux/macOS:**
```bash
# Build
go build -o terraform-provider-nosana .
