# Terraform Provider for Nosana

A Terraform provider for managing Nosana jobs on the Nosana Network.

## Prerequisites

- **Go 1.21+** - [Download Go](https://golang.org/dl/)
- **Terraform 1.0+** - [Download Terraform](https://www.terraform.io/downloads.html)
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

### Using Mock Data (Default)
The provider works out of the box with mock data for testing:

**Windows:**
```powershell
.\dev.ps1 apply
```

**Linux/macOS:**
```bash
./dev.sh apply
```

### Using Real Credentials
Set environment variables with your actual Nosana credentials:

**Windows:**
```powershell
$env:TF_VAR_wallet_address = "your_wallet_address"
$env:TF_VAR_signed_challenge = "your_signed_challenge"
$env:TF_VAR_network = "mainnet"  # or "devnet"

.\dev.ps1 apply
```

**Linux/macOS:**
```bash
export TF_VAR_wallet_address="your_wallet_address"
export TF_VAR_signed_challenge="your_signed_challenge"
export TF_VAR_network="mainnet"  # or "devnet"

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
  wallet_address    = var.wallet_address
  signed_challenge  = var.signed_challenge
  network          = var.network
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
