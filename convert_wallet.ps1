# PowerShell script to convert Phantom wallet private key to Nosana format
param(
    [Parameter(Mandatory=$true)]
    [string]$PrivateKey
)

# Install required NuGet package if not already installed
if (-not (Get-Package -Name "System.Buffers" -ErrorAction SilentlyContinue)) {
    Write-Host "Installing required package..." -ForegroundColor Yellow
    Install-Package -Name "System.Buffers" -Force -SkipDependencies
}

# Base58 decode function (simplified for Solana keys)
function ConvertFrom-Base58 {
    param([string]$Base58String)
    
    $alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
    $result = @()
    
    # Simple base58 decode implementation
    $bigInt = [System.Numerics.BigInteger]::Zero
    $base = [System.Numerics.BigInteger]::One
    
    for ($i = $Base58String.Length - 1; $i -ge 0; $i--) {
        $char = $Base58String[$i]
        $value = $alphabet.IndexOf($char)
        if ($value -eq -1) {
            throw "Invalid base58 character: $char"
        }
        $bigInt += $base * $value
        $base *= 58
    }
    
    # Convert to byte array
    $bytes = $bigInt.ToByteArray()
    
    # Handle leading zeros
    $leadingZeros = 0
    for ($i = 0; $i -lt $Base58String.Length; $i++) {
        if ($Base58String[$i] -eq '1') {
            $leadingZeros++
        } else {
            break
        }
    }
    
    # Reverse bytes (BigInteger is little-endian, we want big-endian)
    [Array]::Reverse($bytes)
    
    # Add leading zeros
    $result = @(0) * $leadingZeros + $bytes
    
    # Remove any trailing zero if present (BigInteger adds it)
    if ($result[-1] -eq 0 -and $result.Length -gt 32) {
        $result = $result[0..($result.Length-2)]
    }
    
    return $result
}

try {
    Write-Host "Converting private key from base58 to byte array..." -ForegroundColor Yellow
    
    # Convert the private key
    $keyBytes = ConvertFrom-Base58 $PrivateKey
    
    Write-Host "Private key length: $($keyBytes.Length) bytes" -ForegroundColor Green
    
    if ($keyBytes.Length -ne 64) {
        Write-Host "Warning: Expected 64 bytes for Solana private key, got $($keyBytes.Length)" -ForegroundColor Yellow
    }
    
    # Create the JSON array format that Nosana expects
    $jsonArray = $keyBytes -join ","
    
    # Backup existing keypair
    $nosanaDir = "$env:USERPROFILE\.nosana"
    $keypairPath = "$nosanaDir\nosana_key.json"
    $backupPath = "$nosanaDir\nosana_key.json.backup"
    
    if (Test-Path $keypairPath) {
        Copy-Item $keypairPath $backupPath
        Write-Host "Backed up existing keypair to: $backupPath" -ForegroundColor Green
    }
    
    # Write new keypair
    "[$jsonArray]" | Out-File -FilePath $keypairPath -Encoding UTF8
    
    Write-Host "Keypair updated successfully!" -ForegroundColor Green
    Write-Host "New keypair saved to: $keypairPath" -ForegroundColor Green
    
} catch {
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "Failed to convert private key. Please check the format." -ForegroundColor Red
}
