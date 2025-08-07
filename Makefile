# Makefile for Terraform Provider Development

# Build the provider
build:
	go build -o terraform-provider-nosana.exe .

# Clean build artifacts
clean:
	rm -f terraform-provider-nosana.exe
	rm -rf .terraform/
	rm -f .terraform.lock.hcl
	rm -f terraform.tfstate*

# Install the provider locally for development
install: build
	mkdir -p %APPDATA%\terraform.d\plugins\localhost\hoomandigital\nosana\1.0.0\windows_amd64
	copy terraform-provider-nosana.exe %APPDATA%\terraform.d\plugins\localhost\hoomandigital\nosana\1.0.0\windows_amd64\

# Initialize Terraform with the local provider
init-local: install
	terraform init

# Plan with the local provider
plan-local: init-local
	terraform plan

# Apply with the local provider
apply-local: init-local
	terraform apply -auto-approve

# Destroy resources
destroy-local:
	terraform destroy -auto-approve

# Run tests
test:
	go test ./...

# Format Go code
fmt:
	go fmt ./...

# Run Go vet
vet:
	go vet ./...

# Development cycle: build, install, and test
dev: clean build install init-local

.PHONY: build clean install init-local plan-local apply-local destroy-local test fmt vet dev
