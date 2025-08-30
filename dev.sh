#!/bin/bash
# Linux/macOS script for Terraform Provider Development
# Usage: ./dev.sh [command]
# Commands: build, clean, install, init, plan, apply, destroy, test, dev, help

set -e

PROVIDER_NAME="terraform-provider-nosana"
PLUGIN_PATH="$HOME/.terraform.d/plugins/localhost/codebrewery/nosana/1.0.0/linux_amd64"

function build() {
    echo -e "\033[32mBuilding provider...\033[0m"
    go build -o "$PROVIDER_NAME" .
    echo -e "\033[32mBuild successful!\033[0m"
}

function clean() {
    echo -e "\033[33mCleaning build artifacts...\033[0m"
    rm -f "$PROVIDER_NAME"
    rm -rf .terraform/
    rm -f .terraform.lock.hcl
    rm -f terraform.tfstate*
    echo -e "\033[32mClean complete!\033[0m"
}

function install() {
    echo -e "\033[32mInstalling provider locally...\033[0m"
    build
    
    # Create plugin directory
    mkdir -p "$PLUGIN_PATH"
    
    # Copy provider binary
    cp "$PROVIDER_NAME" "$PLUGIN_PATH/"
    chmod +x "$PLUGIN_PATH/$PROVIDER_NAME"
    echo -e "\033[32mProvider installed at: $PLUGIN_PATH\033[0m"
}

function init_local() {
    echo -e "\033[32mInitializing Terraform...\033[0m"
    install
    terraform init
}

function plan_local() {
    echo -e "\033[32mRunning Terraform plan...\033[0m"
    init_local
    terraform plan
}

function apply_local() {
    echo -e "\033[32mRunning Terraform apply...\033[0m"
    init_local
    terraform apply -auto-approve
}

function destroy_local() {
    echo -e "\033[31mRunning Terraform destroy...\033[0m"
    terraform destroy -auto-approve
}

function run_tests() {
    echo -e "\033[32mRunning Go tests...\033[0m"
    go test ./...
}

function format_code() {
    echo -e "\033[32mFormatting Go code...\033[0m"
    go fmt ./...
}

function vet_code() {
    echo -e "\033[32mRunning Go vet...\033[0m"
    go vet ./...
}

function dev_cycle() {
    echo -e "\033[36mRunning development cycle...\033[0m"
    clean
    build
    install
    init_local
    echo -e "\033[32mDevelopment cycle complete! Ready to plan/apply.\033[0m"
}

function show_help() {
    echo -e "\033[36mTerraform Provider Development Script\033[0m"
    echo ""
    echo -e "\033[37mUsage: ./dev.sh [command]\033[0m"
    echo ""
    echo -e "\033[33mCommands:\033[0m"
    echo -e "\033[37m  build     - Build the provider binary\033[0m"
    echo -e "\033[37m  clean     - Remove build artifacts and Terraform files\033[0m"
    echo -e "\033[37m  install   - Build and install provider locally\033[0m"
    echo -e "\033[37m  init      - Initialize Terraform with local provider\033[0m"
    echo -e "\033[37m  plan      - Run terraform plan\033[0m"
    echo -e "\033[37m  apply     - Run terraform apply\033[0m"
    echo -e "\033[37m  destroy   - Run terraform destroy\033[0m"
    echo -e "\033[37m  test      - Run Go tests\033[0m"
    echo -e "\033[37m  fmt       - Format Go code\033[0m"
    echo -e "\033[37m  vet       - Run Go vet\033[0m"
    echo -e "\033[37m  dev       - Full development cycle (clean, build, install, init)\033[0m"
    echo ""
    echo -e "\033[33mExamples:\033[0m"
    echo -e "\033[37m  ./dev.sh dev\033[0m"
    echo -e "\033[37m  ./dev.sh plan\033[0m"
    echo -e "\033[37m  ./dev.sh apply\033[0m"
}

# Main command dispatcher
case "${1:-help}" in
    "build") build ;;
    "clean") clean ;;
    "install") install ;;
    "init") init_local ;;
    "plan") plan_local ;;
    "apply") apply_local ;;
    "destroy") destroy_local ;;
    "test") run_tests ;;
    "fmt") format_code ;;
    "vet") vet_code ;;
    "dev") dev_cycle ;;
    "help") show_help ;;
    *) 
        echo -e "\033[31mUnknown command: $1\033[0m"
        show_help
        ;;
esac
