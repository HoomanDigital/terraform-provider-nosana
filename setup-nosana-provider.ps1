# setup-nosana-provider.ps1 - Quick setup script for Nosana Terraform Provider

Write-Host "🚀 Nosana Terraform Provider Setup" -ForegroundColor Cyan
Write-Host "====================================" -ForegroundColor Cyan

# Check if required tools are installed
Write-Host "📋 Checking prerequisites..." -ForegroundColor Yellow

# Check Node.js/npm
try {
    $npmVersion = npm --version
    Write-Host "✅ npm found (version: $npmVersion)" -ForegroundColor Green
} catch {
    Write-Host "❌ npm not found. Please install Node.js first:" -ForegroundColor Red
    Write-Host "   - Visit: https://nodejs.org/" -ForegroundColor White
    exit 1
}

# Check Terraform
try {
    $tfVersion = terraform version
    Write-Host "✅ terraform found" -ForegroundColor Green
} catch {
    Write-Host "❌ terraform not found. Please install Terraform first:" -ForegroundColor Red
    Write-Host "   - Visit: https://terraform.io/downloads" -ForegroundColor White
    exit 1
}

# Install Nosana CLI
Write-Host ""
Write-Host "📦 Installing Nosana CLI..." -ForegroundColor Yellow
try {
    npm install -g @nosana/cli
    Write-Host "✅ Nosana CLI installed successfully" -ForegroundColor Green
} catch {
    Write-Host "❌ Failed to install Nosana CLI" -ForegroundColor Red
    exit 1
}

# Check Nosana CLI version
Write-Host ""
Write-Host "🔍 Nosana CLI version:" -ForegroundColor Yellow
nosana --version

Write-Host ""
Write-Host "🎉 Setup complete!" -ForegroundColor Green
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Cyan
Write-Host "1. Get your Solana private key:" -ForegroundColor White
Write-Host "   - From Phantom: Settings → Security & Privacy → Export Private Key" -ForegroundColor Gray
Write-Host "   - From Solflare: Settings → Export Private Key" -ForegroundColor Gray
Write-Host "   - From CLI: solana-keygen recover 'prompt:' (if you have the seed phrase)" -ForegroundColor Gray
Write-Host ""
Write-Host "2. Set environment variables:" -ForegroundColor White
Write-Host '   $env:NOSANA_PRIVATE_KEY = "your_base58_private_key_here"' -ForegroundColor Yellow
Write-Host '   $env:TF_VAR_market_address = "7AtiXMSH6R1jjBxrcYjehCkkSF7zvYWte63gwEDBcGHq"' -ForegroundColor Yellow
Write-Host ""
Write-Host "3. Ensure your wallet has SOL and NOS tokens:" -ForegroundColor White
Write-Host "   - SOL: For transaction fees (at least 0.01 SOL)" -ForegroundColor Gray
Write-Host "   - NOS: For paying job execution costs" -ForegroundColor Gray
Write-Host "   - You can buy NOS on exchanges like Kraken" -ForegroundColor Gray
Write-Host ""
Write-Host "4. Initialize and run Terraform:" -ForegroundColor White
Write-Host "   terraform init" -ForegroundColor Yellow
Write-Host "   terraform plan" -ForegroundColor Yellow
Write-Host "   terraform apply" -ForegroundColor Yellow
Write-Host ""
Write-Host "📚 For more information, visit: https://docs.nosana.com/" -ForegroundColor Cyan
