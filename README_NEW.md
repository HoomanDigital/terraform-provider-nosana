# Terraform Provider for Nosana

A Terraform provider for managing deployments on the Nosana network - a decentralized compute platform.

## Features

- 🚀 **Deploy containerized applications** to the Nosana decentralized network
- 🔐 **Automatic wallet management** with secure local storage
- ⚙️ **Full lifecycle management** - create, read, update, and delete deployments
- 🔄 **Configurable deployment strategies** (Simple, Infinite, Scheduled)
- ⏱️ **Optional completion waiting** with configurable timeouts
- 📊 **Resource scaling** support (replicas and timeout updates)

## Quick Start

### Prerequisites

1. **Go 1.21+** installed
2. **Terraform 1.0+** installed  
3. **SOL and NOS tokens** in your wallet for deployment costs

### Installation & Testing

1. **Clone and build the provider:**
```bash
git clone <repository-url>
cd terraform-provider-nosana
./build-and-test.sh
```

2. **Deploy a sample application:**
```bash
./deploy.sh
```

3. **Clean up resources:**
```bash
./destroy.sh
```

## Configuration

### Provider Configuration

```hcl
provider "nosana" {
  # Use automatically managed local wallet (recommended)
  use_local_wallet = true
  market_address   = "HanragNudL4S4zFtpLQv85dn6QbdzCm7SNEWEb9sRp17"
  
  # Alternative: use explicit private key
  # private_key = "your_base58_encoded_private_key"
  
  # Alternative: use existing keypair file
  # keypair_path = "/path/to/your/keypair.json"
}
```

### Authentication Methods

The provider supports three authentication methods (in order of precedence):

1. **Explicit Private Key** - Set `private_key` parameter or `NOSANA_PRIVATE_KEY` environment variable
2. **Keypair File** - Set `keypair_path` parameter or `NOSANA_KEYPAIR_PATH` environment variable  
3. **Local Wallet** - Automatically generated and stored at `~/.config/terraform-provider-nosana/wallet.json`

### Resource Configuration

```hcl
resource "nosana_job" "my_app" {
  job_content = jsonencode({
    "ops": [
      {
        "id": "my-container",
        "args": {
          "env": {
            "PORT": "8000"
          },
          "image": "nginx:alpine",
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
  timeout  = 600
  strategy = "SIMPLE"

  # Optional: wait for deployment to complete
  wait_for_completion        = true
  completion_timeout_seconds = 300
}
```

## Resource Schema

### `nosana_job`

#### Arguments

- `job_content` (String, Required) - JSON-encoded job definition content
- `replicas` (Number, Required) - Number of replicas for the deployment
- `timeout` (Number, Optional, Default: 600) - Timeout in seconds for the deployment
- `strategy` (String, Required) - Deployment strategy: `SIMPLE`, `SIMPLE-EXTEND`, `SCHEDULED`, or `INFINITE`
- `schedule` (String, Optional) - Cron expression for `SCHEDULED` strategy
- `wait_for_completion` (Boolean, Optional, Default: false) - Wait for deployment to reach stable state
- `completion_timeout_seconds` (Number, Optional, Default: 300) - Maximum wait time for completion

#### Attributes

- `id` (String) - Terraform resource identifier
- `job_id` (String) - Nosana deployment ID
- `status` (String) - Current deployment status

## Deployment Strategies

- **SIMPLE** - Standard one-time deployment
- **SIMPLE-EXTEND** - Extended simple deployment with additional features
- **SCHEDULED** - Cron-based scheduled deployment (requires `schedule` parameter)
- **INFINITE** - Long-running deployment

## Deployment Statuses

- `DRAFT` - Deployment created but not started
- `STARTING` - Deployment is initializing
- `RUNNING` - Deployment is active and running
- `STOPPING` - Deployment is being stopped
- `STOPPED` - Deployment has been stopped
- `ARCHIVED` - Deployment has been archived/deleted
- `ERROR` - Deployment encountered an error
- `INSUFFICIENT_FUNDS` - Not enough tokens to run deployment

## Development

### Project Structure

```
terraform-provider-nosana/
├── main.go                 # Provider entry point
├── nosana/
│   ├── provider.go         # Provider configuration and authentication
│   ├── client.go          # Nosana API client implementation
│   ├── resource_nosana_job.go # Job resource CRUD operations
│   └── *_test.go          # Test files
├── example/
│   └── main.tf            # Example configuration
├── build-and-test.sh      # Build and test script
├── deploy.sh              # Deployment script
└── destroy.sh             # Cleanup script
```

### Running Tests

```bash
# Unit tests
go test ./nosana -v

# Acceptance tests (requires real credentials)
TF_ACC=1 NOSANA_MARKET_ADDRESS="your_market_address" go test ./nosana -v
```

### Building

```bash
# Build provider binary
go build -o terraform-provider-nosana .

# Install to development location
cp terraform-provider-nosana /home/dhruv/go/bin/
```

## Wallet Management

### Local Wallet

When `use_local_wallet = true`, the provider automatically:

1. Creates a new Solana keypair if none exists
2. Stores it securely at `~/.config/terraform-provider-nosana/wallet.json`
3. Uses restricted file permissions (0600)
4. Displays the public key for funding

### Funding Your Wallet

To use the provider, your wallet needs:

1. **SOL tokens** - For Solana transaction fees
2. **NOS tokens** - For Nosana network deployment costs

You can:
- Transfer tokens to your wallet's public key
- Use a Solana faucet for testnet SOL
- Purchase tokens on supported exchanges

## Troubleshooting

### Common Issues

1. **"Unauthorized" errors** - Check wallet funding and private key format
2. **"Internal Server Error"** - Verify API connectivity and request format
3. **Build failures** - Ensure Go 1.21+ and run `go mod tidy`
4. **Provider not found** - Check development overrides in `~/.terraformrc`

### Debug Mode

Enable debug logging:
```bash
export TF_LOG=DEBUG
terraform plan
```

### API Connectivity Test

```bash
# Test API connectivity (requires NOSANA_PRIVATE_KEY)
./test-api.sh
```

## Security Considerations

- Private keys are stored securely with restricted permissions
- Never commit private keys to version control  
- Use environment variables for sensitive configuration
- Regularly audit wallet access and funding

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run the test suite
6. Submit a pull request

## License

[Add your license information here]

## Support

For issues and questions:
- Create an issue in this repository
- Check the Nosana documentation
- Join the Nosana community Discord
