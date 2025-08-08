# Local Development and Testing Guide

## Quick Start

### 1. Build and Install the Provider

```powershell
# Option 1: Use the PowerShell script (recommended)
.\dev.ps1 dev

# Option 2: Manual commands
go build -o terraform-provider-nosana.exe .
mkdir "$env:APPDATA\terraform.d\plugins\localhost\hoomandigital\nosana\1.0.0\windows_amd64" -Force
copy terraform-provider-nosana.exe "$env:APPDATA\terraform.d\plugins\localhost\hoomandigital\nosana\1.0.0\windows_amd64\"
```

### 2. Test the Provider

```powershell
# Initialize Terraform
.\dev.ps1 init

# Plan changes
.\dev.ps1 plan

# Apply changes
.\dev.ps1 apply

# Destroy resources
.\dev.ps1 destroy
```

## Development Workflow

### Option A: Using the PowerShell Script

```powershell
# Show all available commands
.\dev.ps1 help

# Full development cycle (clean, build, install, init)
.\dev.ps1 dev

# Test your changes
.\dev.ps1 plan
.\dev.ps1 apply

# Clean up
.\dev.ps1 destroy
```

### Option B: Manual Commands

```powershell
# Build the provider
go build -o terraform-provider-nosana.exe .

# Clean up old installations and state
Remove-Item -Recurse -Force .terraform -ErrorAction SilentlyContinue
Remove-Item .terraform.lock.hcl -ErrorAction SilentlyContinue
Remove-Item terraform.tfstate* -ErrorAction SilentlyContinue

# Install provider locally
$pluginPath = "$env:APPDATA\terraform.d\plugins\localhost\hoomandigital\nosana\1.0.0\windows_amd64"
New-Item -ItemType Directory -Path $pluginPath -Force
Copy-Item terraform-provider-nosana.exe $pluginPath -Force

# Initialize and test
terraform init
terraform plan
terraform apply
```

## Environment Configuration

### For Testing with Mock Data (Default)
The `test-local.tf` file uses mock values by default, so you can test immediately:

```powershell
.\dev.ps1 apply
```

### For Real Nosana API Testing
Set environment variables with your real credentials:

```powershell
$env:TF_VAR_wallet_address = "your_actual_wallet_address"
$env:TF_VAR_signed_challenge = "your_actual_signed_challenge"
$env:TF_VAR_network = "mainnet"  # or "devnet"

.\dev.ps1 apply
```

## Debugging

### Enable Terraform Debug Logging
```powershell
$env:TF_LOG = "DEBUG"
$env:TF_LOG_PATH = "terraform.log"
terraform apply
```

### View Provider Logs
The provider outputs debug information to the console. Look for lines starting with `[INFO]`.

### Common Issues

1. **Provider not found**: Ensure the provider is installed in the correct plugin directory
2. **Build failures**: Check Go version (requires Go 1.21+) and dependencies
3. **Permission errors**: Ensure you have write access to `%APPDATA%\terraform.d\`

## Project Structure

```
terraform-provider-nosana/
├── nosana/                   # Provider source code package  
│   ├── provider.go           # Provider configuration and client
│   └── resource_nosana_job.go# Job resource implementation
├── examples/                 # Usage examples
│   ├── README.md             # Example documentation
│   └── main.tf               # Working configuration samples
├── SETUP.md                  # Detailed setup guide
├── DEV_GUIDE.md              # This development guide
├── main.go                   # Entry point
├── dev.ps1 / dev.sh          # Cross-platform development scripts
├── go.mod                    # Go module definition
└── .gitignore                # Git ignore rules
```

## Next Steps

1. **Extend the Provider**: Add more resources or data sources
2. **Real API Integration**: Replace mock API calls with real HTTP requests
3. **Testing**: Add unit tests and integration tests
4. **Documentation**: Generate provider documentation
5. **Publishing**: Package for Terraform Registry

## Useful Commands

```powershell
# Format Go code
go fmt ./...

# Run tests
go test ./...

# Vet code for issues
go vet ./...

# Update dependencies
go mod tidy

# View Terraform state
terraform show

# Import existing resources
terraform import nosana_job.example job-id-123
```
