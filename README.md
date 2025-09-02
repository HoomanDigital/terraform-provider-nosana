# Terraform Provider for Nosana

A simple Terraform provider for managing jobs on the Nosana Network.

## Prerequisites

Before you begin, you will need:

1.  **Terraform 1.0+**: [Install Terraform](https://learn.hashicorp.com/tutorials/terraform/install-cli)
2.  **Nosana CLI**: The Nosana Command Line Interface is required to generate a keypair.
    ```bash
    npm install -g @nosana/cli
    ```
3.  **Nosana Keypair**: You need a funded Nosana wallet to pay for jobs.
    ```bash
    # Generate a new keypair (and save the seed phrase)
    nosana address

    # Fund your new wallet with SOL and NOS tokens
    ```

## 🚀 Quick Start

This repository contains a simple, working example to get you started.

1.  **Clone the Repository**:
    ```bash
    git clone https://github.com/HoomanDigital/terraform-provider-nosana.git
    cd terraform-provider-nosana
    ```

2.  **Run the Example**:
    The `example/` directory contains a pre-configured job.
    ```bash
    cd example
    terraform init
    terraform apply
    ```

That's it! This will submit a job to the Nosana network using the keypair found at the default location (`~/.nosana/nosana_key.json`).
