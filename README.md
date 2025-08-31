# Terraform Provider for Nosana

A Terraform provider for managing Nosana jobs on the Nosana Network. Deploy AI/ML workloads, web services, and containerized applications on decentralized compute infrastructure.

[![Registry](https://img.shields.io/badge/registry-terraform.io-blue)](https://registry.terraform.io/providers/HoomanDigital/nosana)
[![Release](https://img.shields.io/github/v/release/HoomanDigital/terraform-provider-nosana)](https://github.com/HoomanDigital/terraform-provider-nosana/releases)
[![License](https://img.shields.io/github/license/HoomanDigital/terraform-provider-nosana)](LICENSE)

## 🚀 Quick Start

```hcl
terraform {
  required_providers {
    nosana = {
      source  = "registry.terraform.io/HoomanDigital/nosana"
      version = "~> 0.1"
    }
  }
}

provider "nosana" {
  keypair_path   = "~/.config/solana/id.json"
  network        = "mainnet-beta"
  market_address = "nosScmHY2uR24Zh751PmGj9ww9QRNHewh9H59AfrT"
}

resource "nosana_job" "ai_workload" {
  job_definition = jsonencode({
    "ops": [{
      "id": "ai-training",
      "args": {
        "image": "pytorch/pytorch:latest",
        "gpu": true,
        "env": {
          "MODEL_NAME": "bert-base-uncased"
        }
      },
      "type": "container/run"
    }],
    "type": "container",
    "version": "0.1"
  })
  
  wait_for_completion = false
  completion_timeout_seconds = 600
}
```

## 📋 What It Does

Deploy **AI/ML workloads**, **web services**, and **containerized applications** on the Nosana decentralized compute network using familiar Terraform workflows.

**Key Features:**
- 🤖 **GPU-enabled AI workloads** (LLMs, ML inference, training)
- 🌐 **Web services** with automatic port exposure  
- 💰 **Cost-effective** - Pay only for compute time used
- 🔒 **Decentralized** - No single point of failure
- 📊 **Terraform state management** - Full lifecycle support

## 📚 Examples

Check out the [`examples/`](./examples/) directory for comprehensive usage examples:

- **`main.tf`** - Complete AI/ML workload example with GPU support
- **`sample.tf`** - Web services, data processing, and multiple use cases
- **`test.tf`** - Simple test configuration to verify your setup
- **`README.md`** - Detailed documentation for all examples

### Quick Start:
```bash
cd examples
terraform init
terraform plan
terraform apply
```

## 🛠️ Prerequisites

- **Terraform 1.0+**
- **Nosana CLI**: `npm install -g @nosana/cli`
- **Funded Nosana wallet** with SOL and NOS tokens
- **Valid Solana keypair** (usually at `~/.config/solana/id.json`)

## 🔧 Development

### Development Commands

| Command | Description |
|---------|-------------|
| `./dev.sh build` | Build the provider binary |
| `./dev.sh install` | Build and install provider locally |
| `./dev.sh test` | Run Go tests |
| `./dev.sh fmt` | Format Go code |
| `./dev.sh dev` | Full development cycle |

### Automated Releases

This repository uses GitHub Actions for automated releases:

1. **Create a new tag:** `git tag v0.2.0`
2. **Push the tag:** `git push origin v0.2.0`
3. **GitHub Actions automatically:**
   - Builds binaries for all platforms
   - Creates ZIP archives
   - Generates SHA256SUMS and signatures
   - Creates GitHub release
   - Uploads all files

The Terraform Registry will automatically discover and publish new releases!

### Setup Release Automation

Run the setup script to configure GitHub secrets:
```bash
./scripts/setup-secrets.sh
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
├── scripts/             # Automation scripts
│   └── setup-secrets.sh # GitHub secrets setup
├── .github/workflows/   # CI/CD automation
│   └── release.yml      # Automated release workflow
├── SETUP.md             # Detailed setup guide
├── DEV_GUIDE.md         # Development guide
├── dev.sh               # Development script
└── main.go              # Entry point
```

## 📖 Documentation

- **[Provider Documentation](https://registry.terraform.io/providers/HoomanDigital/nosana/latest/docs)** - Official Terraform Registry docs
- **[Setup Guide](SETUP.md)** - Detailed setup instructions
- **[Development Guide](DEV_GUIDE.md)** - Contributing and development
- **[Examples](examples/)** - Real-world usage examples

## 🌍 Registry

This provider is published on the [Terraform Registry](https://registry.terraform.io/providers/HoomanDigital/nosana) and available for public use.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Run `./dev.sh test`
6. Submit a pull request

## 📞 Support

- **GitHub Issues**: [Report bugs or request features](https://github.com/HoomanDigital/terraform-provider-nosana/issues)
- **Documentation**: [Terraform Registry](https://registry.terraform.io/providers/HoomanDigital/nosana)
- **Examples**: [Example configurations](examples/)

---

**Built with ❤️ for the decentralized compute community**