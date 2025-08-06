# PowerShell script for Terraform Provider Development
# Usage: .\dev.ps1 [command]
# Commands: build, clean, install, init, plan, apply, destroy, test, dev

param(
    [Parameter(Position=0)]
    [string]$Command = "help"
)

$ProviderName = "terraform-provider-nosana.exe"
$PluginPath = "$env:APPDATA\terraform.d\plugins\localhost\hoomandigital\nosana\1.0.0\windows_amd64"

function Build {
    Write-Host "Building provider..." -ForegroundColor Green
    go build -o $ProviderName .
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Build successful!" -ForegroundColor Green
    } else {
        Write-Host "Build failed!" -ForegroundColor Red
        exit 1
    }
}

function Clean {
    Write-Host "Cleaning build artifacts..." -ForegroundColor Yellow
    if (Test-Path $ProviderName) { Remove-Item $ProviderName }
    if (Test-Path ".terraform") { Remove-Item ".terraform" -Recurse -Force }
    if (Test-Path ".terraform.lock.hcl") { Remove-Item ".terraform.lock.hcl" }
    if (Test-Path "terraform.tfstate") { Remove-Item "terraform.tfstate" }
    if (Test-Path "terraform.tfstate.backup") { Remove-Item "terraform.tfstate.backup" }
    Write-Host "Clean complete!" -ForegroundColor Green
}

function Install {
    Write-Host "Installing provider locally..." -ForegroundColor Green
    Build
    
    # Create plugin directory
    if (!(Test-Path $PluginPath)) {
        New-Item -ItemType Directory -Path $PluginPath -Force | Out-Null
    }
    
    # Copy provider binary
    Copy-Item $ProviderName $PluginPath -Force
    Write-Host "Provider installed at: $PluginPath" -ForegroundColor Green
}

function InitLocal {
    Write-Host "Initializing Terraform..." -ForegroundColor Green
    Install
    terraform init
}

function PlanLocal {
    Write-Host "Running Terraform plan..." -ForegroundColor Green
    InitLocal
    terraform plan
}

function ApplyLocal {
    Write-Host "Running Terraform apply..." -ForegroundColor Green
    InitLocal
    terraform apply -auto-approve
}

function DestroyLocal {
    Write-Host "Running Terraform destroy..." -ForegroundColor Red
    terraform destroy -auto-approve
}

function RunTests {
    Write-Host "Running Go tests..." -ForegroundColor Green
    go test ./...
}

function FormatCode {
    Write-Host "Formatting Go code..." -ForegroundColor Green
    go fmt ./...
}

function VetCode {
    Write-Host "Running Go vet..." -ForegroundColor Green
    go vet ./...
}

function DevCycle {
    Write-Host "Running development cycle..." -ForegroundColor Cyan
    Clean
    Build
    Install
    InitLocal
    Write-Host "Development cycle complete! Ready to plan/apply." -ForegroundColor Green
}

function ShowHelp {
    Write-Host "Terraform Provider Development Script" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Usage: .\dev.ps1 [command]" -ForegroundColor White
    Write-Host ""
    Write-Host "Commands:" -ForegroundColor Yellow
    Write-Host "  build     - Build the provider binary" -ForegroundColor White
    Write-Host "  clean     - Remove build artifacts and Terraform files" -ForegroundColor White
    Write-Host "  install   - Build and install provider locally" -ForegroundColor White
    Write-Host "  init      - Initialize Terraform with local provider" -ForegroundColor White
    Write-Host "  plan      - Run terraform plan" -ForegroundColor White
    Write-Host "  apply     - Run terraform apply" -ForegroundColor White
    Write-Host "  destroy   - Run terraform destroy" -ForegroundColor White
    Write-Host "  test      - Run Go tests" -ForegroundColor White
    Write-Host "  fmt       - Format Go code" -ForegroundColor White
    Write-Host "  vet       - Run Go vet" -ForegroundColor White
    Write-Host "  dev       - Full development cycle (clean, build, install, init)" -ForegroundColor White
    Write-Host ""
    Write-Host "Examples:" -ForegroundColor Yellow
    Write-Host "  .\dev.ps1 dev" -ForegroundColor Gray
    Write-Host "  .\dev.ps1 plan" -ForegroundColor Gray
    Write-Host "  .\dev.ps1 apply" -ForegroundColor Gray
}

# Main command dispatcher
switch ($Command.ToLower()) {
    "build" { Build }
    "clean" { Clean }
    "install" { Install }
    "init" { InitLocal }
    "plan" { PlanLocal }
    "apply" { ApplyLocal }
    "destroy" { DestroyLocal }
    "test" { RunTests }
    "fmt" { FormatCode }
    "vet" { VetCode }
    "dev" { DevCycle }
    "help" { ShowHelp }
    default { 
        Write-Host "Unknown command: $Command" -ForegroundColor Red
        ShowHelp
    }
}
